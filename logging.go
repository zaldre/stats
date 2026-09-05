package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// LogLevel controls how much of the run is narrated. The three names match the
// LOGLEVEL values the CronJob already sets.
type LogLevel int

const (
	LogNone LogLevel = iota
	LogNormal
	LogDebug
)

// ParseLogLevel is deliberately case-insensitive. The previous implementation
// compared against "Debug" in some places and "debug" in others, so a operator
// setting LOGLEVEL=debug got partial output and no indication why.
func ParseLogLevel(raw string) LogLevel {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "none":
		return LogNone
	case "debug":
		return LogDebug
	default:
		return LogNormal
	}
}

// Logger writes timestamped lines, keeping the format the existing log scrapers
// already see. Errors go to stderr so a failed run is visible in kubectl logs
// even at LOGLEVEL=None.
type Logger struct {
	level  LogLevel
	stdout io.Writer
	stderr io.Writer
	now    func() time.Time
}

func NewLogger(level LogLevel) *Logger {
	return &Logger{level: level, stdout: os.Stdout, stderr: os.Stderr, now: time.Now}
}

const logTimestampLayout = "02.01.2006 15:04:05:"

func (logger *Logger) write(target io.Writer, format string, args ...any) {
	fmt.Fprintf(target, "%s %s\n", logger.now().Format(logTimestampLayout), fmt.Sprintf(format, args...))
}

// Infof records application flow: what the run is doing and how far it got.
func (logger *Logger) Infof(format string, args ...any) {
	if logger.level >= LogNormal {
		logger.write(logger.stdout, format, args...)
	}
}

// Debugf records the detail behind the flow - resolved values, timings, the
// generated page itself.
func (logger *Logger) Debugf(format string, args ...any) {
	if logger.level >= LogDebug {
		logger.write(logger.stdout, format, args...)
	}
}

// Errorf records a condition worth alerting on. It is never suppressed by the
// log level: something the operator needs to see should not depend on how
// chatty they asked the run to be.
func (logger *Logger) Errorf(format string, args ...any) {
	logger.write(logger.stderr, format, args...)
}
