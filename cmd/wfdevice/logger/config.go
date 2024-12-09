// Package logger provides a stage-aware logging infrastructure for the wfdevice command.
package logger

import (
	"os"

	"go.uber.org/zap/zapcore"
)

// Config holds the configuration for the device agent logger
type Config struct {
	Environment string // "production", "staging", "development"
	LogLevel    string // "debug", "info", "warn", "error"
	Sampling    bool   // Enable sampling for high-volume logs
	JSONOutput  bool   // Use JSON output format
	StackTrace  bool   // Include stack traces for errors
	Stage       int    // Current capability stage (1-6)
}

// getConfig determines logging configuration based on environment variables
func getConfig() Config {
	config := Config{
		Environment: os.Getenv("ENVIRONMENT"),
		LogLevel:    os.Getenv("LOG_LEVEL"),
		Sampling:    os.Getenv("LOG_SAMPLING") != "false",
		JSONOutput:  os.Getenv("LOG_JSON") == "true",
		StackTrace:  os.Getenv("LOG_STACKTRACE") != "false",
		Stage:       1, // Default to Stage 1 capabilities
	}

	// Set defaults based on environment
	switch config.Environment {
	case "production":
		if config.LogLevel == "" {
			config.LogLevel = "info"
		}
		if !config.JSONOutput {
			config.JSONOutput = true
		}
	case "staging":
		if config.LogLevel == "" {
			config.LogLevel = "debug"
		}
		if !config.JSONOutput {
			config.JSONOutput = true
		}
	default: // development
		config.Environment = "development"
		if config.LogLevel == "" {
			config.LogLevel = "debug"
		}
		config.Sampling = false // Disable sampling in development
	}

	return config
}

// getEncoderConfig creates a zapcore.EncoderConfig with standardized settings
func getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
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
}

// parseLogLevel converts a string log level to zapcore.Level
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
