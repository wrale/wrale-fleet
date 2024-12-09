// Package logger provides logging configuration for the wfcentral command.
package logger

import (
	"fmt"
	"strings"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config defines logging configuration.
type Config struct {
	Level string
}

// New creates a new logger with the given configuration.
func New(cfg Config) (*zap.Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", cfg.Level, err)
	}

	zapCfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(level),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := zapCfg.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, fmt.Errorf("building logger: %w", err)
	}

	return logger, nil
}

// Sync safely syncs the logger, handling expected platform-specific errors.
// A nil logger is treated as a no-op case, returning nil to support graceful
// shutdown scenarios. This aligns with the platform's high availability goals
// by handling edge cases robustly.
func Sync(logger *zap.Logger) error {
	// Handle nil logger case gracefully
	if logger == nil {
		return nil
	}

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
