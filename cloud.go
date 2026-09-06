package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// TreeSize is what one `rclone size` call reports about a whole tree.
type TreeSize struct {
	Bytes int64 `json:"bytes"`
	Count int   `json:"count"`
}

// CloudProgress is one measurement of each side of the sync: the local media
// tree and the rclone remote that rclone/cloud_move.sh sends it to.
//
// Only the two measurements are stored. Every figure the page shows - total,
// uploaded, remaining, and both percentages - is derived from them, so the
// numbers on the page cannot disagree with each other or with rclone.
type CloudProgress struct {
	Local     TreeSize  `json:"local"`
	Remote    TreeSize  `json:"remote"`
	Generated time.Time `json:"generated"`
}

// StaleAfter is how old a snapshot may be before the page stops presenting it
// as current. The refresh interval is twelve hours by default, so a snapshot
// past this point means refreshes have been failing rather than merely resting.
const StaleAfter = 26 * time.Hour

// Stale reports whether the snapshot is old enough that the page should say so.
func (progress *CloudProgress) Stale(now time.Time) bool {
	return now.Sub(progress.Generated) > StaleAfter
}

// UploadedBytes is how much of the local tree the remote already holds, capped
// at the local total.
//
// The remote can briefly hold more than the local tree does: a file deleted
// locally survives there until the next sync pass removes it. That drift is
// transient, and an upload reported as more than complete is worse than one
// that stops at complete.
func (progress *CloudProgress) UploadedBytes() int64 {
	if progress.Remote.Bytes > progress.Local.Bytes {
		return progress.Local.Bytes
	}
	return progress.Remote.Bytes
}

// RemainingBytes is what still has to go up.
func (progress *CloudProgress) RemainingBytes() int64 {
	return progress.Local.Bytes - progress.UploadedBytes()
}

// UploadedPercent and RemainingPercent share the local total as their
// denominator, so the two always sum to 100.
func (progress *CloudProgress) UploadedPercent() float64 {
	return percentOf(progress.UploadedBytes(), progress.Local.Bytes)
}

func (progress *CloudProgress) RemainingPercent() float64 {
	return percentOf(progress.RemainingBytes(), progress.Local.Bytes)
}

func percentOf(part, whole int64) float64 {
	if whole <= 0 {
		return 0
	}
	return float64(part) * 100 / float64(whole)
}

// CollectCloudProgress returns the pair of measurements, preferring a recent
// cached snapshot and falling back to any older one if a refresh fails.
//
// It never returns an error. The download figure is worth publishing on its
// own, and a stale measurement beats replacing the whole page with nothing; a
// nil result means there is no figure to show at all.
func CollectCloudProgress(ctx context.Context, config *Config, logger *Logger) *CloudProgress {
	cached, cacheErr := readCloudCache(config.CloudCache)
	if cacheErr == nil {
		if age := time.Since(cached.Generated); age < config.CloudMaxAge {
			logger.Infof("Reusing cloud measurements from %s ago", age.Round(time.Minute))
			return cached
		}
	}

	logger.Infof("Measuring %s and %s", config.CloudSource, config.CloudDest)
	started := time.Now()

	fresh, err := MeasureTrees(ctx, config.CloudSource, config.CloudDest)
	if err != nil {
		logger.Errorf("Cloud measurement failed: %v", err)
		if cacheErr != nil {
			logger.Errorf("No cached cloud progress to fall back on: %v", cacheErr)
			return nil
		}
		logger.Infof("Falling back to cloud measurements from %s", cached.Generated.Format(time.RFC3339))
		return cached
	}

	logger.Infof("Cloud measurement took %s", time.Since(started).Round(time.Second))
	logger.Debugf("Local: %d bytes in %d files", fresh.Local.Bytes, fresh.Local.Count)
	logger.Debugf("Remote: %d bytes in %d files", fresh.Remote.Bytes, fresh.Remote.Count)

	if err := writeCloudCache(config.CloudCache, fresh); err != nil {
		logger.Errorf("Could not cache cloud progress: %v", err)
	}
	return fresh
}

// MeasureTrees sizes both sides, once each.
func MeasureTrees(ctx context.Context, source, dest string) (*CloudProgress, error) {
	var (
		local, remote       TreeSize
		localErr, remoteErr error
		waitGroup           sync.WaitGroup
	)

	// The remote measurement dominates the runtime, so overlap the local walk
	// with it rather than paying for both in sequence.
	waitGroup.Add(2)
	go func() {
		defer waitGroup.Done()
		// fastList buys nothing on a local filesystem, where the walk already
		// takes a couple of seconds.
		local, localErr = MeasureTree(ctx, source, false)
	}()
	go func() {
		defer waitGroup.Done()
		remote, remoteErr = MeasureTree(ctx, dest, true)
	}()
	waitGroup.Wait()

	if localErr != nil {
		return nil, fmt.Errorf("sizing source %s: %w", source, localErr)
	}
	if remoteErr != nil {
		return nil, fmt.Errorf("sizing destination %s: %w", dest, remoteErr)
	}

	return &CloudProgress{Local: local, Remote: remote, Generated: time.Now()}, nil
}

// MeasureTree returns what rclone reports for a whole tree in one call.
//
// Both sides are sized with rclone rather than one of them walked in Go:
// identical tooling gives identical size semantics, which is what makes the two
// figures comparable at all.
//
// fastList collapses a per-directory walk into one recursive listing. Against
// the crypt-over-Dropbox remote that is the difference between roughly a hundred
// seconds and well over ten minutes, so it is essential there.
func MeasureTree(ctx context.Context, target string, fastList bool) (TreeSize, error) {
	args := []string{"size", "--json"}
	if fastList {
		args = append(args, "--fast-list")
	}
	args = append(args, "--", target)

	command := exec.CommandContext(ctx, "rclone", args...)
	var stderr strings.Builder
	command.Stderr = &stderr

	output, err := command.Output()
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if ctx.Err() != nil {
			return TreeSize{}, fmt.Errorf("rclone size of %s did not finish: %w: %s", target, ctx.Err(), message)
		}
		return TreeSize{}, fmt.Errorf("rclone size of %s exited with %w: %s", target, err, message)
	}
	return ParseTreeSize(output)
}

// ParseTreeSize reads the single object `rclone size --json` prints.
func ParseTreeSize(output []byte) (TreeSize, error) {
	var size TreeSize
	if err := json.Unmarshal(output, &size); err != nil {
		return TreeSize{}, fmt.Errorf("parsing rclone size output %q: %w", strings.TrimSpace(string(output)), err)
	}
	// rclone reports -1 for a tree it could list but not size. Everything
	// downstream treats bytes as a count, so refuse it here rather than let it
	// subtract from the other side.
	if size.Bytes < 0 || size.Count < 0 {
		return TreeSize{}, fmt.Errorf("rclone reported an unusable size: %d bytes in %d files", size.Bytes, size.Count)
	}
	return size, nil
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
	// A cache written by an older build parses cleanly into an empty local
	// measurement, which would publish a confident 0 Bytes for up to a whole
	// max-age. Treat it as no cache at all.
	if progress.Local.Bytes <= 0 {
		return nil, fmt.Errorf("cached cloud progress at %s has no local measurement", path)
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
