package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTreeSize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected TreeSize
		wantErr  string
	}{
		{
			name:     "a sized tree",
			input:    `{"count":1234,"bytes":21468607036351}`,
			expected: TreeSize{Bytes: 21468607036351, Count: 1234},
		},
		{
			name:     "an empty tree",
			input:    `{"count":0,"bytes":0}`,
			expected: TreeSize{},
		},
		{
			name:     "fields rclone adds are ignored",
			input:    `{"count":2,"bytes":300,"sizeless":1}`,
			expected: TreeSize{Bytes: 300, Count: 2},
		},
		{
			name:    "unparseable output is an error",
			input:   "Total objects: 2\n",
			wantErr: "parsing rclone size output",
		},
		{
			name:    "a tree rclone could not size is an error",
			input:   `{"count":2,"bytes":-1}`,
			wantErr: "unusable size",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			size, err := ParseTreeSize([]byte(test.input))
			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, size)
		})
	}
}

func TestCloudProgressFigures(t *testing.T) {
	tests := []struct {
		name              string
		local             TreeSize
		remote            TreeSize
		wantUploadedBytes int64
		wantRemainBytes   int64
		wantUploadedPct   float64
		wantRemainPct     float64
	}{
		{
			name:            "nothing uploaded yet",
			local:           TreeSize{Bytes: 400, Count: 2},
			remote:          TreeSize{},
			wantRemainBytes: 400,
			wantRemainPct:   100,
		},
		{
			name:              "partly uploaded",
			local:             TreeSize{Bytes: 400, Count: 2},
			remote:            TreeSize{Bytes: 100, Count: 1},
			wantUploadedBytes: 100,
			wantRemainBytes:   300,
			wantUploadedPct:   25,
			wantRemainPct:     75,
		},
		{
			name:              "fully uploaded",
			local:             TreeSize{Bytes: 400, Count: 2},
			remote:            TreeSize{Bytes: 400, Count: 2},
			wantUploadedBytes: 400,
			wantUploadedPct:   100,
		},
		{
			name: "a remote holding more than the local tree stops at complete",
			// A file deleted locally survives on the remote until the next sync
			// pass, which must not read as more than 100% uploaded.
			local:             TreeSize{Bytes: 400, Count: 2},
			remote:            TreeSize{Bytes: 9999, Count: 30},
			wantUploadedBytes: 400,
			wantUploadedPct:   100,
		},
		{
			name:   "an empty source does not divide by zero",
			local:  TreeSize{},
			remote: TreeSize{Bytes: 1, Count: 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			progress := &CloudProgress{Local: test.local, Remote: test.remote}

			assert.Equal(t, test.wantUploadedBytes, progress.UploadedBytes(), "uploaded bytes")
			assert.Equal(t, test.wantRemainBytes, progress.RemainingBytes(), "remaining bytes")
			assert.InDelta(t, test.wantUploadedPct, progress.UploadedPercent(), 0.001, "uploaded percent")
			assert.InDelta(t, test.wantRemainPct, progress.RemainingPercent(), 0.001, "remaining percent")

			// Every local byte lands in exactly one of the two figures.
			assert.Equal(t, test.local.Bytes, progress.UploadedBytes()+progress.RemainingBytes())
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
