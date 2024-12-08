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
	return logger.With(zap.Int("stage", stage))
}

// StageCheck verifies if a requested operation is supported in the current stage.
// It logs appropriate warnings for unsupported operations and returns false if
// the operation is not supported in the current stage.
func StageCheck(logger *zap.Logger, requiredStage int, operation string) bool {
	currentStage := MinStage // Default to Stage 1 if not specified

	// Extract current stage from logger context if available
	if stage := logger.Check(zapcore.InfoLevel, ""); stage != nil {
		if stageField := stage.Entry.ContextMap()["stage"]; stageField != nil {
			if s, ok := stageField.(int); ok {
				currentStage = s
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
	return zap.Int("stage", stage)
}
