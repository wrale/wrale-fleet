package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// MinStage is the minimum supported stage (Stage 1)
	MinStage = 1

	// MaxStage is the maximum supported stage (Stage 6)
	MaxStage = 6

	// stageKey is the key used to store stage information in the logger
	stageKey = "stage"
)

// WithStage adds stage information to a logger, enabling stage-aware logging
// and proper capability gating.
func WithStage(logger *zap.Logger, stage int) *zap.Logger {
	if stage < MinStage {
		stage = MinStage // Ensure minimum of Stage 1
	}
	if stage > MaxStage {
		stage = MaxStage // Cap at maximum Stage 6
	}
	return logger.With(zap.Int(stageKey, stage))
}

// StageCheck verifies if a requested operation is supported in the current stage.
// It logs appropriate warnings for unsupported operations and returns false if
// the operation is not supported in the current stage.
func StageCheck(logger *zap.Logger, requiredStage int, operation string) bool {
	currentStage := MinStage // Default to Stage 1 if not specified

	// Extract current stage from logger context
	if ce := logger.Check(zapcore.InfoLevel, ""); ce != nil {
		for _, f := range ce.Context {
			if f.Key == stageKey {
				if stage, ok := f.Integer; ok {
					currentStage = int(stage)
					break
				}
			}
		}
	}

	// Check if operation is supported in current stage
	if currentStage < requiredStage {
		logger.Warn("operation requires higher stage capability",
			zap.String("operation", operation),
			zap.Int("required_stage", requiredStage),
			zap.Int("current_stage", currentStage),
		)
		return false
	}

	return true
}

// StageField adds stage information as a structured field.
// This is useful when adding stage context to individual log entries.
func StageField(stage int) zap.Field {
	if stage < MinStage {
		stage = MinStage
	}
	if stage > MaxStage {
		stage = MaxStage
	}
	return zap.Int(stageKey, stage)
}