package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

func TestLoggerCreation(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		logLevel    string
		wantLevel   zapcore.Level
		wantJSON    bool
	}{
		{
			name:        "development defaults",
			environment: "development",
			wantLevel:   zapcore.DebugLevel,
			wantJSON:    false,
		},
		{
			name:        "production defaults",
			environment: "production",
			wantLevel:   zapcore.InfoLevel,
			wantJSON:    true,
		},
		{
			name:        "custom level",
			environment: "development",
			logLevel:    "error",
			wantLevel:   zapcore.ErrorLevel,
			wantJSON:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save environment
			prevEnv := os.Getenv("ENVIRONMENT")
			prevLevel := os.Getenv("LOG_LEVEL")
			defer func() {
				os.Setenv("ENVIRONMENT", prevEnv)
				os.Setenv("LOG_LEVEL", prevLevel)
			}()

			// Set test environment
			os.Setenv("ENVIRONMENT", tt.environment)
			if tt.logLevel != "" {
				os.Setenv("LOG_LEVEL", tt.logLevel)
			} else {
				os.Unsetenv("LOG_LEVEL")
			}

			// Create logger
			logger, err := New()
			require.NoError(t, err)
			defer func() {
				assert.NoError(t, Sync(logger))
			}()

			// Verify configuration
			assert.Equal(t, tt.wantLevel, getLoggerLevel(logger))
		})
	}
}

func TestStageCheck(t *testing.T) {
	logger := zap.NewExample()

	tests := []struct {
		name          string
		currentStage  int
		requiredStage int
		operation     string
		want          bool
	}{
		{
			name:          "supported operation",
			currentStage:  2,
			requiredStage: 1,
			operation:     "basic_op",
			want:          true,
		},
		{
			name:          "unsupported operation",
			currentStage:  1,
			requiredStage: 2,
			operation:     "advanced_op",
			want:          false,
		},
		{
			name:          "equal stages",
			currentStage:  3,
			requiredStage: 3,
			operation:     "current_op",
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create logger with stage
			stagedLogger := WithStage(logger, tt.currentStage)

			// Check operation support
			got := StageCheck(stagedLogger, tt.requiredStage, tt.operation)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLoggerSync(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Test normal sync
	assert.NoError(t, Sync(logger))

	// Test sync with nil logger
	assert.NoError(t, Sync(nil))
}

func TestWithStage(t *testing.T) {
	logger := zap.NewExample()

	tests := []struct {
		name       string
		inputStage int
		wantStage  int
	}{
		{
			name:       "normal stage",
			inputStage: 3,
			wantStage:  3,
		},
		{
			name:       "below minimum",
			inputStage: 0,
			wantStage:  MinStage,
		},
		{
			name:       "above maximum",
			inputStage: MaxStage + 1,
			wantStage:  MaxStage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stagedLogger := WithStage(logger, tt.inputStage)
			stage := extractStage(t, stagedLogger)
			assert.Equal(t, tt.wantStage, stage)
		})
	}
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

// extractStage gets the stage value from a logger's context
func extractStage(t *testing.T, logger *zap.Logger) int {
	if stage := logger.Check(zapcore.InfoLevel, ""); stage != nil {
		if stageField := stage.Entry.ContextMap()["stage"]; stageField != nil {
			if s, ok := stageField.(int); ok {
				return s
			}
		}
	}
	t.Fatal("Failed to extract stage from logger")
	return 0
}
