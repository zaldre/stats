package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LogLevel
	}{
		{"none", "None", LogNone},
		{"lowercase none", "none", LogNone},
		{"debug", "Debug", LogDebug},
		{"lowercase debug matches too", "debug", LogDebug},
		{"padded", "  DEBUG  ", LogDebug},
		{"normal", "Normal", LogNormal},
		{"unknown falls back to normal", "chatty", LogNormal},
		{"empty falls back to normal", "", LogNormal},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, ParseLogLevel(test.input))
		})
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	config, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "/mnt/core/pub", config.CloudSource)
	assert.Equal(t, "pub:", config.CloudDest)
	assert.Equal(t, 12*time.Hour, config.CloudMaxAge)
	assert.Equal(t, 600*time.Second, config.CloudTimeout)

	// The three sidecar files default beside the page, not beside the binary.
	assert.Equal(t, "/container/data/stats/maintenance.txt", config.MaintenanceFile)
	assert.Equal(t, "/container/data/stats/favicon.ico", config.FaviconFile)
	assert.Equal(t, "/container/data/stats/cloud-progress.json", config.CloudCache)
}

func TestLoadConfigOverrides(t *testing.T) {
	t.Setenv("STATSFILE", "/srv/www/index.html")
	t.Setenv("CLOUDMAXAGE", "60")

	config, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, time.Minute, config.CloudMaxAge)
	// Sidecars follow STATSFILE when it moves.
	assert.Equal(t, "/srv/www/favicon.ico", config.FaviconFile)
	assert.Equal(t, "/srv/www/maintenance.txt", config.MaintenanceFile)
}

func TestLoadConfigRejectsBadValues(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr string
	}{
		{"non-numeric port", "SABPORT", "eighty", "SABPORT must be a whole number"},
		{"non-numeric timeout", "CLOUDTIMEOUT", "5m", "CLOUDTIMEOUT must be a whole number"},
		{"zero timeout", "CLOUDTIMEOUT", "0", "must be a positive number of seconds"},
		{"negative max age", "CLOUDMAXAGE", "-1", "must be a positive number of seconds"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv(test.key, test.value)

			_, err := LoadConfig()
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.wantErr)
		})
	}
}
