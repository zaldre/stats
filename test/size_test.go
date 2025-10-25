package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/snowpea/stats/internal/config"
	"github.com/snowpea/stats/pkg/size"
)

func TestGetMediaSize(t *testing.T) {
	// Create temporary directories for testing
	tempDir := t.TempDir()
	dir1 := filepath.Join(tempDir, "dir1")
	dir2 := filepath.Join(tempDir, "dir2")

	err := os.MkdirAll(dir1, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	err = os.MkdirAll(dir2, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create test files with known sizes
	testFile1 := filepath.Join(dir1, "test1.txt")
	testFile2 := filepath.Join(dir2, "test2.txt")

	// Create files with specific sizes
	err = os.WriteFile(testFile1, []byte("test content 1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = os.WriteFile(testFile2, []byte("test content 2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create config with test directories
	cfg := &config.Config{
		MediaDirs: []string{dir1, dir2},
		LogLevel:  "None",
	}

	// Get the actual size using du command
	expectedSize := int64(0)
	for _, dir := range cfg.MediaDirs {
		cmd := exec.Command("du", "-sb", dir)
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to get directory size: %v", err)
		}

		parts := strings.Split(string(output), "\t")
		if len(parts) < 1 {
			t.Fatalf("Invalid du output: %s", string(output))
		}

		size, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
		if err != nil {
			t.Fatalf("Failed to parse size: %v", err)
		}

		expectedSize += size
	}

	// Test GetMediaSize
	result, err := size.GetMediaSize(cfg)
	if err != nil {
		t.Errorf("GetMediaSize() error: %v", err)
	}

	// The result should be close to expected size (allowing for some filesystem overhead)
	if result < expectedSize {
		t.Errorf("GetMediaSize() = %d, expected at least %d", result, expectedSize)
	}
}

func TestGetMediaSizeWithNonExistentDirectory(t *testing.T) {
	// Create config with non-existent directories
	cfg := &config.Config{
		MediaDirs: []string{"/non/existent/dir1", "/non/existent/dir2"},
		LogLevel:  "None",
	}

	// GetMediaSize should handle non-existent directories gracefully
	result, err := size.GetMediaSize(cfg)
	if err != nil {
		t.Errorf("GetMediaSize() should handle non-existent directories gracefully, got error: %v", err)
	}

	// Result should be 0 for non-existent directories
	if result != 0 {
		t.Errorf("GetMediaSize() with non-existent directories should return 0, got %d", result)
	}
}

func TestGetMediaSizeWithMixedDirectories(t *testing.T) {
	// Create one valid directory and one invalid
	tempDir := t.TempDir()
	validDir := filepath.Join(tempDir, "valid")
	invalidDir := "/non/existent/dir"

	err := os.MkdirAll(validDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(validDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create config with mixed valid/invalid directories
	cfg := &config.Config{
		MediaDirs: []string{validDir, invalidDir},
		LogLevel:  "None",
	}

	// Get the expected size for the valid directory
	cmd := exec.Command("du", "-sb", validDir)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get directory size: %v", err)
	}

	parts := strings.Split(string(output), "\t")
	if len(parts) < 1 {
		t.Fatalf("Invalid du output: %s", string(output))
	}

	expectedSize, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64)
	if err != nil {
		t.Fatalf("Failed to parse size: %v", err)
	}

	// Test GetMediaSize
	result, err := size.GetMediaSize(cfg)
	if err != nil {
		t.Errorf("GetMediaSize() error: %v", err)
	}

	// Result should be the size of the valid directory only
	if result != expectedSize {
		t.Errorf("GetMediaSize() = %d, expected %d", result, expectedSize)
	}
}

func TestGetMediaSizeEmptyDirectories(t *testing.T) {
	// Create empty directories
	tempDir := t.TempDir()
	emptyDir1 := filepath.Join(tempDir, "empty1")
	emptyDir2 := filepath.Join(tempDir, "empty2")

	err := os.MkdirAll(emptyDir1, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	err = os.MkdirAll(emptyDir2, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create config with empty directories
	cfg := &config.Config{
		MediaDirs: []string{emptyDir1, emptyDir2},
		LogLevel:  "None",
	}

	// Test GetMediaSize
	result, err := size.GetMediaSize(cfg)
	if err != nil {
		t.Errorf("GetMediaSize() error: %v", err)
	}

	// Result should be 0 for empty directories
	if result != 0 {
		t.Errorf("GetMediaSize() with empty directories should return 0, got %d", result)
	}
}
