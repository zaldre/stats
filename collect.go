package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
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

// DescribeCloudProgress renders the three tree figures for the page, flagging a
// snapshot old enough that it no longer describes the current state.
//
// All three come from the same pair of measurements, so they always agree:
// uploaded and remaining add up to the total, and their percentages to 100.
func DescribeCloudProgress(progress *CloudProgress, now time.Time) (total, uploaded, remaining string) {
	if progress == nil {
		return Unavailable, Unavailable, Unavailable
	}

	totalSize, totalErr := FormatBytes(progress.Local.Bytes)
	uploadedSize, uploadedErr := FormatBytes(progress.UploadedBytes())
	remainingSize, remainingErr := FormatBytes(progress.RemainingBytes())
	if totalErr != nil || uploadedErr != nil || remainingErr != nil {
		return Unavailable, Unavailable, Unavailable
	}

	// One timestamp for all three figures: they describe a single snapshot, so
	// either all of them are current or none of them are.
	var asOf string
	if progress.Stale(now) {
		asOf = "as of " + progress.Generated.Local().Format("2 Jan 15:04")
	}

	// The percentage and the timestamp share one set of parentheses, so a stale
	// figure reads as one qualified number rather than two.
	qualify := func(size, note string) string {
		if asOf != "" {
			if note != "" {
				note += ", "
			}
			note += asOf
		}
		if note == "" {
			return size
		}
		return fmt.Sprintf("%s (%s)", size, note)
	}

	return qualify(totalSize, ""),
		qualify(uploadedSize, fmt.Sprintf("%.1f%%", progress.UploadedPercent())),
		qualify(remainingSize, fmt.Sprintf("%.1f%%", progress.RemainingPercent()))
}
