// Package logger provides logging configuration for the wfcentral command.
package logger

import (
	"fmt"

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
func Sync(logger *zap.Logger) error {
	err := logger.Sync()
	if err != nil {
		// Logger sync errors are expected on some platforms
		// Just log to stderr but don't fail
		fmt.Printf("logger sync warning: %v\n", err)
	}
	return nil
}
