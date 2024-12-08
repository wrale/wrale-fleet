package main

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// setupLogger configures the application logger based on the environment.
// It returns a configured zap.Logger with appropriate log levels and encoding.
func setupLogger() (*zap.Logger, error) {
	var config zap.Config

	if env := os.Getenv("ENVIRONMENT"); env == "production" {
		// Production configuration
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	} else {
		// Development configuration
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	}

	// Common configuration
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	return config.Build()
}

// safeSync attempts to sync the logger, ignoring common "bad file descriptor" errors
// that occur when syncing stdout/stderr in tests
func safeSync(logger *zap.Logger) error {
	err := logger.Sync()
	if err != nil && !strings.Contains(err.Error(), "bad file descriptor") {
		return err
	}
	return nil
}

// getLoggerLevel extracts the configured level from a zap.Logger
func getLoggerLevel(logger *zap.Logger) zapcore.Level {
	// Type assert to get the atomic level
	if atomic, ok := logger.Core().(interface{ Level() zapcore.Level }); ok {
		return atomic.Level()
	}
	// Fallback to checking each level
	for l := zapcore.DebugLevel; l <= zapcore.FatalLevel; l++ {
		if logger.Core().Enabled(l) {
			return l
		}
	}
	return zapcore.InfoLevel // Default fallback
}
