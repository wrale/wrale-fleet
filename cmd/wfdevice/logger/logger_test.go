package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"
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
	// Create an observed logger to capture log output
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	tests := []struct {
		name          string
		currentStage  int
		requiredStage int
		operation     string
		want          bool
		wantWarning   bool // Should we expect a warning log?
		wantError     bool // Should we expect an error log?
	}{
		{
			name:          "supported operation",
			currentStage:  2,
			requiredStage: 1,
			operation:     "basic_op",
			want:          true,
			wantWarning:   false,
			wantError:     false,
		},
		{
			name:          "unsupported operation",
			currentStage:  1,
			requiredStage: 2,
			operation:     "advanced_op",
			want:          false,
			wantWarning:   true,
			wantError:     false,
		},
		{
			name:          "equal stages",
			currentStage:  3,
			requiredStage: 3,
			operation:     "current_op",
			want:          false, // With new implementation, Stage 2+ requires explicit stage setting
			wantWarning:   true,
			wantError:     false,
		},
		{
			name:          "invalid required stage",
			currentStage:  1,
			requiredStage: MaxStage + 1,
			operation:     "invalid_op",
			want:          false,
			wantWarning:   false,
			wantError:     true,
		},
		{
			name:          "stage 1 operation always allowed",
			currentStage:  1,
			requiredStage: 1,
			operation:     "basic_op",
			want:          true,
			wantWarning:   false,
			wantError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs.TakeAll() // Clear previous logs

			// Create logger with stage if needed
			testLogger := WithStage(logger, tt.currentStage)

			// Check operation support
			got := StageCheck(testLogger, tt.requiredStage, tt.operation)
			assert.Equal(t, tt.want, got)

			// Verify logging behavior
			logEntries := logs.TakeAll()
			hasWarning := false
			hasError := false
			for _, entry := range logEntries {
				switch entry.Level {
				case zapcore.WarnLevel:
					hasWarning = true
					assert.Contains(t, entry.Message, "operation requires higher stage capability")
				case zapcore.ErrorLevel:
					hasError = true
					assert.Contains(t, entry.Message, "invalid required stage")
				}
			}
			assert.Equal(t, tt.wantWarning, hasWarning, "warning log presence")
			assert.Equal(t, tt.wantError, hasError, "error log presence")
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
	// Create an observed logger to capture log output
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

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
			logs.TakeAll() // Clear previous logs

			// Create logger with stage
			stagedLogger := WithStage(logger, tt.inputStage)

			// Generate a log entry to verify stage field
			stagedLogger.Info("test message")

			// Verify the stage field in the log entry
			entries := logs.TakeAll()
			require.Len(t, entries, 1, "should have one log entry")

			// Check if stage field was set correctly
			stageField, ok := entries[0].ContextMap()[stageKey]
			require.True(t, ok, "stage field should be present")
			assert.Equal(t, tt.wantStage, stageField)
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
