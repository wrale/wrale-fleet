// Package logger provides a stage-aware logging infrastructure for the wfdevice command.
package logger

import (
	"os"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new zap.Logger with appropriate configuration for the device agent.
// It supports stage-aware logging and proper environment configuration.
func New() (*zap.Logger, error) {
	cfg := getConfig()
	encConfig := getEncoderConfig()

	// Create appropriate encoder based on configuration
	var encoder zapcore.Encoder
	if cfg.JSONOutput {
		encoder = zapcore.NewJSONEncoder(encConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encConfig)
	}

	// Create core with appropriate level
	baseLevel := parseLogLevel(cfg.LogLevel)
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		baseLevel,
	)

	// Apply sampling if enabled (except for error logs)
	if cfg.Sampling {
		// Create separate core for error logs (unsampled)
		errorCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= zapcore.ErrorLevel
			}),
		)

		// Create sampled core for info and debug logs
		sampledCore := zapcore.NewSamplerWithOptions(
			zapcore.NewCore(
				encoder,
				zapcore.AddSync(os.Stdout),
				zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
					return lvl < zapcore.ErrorLevel && lvl >= baseLevel
				}),
			),
			time.Second, // Tick
			100,         // First
			100,         // Thereafter
		)

		// Combine cores
		core = zapcore.NewTee(errorCore, sampledCore)
	}

	// Build logger with appropriate options
	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)

	// Add stack traces for errors if enabled
	if cfg.StackTrace {
		logger = logger.WithOptions(zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// Add global fields for log aggregation and filtering
	logger = logger.With(
		zap.String("environment", cfg.Environment),
		zap.String("app", "wfdevice"),
		zap.Time("boot_time", time.Now().UTC()),
		zap.Int("stage", cfg.Stage),
	)

	return logger, nil
}

// NewNop creates a no-op logger useful for testing and development.
// All operations on the returned logger will succeed but will not
// produce any output.
func NewNop() *zap.Logger {
	return zap.NewNop()
}

// NewTest creates a logger suitable for testing with output
// captured by the testing framework.
func NewTest(tb testing.TB) *zap.Logger {
	return zap.NewExample()
}
