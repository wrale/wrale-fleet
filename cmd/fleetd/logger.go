package main

import (
	"os"

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
