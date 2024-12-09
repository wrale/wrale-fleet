package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

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
			want:          false,
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

func TestWithStage(t *testing.T) {
	// Create an observed logger to capture log output
	core, logs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	tests := []struct {
		name       string
		inputStage int
		wantStage  int64 // Changed to int64 to match zap's internal representation
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

func TestGetStage(t *testing.T) {
	logger := zap.NewExample()

	tests := []struct {
		name       string
		inputStage int
		want       int
	}{
		{
			name:       "normal stage",
			inputStage: 3,
			want:       3,
		},
		{
			name:       "no stage set",
			inputStage: 0,
			want:       MinStage,
		},
		{
			name:       "maximum stage",
			inputStage: MaxStage,
			want:       MaxStage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.inputStage > 0 {
				logger = WithStage(logger, tt.inputStage)
			}
			got := GetStage(logger)
			assert.Equal(t, tt.want, got)
		})
	}
}
