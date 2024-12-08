package main

import (
	"os"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggingConfig holds the configuration for the application logger
type LoggingConfig struct {
	Environment string // "production", "staging", "development"
	LogLevel    string // "debug", "info", "warn", "error"
	Sampling    bool   // Enable sampling for high-volume logs
	JSONOutput  bool   // Use JSON output format
	StackTrace  bool   // Include stack traces for errors
}

// getLoggingConfig determines logging configuration based on environment variables
func getLoggingConfig() LoggingConfig {
	config := LoggingConfig{
		Environment: os.Getenv("ENVIRONMENT"),
		LogLevel:    os.Getenv("LOG_LEVEL"),
		Sampling:    os.Getenv("LOG_SAMPLING") != "false",
		JSONOutput:  os.Getenv("LOG_JSON") == "true",
		StackTrace:  os.Getenv("LOG_STACKTRACE") != "false",
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

// setupLogger configures the application logger based on environment variables.
// It returns a configured zap.Logger with appropriate log levels and encoding.
func setupLogger() (*zap.Logger, error) {
	cfg := getLoggingConfig()

	// Base encoder config with improved timestamp format
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

	// Create the encoder based on configuration
	var encoder zapcore.Encoder
	if cfg.JSONOutput {
		encoder = zapcore.NewJSONEncoder(encConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encConfig)
	}

	// Create the core with appropriate level and sampling
	baseLevel := parseLogLevel(cfg.LogLevel)
	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		baseLevel,
	)

	// Apply sampling if enabled (except for error logs)
	if cfg.Sampling {
		// Create a tee of cores - one for error logs (unsampled) and one for everything else
		errorCore := zapcore.NewCore(
			encoder,
			zapcore.AddSync(os.Stdout),
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return lvl >= zapcore.ErrorLevel
			}),
		)

		// Sampled core for info and below
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

		// Combine the cores
		core = zapcore.NewTee(errorCore, sampledCore)
	}

	// Build the logger
	logger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)

	// Add stack traces for errors if enabled
	if cfg.StackTrace {
		logger = logger.WithOptions(zap.AddStacktrace(zapcore.ErrorLevel))
	}

	// Add global fields
	logger = logger.With(
		zap.String("environment", cfg.Environment),
		zap.String("app", "fleetd"),
		zap.Time("boot_time", time.Now().UTC()),
	)

	return logger, nil
}

// parseLogLevel converts a string log level to zapcore.Level
func parseLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
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

// selectEncoding returns the appropriate encoding based on configuration
func selectEncoding(useJSON bool) string {
	if useJSON {
		return "json"
	}
	return "console"
}

// safeSync attempts to sync the logger, handling common sync issues gracefully.
// It returns nil for expected sync errors that shouldn't impact application operation.
func safeSync(logger *zap.Logger) error {
	err := logger.Sync()
	if err == nil {
		return nil
	}

	// Convert to error string for pattern matching
	errStr := err.Error()

	// Common stdout/stderr sync issues that can be safely ignored
	if strings.Contains(errStr, "invalid argument") ||
		strings.Contains(errStr, "inappropriate ioctl for device") ||
		strings.Contains(errStr, "bad file descriptor") {
		return nil
	}

	// Check for specific syscall errors
	if err == syscall.EINVAL {
		return nil
	}

	// Return unexpected sync errors for handling
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
