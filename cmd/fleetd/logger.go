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

// createSamplingCore wraps a zapcore.Core with appropriate sampling configuration
// based on the log level. This ensures we maintain important logs while reducing
// volume for less critical entries.
func createSamplingCore(core zapcore.Core, cfg LoggingConfig) zapcore.Core {
	if !cfg.Sampling {
		return core
	}

	// Create a tee of cores with different sampling configs for different levels
	var cores []zapcore.Core

	// Never sample error or above
	errorCore := core.With([]zapcore.Field{
		zap.String("sampling", "none"),
	})
	errorCore = zapcore.NewIncrementLogLevelCore(errorCore, zapcore.ErrorLevel)
	cores = append(cores, errorCore)

	// Sample warning logs: keep first 100, then sample 1/10
	warnCore := zapcore.NewSampler(
		zapcore.NewIncrementLogLevelCore(core, zapcore.WarnLevel),
		time.Second,
		100,
		10,
	)
	warnCore = warnCore.With([]zapcore.Field{
		zap.String("sampling", "warn"),
	})
	cores = append(cores, warnCore)

	// Sample info logs: keep first 100, then sample 1/100
	infoCore := zapcore.NewSampler(
		zapcore.NewIncrementLogLevelCore(core, zapcore.InfoLevel),
		time.Second,
		100,
		100,
	)
	infoCore = infoCore.With([]zapcore.Field{
		zap.String("sampling", "info"),
	})
	cores = append(cores, infoCore)

	// Heavily sample debug logs: keep first 50, then sample 1/1000
	debugCore := zapcore.NewSampler(
		zapcore.NewIncrementLogLevelCore(core, zapcore.DebugLevel),
		time.Second,
		50,
		1000,
	)
	debugCore = debugCore.With([]zapcore.Field{
		zap.String("sampling", "debug"),
	})
	cores = append(cores, debugCore)

	return zapcore.NewTee(cores...)
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

	// Create zap config based on environment
	zapConfig := zap.Config{
		Level:            zap.NewAtomicLevelAt(parseLogLevel(cfg.LogLevel)),
		Development:      cfg.Environment == "development",
		Encoding:         selectEncoding(cfg.JSONOutput),
		EncoderConfig:    encConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// Configure stack traces
	if !cfg.StackTrace {
		zapConfig.DisableStacktrace = true
	}

	// Create the base logger
	baseLogger, err := zapConfig.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	// If sampling is enabled, wrap the core with our sampling configuration
	if cfg.Sampling {
		core := baseLogger.Core()
		sampledCore := createSamplingCore(core, cfg)
		baseLogger = zap.New(sampledCore)
	}

	// Add global fields
	logger := baseLogger.With(
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
// This function handles various sync-related errors that can occur across different
// platforms and environments (especially in CI/CD pipelines), including:
// - "invalid argument" errors when syncing stdout/stderr
// - "inappropriate ioctl for device" on some Unix systems
// - "bad file descriptor" errors during shutdown
// - General EINVAL errors from syscall operations
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
