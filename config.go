package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Config is the fully resolved runtime configuration. Everything downstream
// reads from here rather than reaching into the environment itself, so the
// whole contract with the operator is visible in one place.
type Config struct {
	SABHost    string
	SABPort    int
	SABAPIKey  string
	WebTimeout time.Duration

	StatsFile       string
	FaviconFile     string
	MaintenanceFile string
	UptimeImageURL  string
	PlexURL         string

	CloudSource  string
	CloudDest    string
	CloudCache   string
	CloudTimeout time.Duration
	CloudMaxAge  time.Duration

	LogLevel LogLevel
}

// Every duration below is configured in whole seconds, matching the WEBTIMEOUT
// and SABPORT convention this program already exposed to its CronJob.
const (
	defaultWebTimeoutSeconds = 15
	// Dropbox answers a throttle with "trying again in 300 seconds", so a budget
	// of 300 could never survive one: the backoff alone consumed it and the run
	// timed out having done nothing. 600 covers one backoff plus the roughly
	// 100-second remote measurement that follows, and still fits inside the Job's
	// activeDeadlineSeconds.
	defaultCloudTimeoutSeconds = 600
	// Sizing the remote walks it in full, and Dropbox throttles hard once it has
	// seen a few of those in quick succession - an unthrottled pass takes about
	// 100 seconds, a throttled one can exceed ten minutes. Refreshing
	// twice a day keeps the figure current enough for an upload measured in weeks
	// while leaving the sync job's own API budget alone.
	defaultCloudMaxAgeSeconds = 12 * 60 * 60
)

// LoadConfig reads configuration from the environment, applying defaults for
// anything unset. A malformed value is an error rather than a silent fallback:
// a mistyped timeout should be fixed, not quietly ignored.
func LoadConfig() (*Config, error) {
	config := &Config{
		SABHost:        envString("SABHOST", "https://sab.zaldre.com"),
		SABAPIKey:      envString("SABAPIKEY", "YOURKEY"),
		StatsFile:      envString("STATSFILE", "/container/data/stats/index.html"),
		UptimeImageURL: envString("UPTIME", "https://app.statuscake.com/button/index.php?Track=6422414&Days=30&Design=2"),
		PlexURL:        envString("PLEXURL", "https://app.plex.tv"),
		// The whole of pub is what the sync job uploads, so it is what the page
		// measures: anything narrower reports progress against a subset of the
		// work and reaches 100% while uploads are still running.
		CloudSource: envString("CLOUDSRC", "/mnt/core/pub"),
		CloudDest:   envString("CLOUDDST", "pub:"),
		LogLevel:    ParseLogLevel(envString("LOGLEVEL", "Normal")),
	}

	outputDir := filepath.Dir(config.StatsFile)

	// These three all default alongside the generated page rather than alongside
	// the binary. The maintenance notice used to resolve against os.Args[0],
	// which put it at /usr/bin/maintenance.txt inside the container image: a
	// read-only path that never held the file, so the notice never appeared.
	config.MaintenanceFile = envString("MAINTENANCEFILE", filepath.Join(outputDir, "maintenance.txt"))
	config.FaviconFile = envString("FAVICONFILE", filepath.Join(outputDir, "favicon.ico"))
	config.CloudCache = envString("CLOUDCACHE", filepath.Join(outputDir, "cloud-progress.json"))

	var err error
	if config.SABPort, err = envInt("SABPORT", 443); err != nil {
		return nil, err
	}
	if config.WebTimeout, err = envSeconds("WEBTIMEOUT", defaultWebTimeoutSeconds); err != nil {
		return nil, err
	}
	if config.CloudTimeout, err = envSeconds("CLOUDTIMEOUT", defaultCloudTimeoutSeconds); err != nil {
		return nil, err
	}
	if config.CloudMaxAge, err = envSeconds("CLOUDMAXAGE", defaultCloudMaxAgeSeconds); err != nil {
		return nil, err
	}

	return config, nil
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envInt(key string, fallback int) (int, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be a whole number, got %q: %w", key, raw, err)
	}
	return value, nil
}

// envSeconds reads a duration expressed in whole seconds. Zero and negative
// values are rejected rather than normalised: every caller uses the result as a
// timeout or a cache lifetime, where a non-positive value silently disables the
// behaviour the operator was trying to tune.
func envSeconds(key string, fallbackSeconds int) (time.Duration, error) {
	seconds, err := envInt(key, fallbackSeconds)
	if err != nil {
		return 0, err
	}
	if seconds <= 0 {
		return 0, fmt.Errorf("%s must be a positive number of seconds, got %d", key, seconds)
	}
	return time.Duration(seconds) * time.Second, nil
}
