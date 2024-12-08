package logger

import (
	"strings"
	"syscall"

	"go.uber.org/zap"
)

// Sync attempts to sync the logger, handling common sync issues gracefully.
// It returns nil for expected sync errors that shouldn't impact operation.
// This is particularly useful for handling platform-specific sync behaviors
// and ensuring clean shutdown.
func Sync(logger *zap.Logger) error {
	err := logger.Sync()
	if err == nil {
		return nil
	}

	// Convert to error string for pattern matching
	errStr := err.Error()

	// Handle common stdout/stderr sync issues that can be safely ignored
	if strings.Contains(errStr, "invalid argument") ||
		strings.Contains(errStr, "inappropriate ioctl for device") ||
		strings.Contains(errStr, "bad file descriptor") {
		return nil
	}

	// Handle specific syscall errors that are expected
	if err == syscall.EINVAL {
		return nil
	}

	// Return unexpected sync errors for handling
	return err
}

// MustSync attempts to sync the logger and panics on unexpected errors.
// This should only be used during program shutdown where a panic is acceptable.
func MustSync(logger *zap.Logger) {
	if err := Sync(logger); err != nil {
		panic(err)
	}
}
