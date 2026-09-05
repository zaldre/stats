package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
		wantErr  bool
	}{
		{name: "zero", input: 0, expected: "0 Bytes"},
		{name: "bytes stay whole", input: 512, expected: "512 Bytes"},
		{name: "kilobytes", input: 2048, expected: "2.00 KB"},
		{name: "megabytes", input: 5 * 1024 * 1024, expected: "5.00 MB"},
		{name: "terabytes", input: 21468607036351, expected: "19.53 TB"},
		{name: "negative is an error", input: -1, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := FormatBytes(test.input)
			if test.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestReadMaintenanceNotice(t *testing.T) {
	dir := t.TempDir()

	t.Run("missing file is not an error", func(t *testing.T) {
		notice, err := ReadMaintenanceNotice(filepath.Join(dir, "absent.txt"))
		require.NoError(t, err)
		assert.Empty(t, notice)
	})

	t.Run("content is trimmed", func(t *testing.T) {
		path := filepath.Join(dir, "notice.txt")
		require.NoError(t, os.WriteFile(path, []byte("  Down Sunday 0200\n\n"), 0o644))

		notice, err := ReadMaintenanceNotice(path)
		require.NoError(t, err)
		assert.Equal(t, "Down Sunday 0200", notice)
	})

	t.Run("whitespace-only file reads as no notice", func(t *testing.T) {
		path := filepath.Join(dir, "blank.txt")
		require.NoError(t, os.WriteFile(path, []byte("\n  \n"), 0o644))

		notice, err := ReadMaintenanceNotice(path)
		require.NoError(t, err)
		assert.Empty(t, notice)
	})
}

func TestDescribeCloudProgress(t *testing.T) {
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)

	t.Run("nil progress is unavailable", func(t *testing.T) {
		uploaded, remaining := DescribeCloudProgress(nil, now)
		assert.Equal(t, Unavailable, uploaded)
		assert.Equal(t, Unavailable, remaining)
	})

	t.Run("fresh snapshot carries the percentage and no timestamp", func(t *testing.T) {
		uploaded, remaining := DescribeCloudProgress(&CloudProgress{
			UploadedBytes:  21468607036351,
			RemainingBytes: 29704550472391,
			Percent:        41.95,
			Generated:      now.Add(-time.Hour),
		}, now)

		assert.Equal(t, "19.53 TB (42.0%)", uploaded)
		assert.Equal(t, "27.02 TB", remaining)
	})

	t.Run("stale snapshot is labelled on both figures", func(t *testing.T) {
		uploaded, remaining := DescribeCloudProgress(&CloudProgress{
			UploadedBytes:  1024,
			RemainingBytes: 2048,
			Percent:        33.3,
			Generated:      now.Add(-48 * time.Hour),
		}, now)

		assert.Contains(t, uploaded, "as of")
		assert.Contains(t, remaining, "as of")
	})
}

func TestMeasureMediaDirs(t *testing.T) {
	logger := NewLogger(LogNone)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "a.bin"), make([]byte, 1500), 0o644))

	t.Run("sums file sizes", func(t *testing.T) {
		total, err := MeasureMediaDirs([]string{dir}, logger)
		require.NoError(t, err)
		assert.Equal(t, int64(1500), total)
	})

	t.Run("a missing directory alongside a good one is skipped", func(t *testing.T) {
		total, err := MeasureMediaDirs([]string{dir, filepath.Join(dir, "absent")}, logger)
		require.NoError(t, err)
		assert.Equal(t, int64(1500), total)
	})

	t.Run("all directories missing is an error", func(t *testing.T) {
		_, err := MeasureMediaDirs([]string{filepath.Join(dir, "absent")}, logger)
		require.Error(t, err)
	})
}
