package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Unavailable is what the page shows for a figure this run could not obtain.
// Every collector degrades to it rather than failing the run: a page missing
// one number is far more useful than no page at all.
const Unavailable = "Unavailable"

// FormatBytes renders a byte count in binary units.
func FormatBytes(byteCount int64) (string, error) {
	if byteCount < 0 {
		return "", fmt.Errorf("byte count cannot be negative: %d", byteCount)
	}
	if byteCount == 0 {
		return "0 Bytes", nil
	}

	units := []string{"Bytes", "KB", "MB", "GB", "TB", "PB"}
	unitIndex := int(math.Floor(math.Log(float64(byteCount)) / math.Log(1024)))
	if unitIndex >= len(units) {
		unitIndex = len(units) - 1
	}
	if unitIndex == 0 {
		return fmt.Sprintf("%d Bytes", byteCount), nil
	}

	size := float64(byteCount) / math.Pow(1024, float64(unitIndex))
	return fmt.Sprintf("%.2f %s", size, units[unitIndex]), nil
}

// sabQueue is the slice of the SABnzbd queue response this program uses.
type sabQueue struct {
	Queue struct {
		MBLeft string `json:"mbleft"`
	} `json:"queue"`
}

// FetchDownloadSize returns the bytes left in the SABnzbd queue.
func FetchDownloadSize(ctx context.Context, config *Config) (int64, error) {
	url := fmt.Sprintf("%s:%d/sabnzbd/api?mode=queue&output=json&apikey=%s",
		config.SABHost, config.SABPort, config.SABAPIKey)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("building SABnzbd request: %w", err)
	}

	client := &http.Client{Timeout: config.WebTimeout}
	response, err := client.Do(request)
	if err != nil {
		return 0, fmt.Errorf("querying SABnzbd: %w", err)
	}
	defer func() {
		// Nothing useful can be done if the body fails to close and the figure has
		// already been read, but discarding the error silently is worse than
		// saying so.
		if closeErr := response.Body.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "closing SABnzbd response body: %v\n", closeErr)
		}
	}()

	if response.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("SABnzbd returned HTTP %d", response.StatusCode)
	}

	var queue sabQueue
	if err := json.NewDecoder(response.Body).Decode(&queue); err != nil {
		return 0, fmt.Errorf("decoding SABnzbd response: %w", err)
	}

	megabytesLeft, err := strconv.ParseFloat(queue.Queue.MBLeft, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing SABnzbd mbleft %q: %w", queue.Queue.MBLeft, err)
	}

	megabytesLeft = math.Round(megabytesLeft*100) / 100
	return int64(megabytesLeft * 1024 * 1024), nil
}

// MeasureMediaDirs sums the apparent size of every file under the configured
// media directories.
//
// Unreadable entries are skipped rather than aborting the walk: one bad file on
// an NFS mount should not cost the whole figure. A directory that cannot be
// opened at all is reported, since that usually means a missing mount.
func MeasureMediaDirs(dirs []string, logger *Logger) (int64, error) {
	var total int64
	var measured int

	for _, dir := range dirs {
		size, err := measureDir(dir)
		if err != nil {
			logger.Errorf("Could not measure %s: %v", dir, err)
			continue
		}
		total += size
		measured++
	}

	if measured == 0 {
		return 0, fmt.Errorf("none of the %d media directories could be measured", len(dirs))
	}
	return total, nil
}

func measureDir(dir string) (int64, error) {
	// Check the root before walking. WalkDir reports an unopenable root through
	// the same callback as any other entry, and that callback deliberately
	// swallows errors so one bad file cannot cost the whole figure - which meant
	// an absent NFS mount measured as a silent zero rather than as a failure.
	if _, err := os.Stat(dir); err != nil {
		return 0, err
	}

	var size int64
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			// Keep walking past a single unreadable entry.
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return nil
		}
		size += info.Size()
		return nil
	})
	if err != nil {
		return 0, err
	}
	return size, nil
}

// ReadMaintenanceNotice reads the optional maintenance banner. No file means
// there is nothing to announce, which is the normal case and not an error.
func ReadMaintenanceNotice(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("reading maintenance notice at %s: %w", path, err)
	}
	return strings.TrimSpace(string(content)), nil
}

// DescribeCloudProgress renders the two upload figures for the page, flagging a
// snapshot old enough that it no longer describes the current state.
func DescribeCloudProgress(progress *CloudProgress, now time.Time) (uploaded, remaining string) {
	if progress == nil {
		return Unavailable, Unavailable
	}

	uploaded, err := FormatBytes(progress.UploadedBytes)
	if err != nil {
		return Unavailable, Unavailable
	}
	remaining, err = FormatBytes(progress.RemainingBytes)
	if err != nil {
		return Unavailable, Unavailable
	}

	percent := fmt.Sprintf("%.1f%%", progress.Percent)
	if progress.Stale(now) {
		asOf := progress.Generated.Local().Format("2 Jan 15:04")
		return fmt.Sprintf("%s (%s, as of %s)", uploaded, percent, asOf),
			fmt.Sprintf("%s (as of %s)", remaining, asOf)
	}
	return fmt.Sprintf("%s (%s)", uploaded, percent), remaining
}
