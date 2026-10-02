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

	CloudSource   string
	CloudDest     string
	CloudCache    string
	CloudTimeout  time.Duration
	CloudMaxAge   time.Duration
	CloudTPSLimit int

	LogLevel LogLevel
}

// Every duration below is configured in whole seconds, matching the WEBTIMEOUT
// and SABPORT convention this program already exposed to its CronJob.
const (
	defaultWebTimeoutSeconds = 15
	// The remote walk costs one list call per directory, about 6,300 of them as
	// of October 2026, so at defaultCloudTPSLimit it needs roughly 26 minutes.
	// Dropbox answers a throttle with "trying again in 300 seconds"; 45 minutes
	// leaves room for three of those on top of the walk. The previous 600 assumed
	// a 100-second walk that the tree has long outgrown: every run timed out
	// mid-walk, so the cache was never written and the page never had a figure.
	// The CronJob's activeDeadlineSeconds must stay above this.
	defaultCloudTimeoutSeconds = 45 * 60
	// Sizing the remote walks it in full, so refreshing twice a day keeps the
	// figure current enough for an upload measured in weeks while leaving the
	// sync job's own API budget alone.
	defaultCloudMaxAgeSeconds = 12 * 60 * 60
	// The sync job runs at --tpslimit 8 and went a full day unthrottled, except
	// for the one penalty it took twenty seconds after the unlimited walk drew its
	// own. Four more keeps the pair at twelve calls a second.
	defaultCloudTPSLimit = 4
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
	if config.CloudTPSLimit, err = envPositiveInt("CLOUDTPSLIMIT", defaultCloudTPSLimit); err != nil {
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

// envPositiveInt rejects zero and negative values rather than normalising them:
// every caller uses the result as a timeout, a cache lifetime or a rate limit,
// where a non-positive value silently disables the behaviour the operator was
// trying to tune.
func envPositiveInt(key string, fallback int) (int, error) {
	value, err := envInt(key, fallback)
	if err != nil {
		return 0, err
	}
	if value <= 0 {
		return 0, fmt.Errorf("%s must be positive, got %d", key, value)
	}
	return value, nil
}

// envSeconds reads a duration expressed in whole seconds.
func envSeconds(key string, fallbackSeconds int) (time.Duration, error) {
	seconds, err := envPositiveInt(key, fallbackSeconds)
	if err != nil {
		return 0, err
	}
	return time.Duration(seconds) * time.Second, nil
}
