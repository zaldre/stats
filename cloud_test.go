package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseListing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]int64
		wantErr  string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: map[string]int64{},
		},
		{
			name:  "sizes and paths",
			input: "1024|movies/a.mkv\n2048|tv/show/s01e01.mkv\n",
			expected: map[string]int64{
				"movies/a.mkv":       1024,
				"tv/show/s01e01.mkv": 2048,
			},
		},
		{
			name:     "path containing the separator splits on the first only",
			input:    "512|music/AC|DC/song.flac\n",
			expected: map[string]int64{"music/AC|DC/song.flac": 512},
		},
		{
			name:     "blank lines are skipped",
			input:    "\n100|a\n\n",
			expected: map[string]int64{"a": 100},
		},
		{
			name:    "missing separator is an error",
			input:   "1024 movies/a.mkv\n",
			wantErr: "no separator",
		},
		{
			name:    "non-numeric size is an error",
			input:   "big|movies/a.mkv\n",
			wantErr: "malformed listing line",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			files, err := ParseListing(strings.NewReader(test.input))
			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, files)
		})
	}
}

func TestBucketFiles(t *testing.T) {
	tests := []struct {
		name              string
		local             map[string]int64
		remote            map[string]int64
		wantUploadedBytes int64
		wantUploadedFiles int
		wantRemainBytes   int64
		wantRemainFiles   int
		wantTotalBytes    int64
		wantPercent       float64
	}{
		{
			name:  "nothing uploaded yet",
			local: map[string]int64{"a": 100, "b": 300},
			// A remote holding none of it.
			remote:          map[string]int64{},
			wantRemainBytes: 400,
			wantRemainFiles: 2,
			wantTotalBytes:  400,
			wantPercent:     0,
		},
		{
			name:              "fully uploaded",
			local:             map[string]int64{"a": 100, "b": 300},
			remote:            map[string]int64{"a": 100, "b": 300},
			wantUploadedBytes: 400,
			wantUploadedFiles: 2,
			wantTotalBytes:    400,
			wantPercent:       100,
		},
		{
			name:              "size mismatch still counts as remaining",
			local:             map[string]int64{"a": 100, "b": 300},
			remote:            map[string]int64{"a": 100, "b": 299},
			wantUploadedBytes: 100,
			wantUploadedFiles: 1,
			wantRemainBytes:   300,
			wantRemainFiles:   1,
			wantTotalBytes:    400,
			wantPercent:       25,
		},
		{
			name:              "remote-only files are ignored entirely",
			local:             map[string]int64{"a": 100},
			remote:            map[string]int64{"a": 100, "deleted-locally": 9999},
			wantUploadedBytes: 100,
			wantUploadedFiles: 1,
			wantTotalBytes:    100,
			wantPercent:       100,
		},
		{
			name:   "empty source does not divide by zero",
			local:  map[string]int64{},
			remote: map[string]int64{"a": 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			progress := BucketFiles(test.local, test.remote)

			assert.Equal(t, test.wantUploadedBytes, progress.UploadedBytes, "uploaded bytes")
			assert.Equal(t, test.wantUploadedFiles, progress.UploadedFiles, "uploaded files")
			assert.Equal(t, test.wantRemainBytes, progress.RemainingBytes, "remaining bytes")
			assert.Equal(t, test.wantRemainFiles, progress.RemainingFiles, "remaining files")
			assert.Equal(t, test.wantTotalBytes, progress.TotalBytes, "total bytes")
			assert.InDelta(t, test.wantPercent, progress.Percent, 0.001, "percent")

			// Every local byte lands in exactly one bucket.
			assert.Equal(t, progress.TotalBytes, progress.UploadedBytes+progress.RemainingBytes)
			assert.Equal(t, progress.TotalFiles, progress.UploadedFiles+progress.RemainingFiles)
		})
	}
}

func TestCloudProgressStale(t *testing.T) {
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		generated time.Time
		wantStale bool
	}{
		{"just generated", now, false},
		{"within the window", now.Add(-25 * time.Hour), false},
		{"past the window", now.Add(-27 * time.Hour), true},
		{"zero value is stale", time.Time{}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			progress := &CloudProgress{Generated: test.generated}
			assert.Equal(t, test.wantStale, progress.Stale(now))
		})
	}
}
