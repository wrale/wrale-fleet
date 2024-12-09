// Package logger provides logging configuration for the wfcentral command.
package logger

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config defines logging configuration.
type Config struct {
	Level    string // Log level: debug, info, warn, error
	FilePath string // Optional file path for log output (empty for stdout)
}

// New creates a new logger with the given configuration.
// It supports both file and stdout output with proper error handling
// for airgapped environments.
func New(cfg Config) (*zap.Logger, error) {
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", cfg.Level, err)
	}

	// Create encoder config with standardized settings
	encConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create the core with appropriate output
	var output zapcore.WriteSyncer
	if cfg.FilePath != "" {
		// Ensure parent directory exists with restricted permissions
		// 0750 allows owner full access and group read/execute only
		if err := os.MkdirAll(strings.TrimSuffix(cfg.FilePath, "/"), 0750); err != nil {
			return nil, fmt.Errorf("creating log directory: %w", err)
		}

		// Open log file with restricted permissions
		// 0600 ensures only the owner can read/write the log files
		f, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, fmt.Errorf("opening log file: %w", err)
		}
		output = zapcore.AddSync(f)
	} else {
		output = zapcore.AddSync(os.Stdout)
	}

	// Create the core with JSON encoding for structured logging
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encConfig),
		output,
		level,
	)

	// Build the logger with appropriate options
	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	// Add global fields for log aggregation and filtering
	logger = logger.With(
		zap.String("component", "wfcentral"),
		zap.Time("boot_time", time.Now().UTC()),
	)

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

// NewNop creates a no-op logger useful for testing.
// All operations on the returned logger will succeed but will not
// produce any output.
func NewNop() *zap.Logger {
	return zap.NewNop()
}
