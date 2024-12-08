package main

import (
	"os"
	"strings"
	"syscall"

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

// safeSync attempts to sync the logger, handling common sync issues gracefully.
// It returns nil for expected sync errors that shouldn't impact application operation.
func safeSync(logger *zap.Logger) error {
	err := logger.Sync()
	if err == nil {
		return nil
	}

	// Check for common sync issues that can be safely ignored
	if strings.Contains(err.Error(), "inappropriate ioctl for device") {
		return nil // Common stdout/stderr sync issue
	}
	if err == syscall.EINVAL {
		return nil // Another common sync error
	}
	if strings.Contains(err.Error(), "bad file descriptor") {
		return nil // Common during shutdown
	}

	// Return other sync errors for handling
	return err
}

// getLoggerLevel extracts the configured level from a zap.Logger.
// This is primarily used for testing and verification.
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
