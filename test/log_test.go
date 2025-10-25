package test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/snowpea/stats/internal/logger"
)

func TestOutLog(t *testing.T) {
	// Set log level for testing
	logger.SetLogLevel("Normal")

	tests := []struct {
		name     string
		logLevel string
		text     string
		obj      interface{}
		expected string
	}{
		{
			name:     "Normal log level with text",
			logLevel: "Normal",
			text:     "Test message",
			obj:      nil,
			expected: "Test message",
		},
		{
			name:     "Normal log level with object",
			logLevel: "Normal",
			text:     "Test message",
			obj:      map[string]string{"key": "value"},
			expected: "Test message",
		},
		{
			name:     "None log level",
			logLevel: "None",
			text:     "Test message",
			obj:      nil,
			expected: "",
		},
		{
			name:     "Debug log level",
			logLevel: "Debug",
			text:     "Debug message",
			obj:      nil,
			expected: "Debug message",
		},
		{
			name:     "Empty text",
			logLevel: "Normal",
			text:     "",
			obj:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set log level
			logger.SetLogLevel(tt.logLevel)

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			logger.OutLog(tt.text, tt.obj)

			// Close write end and restore stdout
			w.Close()
			os.Stdout = old

			// Read captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if tt.logLevel == "None" {
				if output != "" {
					t.Errorf("OutLog() with logLevel='None' should produce no output, got: %s", output)
				}
				return
			}

			if tt.text == "" && tt.obj == nil {
				if output != "" {
					t.Errorf("OutLog() with empty text and nil obj should produce no output, got: %s", output)
				}
				return
			}

			// Check that output contains expected text
			if tt.text != "" && !strings.Contains(output, tt.text) {
				t.Errorf("OutLog() output should contain '%s', got: %s", tt.text, output)
			}

			// Check timestamp format
			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) > 0 {
				firstLine := lines[0]
				// Check if line starts with timestamp (DD.MM.YYYY HH:MM:SS:)
				if !strings.Contains(firstLine, ":") {
					t.Errorf("OutLog() should include timestamp, got: %s", firstLine)
				}
			}
		})
	}
}

func TestOutLogTimestamp(t *testing.T) {
	// Set log level for testing
	logger.SetLogLevel("Normal")

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger.OutLog("Test message", nil)

	// Close write end and restore stdout
	w.Close()
	os.Stdout = old

	// Read captured output
	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := strings.TrimSpace(buf.String())

	// Check timestamp format (DD.MM.YYYY HH:MM:SS:)
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		firstLine := lines[0]
		// Should contain timestamp and message
		parts := strings.Split(firstLine, " ")
		if len(parts) < 2 {
			t.Errorf("OutLog() should include timestamp and message, got: %s", firstLine)
		}

		// Check timestamp format
		timestamp := parts[0]
		// The timestamp format is "DD.MM.YYYY HH:MM:SS:" with a colon at the end
		// But we need to handle the case where the timestamp might be split across parts
		if len(parts) >= 2 {
			// If we have multiple parts, the timestamp might be split
			// Let's check if the first part ends with a colon
			if strings.HasSuffix(timestamp, ":") {
				// Remove the trailing colon for parsing
				timestampWithoutColon := strings.TrimSuffix(timestamp, ":")
				_, err := time.Parse("02.01.2006 15:04:05", timestampWithoutColon)
				if err != nil {
					t.Errorf("OutLog() timestamp format should be DD.MM.YYYY HH:MM:SS, got: %s", timestampWithoutColon)
				}
			} else {
				// The timestamp might be split across parts, let's check the first two parts
				if len(parts) >= 2 {
					fullTimestamp := timestamp + " " + parts[1]
					if strings.HasSuffix(fullTimestamp, ":") {
						timestampWithoutColon := strings.TrimSuffix(fullTimestamp, ":")
						_, err := time.Parse("02.01.2006 15:04:05", timestampWithoutColon)
						if err != nil {
							t.Errorf("OutLog() timestamp format should be DD.MM.YYYY HH:MM:SS, got: %s", timestampWithoutColon)
						}
					} else {
						t.Errorf("OutLog() timestamp should end with colon, got: %s", fullTimestamp)
					}
				} else {
					t.Errorf("OutLog() timestamp should end with colon, got: %s", timestamp)
				}
			}
		} else {
			t.Errorf("OutLog() should include timestamp and message, got: %s", firstLine)
		}
	}
}
