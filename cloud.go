package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// CloudProgress is the result of comparing the local media tree against the
// rclone remote that rclone/cloud_move.sh syncs it to.
type CloudProgress struct {
	UploadedBytes  int64     `json:"uploaded_bytes"`
	UploadedFiles  int       `json:"uploaded_files"`
	RemainingBytes int64     `json:"remaining_bytes"`
	RemainingFiles int       `json:"remaining_files"`
	TotalBytes     int64     `json:"total_bytes"`
	TotalFiles     int       `json:"total_files"`
	Percent        float64   `json:"percent"`
	Generated      time.Time `json:"generated"`
}

// StaleAfter is how old a snapshot may be before the page stops presenting it
// as current. The refresh interval is twelve hours by default, so a snapshot
// past this point means refreshes have been failing rather than merely resting.
const StaleAfter = 26 * time.Hour

// Stale reports whether the snapshot is old enough that the page should say so.
func (progress *CloudProgress) Stale(now time.Time) bool {
	return now.Sub(progress.Generated) > StaleAfter
}

// CollectCloudProgress returns the upload comparison, preferring a recent
// cached snapshot and falling back to any older one if a refresh fails.
//
// It never returns an error. The download and media figures are worth
// publishing on their own, and a stale upload number beats replacing the whole
// page with nothing; a nil result means there is no figure to show at all.
func CollectCloudProgress(ctx context.Context, config *Config, logger *Logger) *CloudProgress {
	cached, cacheErr := readCloudCache(config.CloudCache)
	if cacheErr == nil {
		if age := time.Since(cached.Generated); age < config.CloudMaxAge {
			logger.Infof("Reusing cloud comparison from %s ago", age.Round(time.Minute))
			return cached
		}
	}

	logger.Infof("Comparing %s against %s", config.CloudSource, config.CloudDest)
	started := time.Now()

	fresh, err := CompareTrees(ctx, config.CloudSource, config.CloudDest)
	if err != nil {
		logger.Errorf("Cloud comparison failed: %v", err)
		if cacheErr != nil {
			logger.Errorf("No cached cloud progress to fall back on: %v", cacheErr)
			return nil
		}
		logger.Infof("Falling back to cloud comparison from %s", cached.Generated.Format(time.RFC3339))
		return cached
	}

	logger.Infof("Cloud comparison took %s", time.Since(started).Round(time.Second))
	if err := writeCloudCache(config.CloudCache, fresh); err != nil {
		logger.Errorf("Could not cache cloud progress: %v", err)
	}
	return fresh
}

// CompareTrees lists both sides and buckets the local files against the remote.
func CompareTrees(ctx context.Context, source, dest string) (*CloudProgress, error) {
	var (
		localFiles, remoteFiles map[string]int64
		localErr, remoteErr     error
		waitGroup               sync.WaitGroup
	)

	// The remote listing dominates the runtime, so overlap the local walk with it
	// rather than paying for both in sequence.
	waitGroup.Add(2)
	go func() {
		defer waitGroup.Done()
		// fastList buys nothing on a local filesystem, where the walk already
		// takes a couple of seconds.
		localFiles, localErr = listFiles(ctx, source, false)
	}()
	go func() {
		defer waitGroup.Done()
		remoteFiles, remoteErr = listFiles(ctx, dest, true)
	}()
	waitGroup.Wait()

	if localErr != nil {
		return nil, fmt.Errorf("listing source %s: %w", source, localErr)
	}
	if remoteErr != nil {
		return nil, fmt.Errorf("listing destination %s: %w", dest, remoteErr)
	}

	progress := BucketFiles(localFiles, remoteFiles)
	progress.Generated = time.Now()
	return progress, nil
}

// BucketFiles splits the local files into what the remote already holds and
// what still has to go.
//
// Comparing the two totals alone cannot do this: it gives no way to separate
// bytes not yet uploaded from bytes that exist only on the remote. Files present
// remotely but missing locally are deliberately ignored - rclone sync deletes
// them on its next pass, so they are transient drift, not part of the upload.
func BucketFiles(localFiles, remoteFiles map[string]int64) *CloudProgress {
	progress := &CloudProgress{}

	for path, size := range localFiles {
		progress.TotalBytes += size
		progress.TotalFiles++

		// Same path and same size, or it still has to go: a size mismatch means a
		// partial or superseded copy that rclone will send again.
		if remoteSize, found := remoteFiles[path]; found && remoteSize == size {
			progress.UploadedBytes += size
			progress.UploadedFiles++
			continue
		}
		progress.RemainingBytes += size
		progress.RemainingFiles++
	}

	if progress.TotalBytes > 0 {
		progress.Percent = float64(progress.UploadedBytes) * 100 / float64(progress.TotalBytes)
	}
	return progress
}

// listFiles returns path -> size for every file under target.
//
// Both sides are listed with rclone rather than walking the local tree in Go:
// identical tooling gives identical relative paths and identical size semantics,
// which is what makes the two maps comparable key for key.
//
// fastList collapses a per-directory walk into one recursive listing. Against
// the crypt-over-Dropbox remote that is the difference between roughly a hundred
// seconds and well over ten minutes, so it is essential there.
func listFiles(ctx context.Context, target string, fastList bool) (map[string]int64, error) {
	args := []string{"lsf", "-R", "--files-only", "--format", "sp", "--separator", "|"}
	if fastList {
		args = append(args, "--fast-list")
	}
	args = append(args, "--", target)

	command := exec.CommandContext(ctx, "rclone", args...)
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("opening rclone output pipe: %w", err)
	}

	var stderr strings.Builder
	command.Stderr = &stderr

	if err := command.Start(); err != nil {
		return nil, fmt.Errorf("starting rclone: %w", err)
	}

	// Drain the pipe before waiting: rclone blocks once the buffer fills, and a
	// full listing is far larger than the buffer.
	files, parseErr := ParseListing(stdout)

	if err := command.Wait(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if ctx.Err() != nil {
			return nil, fmt.Errorf("rclone listing of %s did not finish: %w: %s", target, ctx.Err(), message)
		}
		return nil, fmt.Errorf("rclone exited with %w: %s", err, message)
	}
	if parseErr != nil {
		return nil, fmt.Errorf("reading rclone output: %w", parseErr)
	}
	return files, nil
}

// ParseListing reads the "size|path" lines produced by rclone lsf.
func ParseListing(reader io.Reader) (map[string]int64, error) {
	files := make(map[string]int64)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Split on the first separator only: a filename may contain a pipe, a
		// size never can.
		separator := strings.Index(line, "|")
		if separator < 0 {
			return nil, fmt.Errorf("malformed listing line %q: no separator", line)
		}

		size, err := strconv.ParseInt(line[:separator], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("malformed listing line %q: %w", line, err)
		}
		files[line[separator+1:]] = size
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return files, nil
}

func readCloudCache(path string) (*CloudProgress, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var progress CloudProgress
	if err := json.Unmarshal(content, &progress); err != nil {
		return nil, fmt.Errorf("parsing cached cloud progress at %s: %w", path, err)
	}
	return &progress, nil
}

func writeCloudCache(path string, progress *CloudProgress) error {
	content, err := json.MarshalIndent(progress, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding cloud progress: %w", err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("writing cloud cache to %s: %w", path, err)
	}
	return nil
}
