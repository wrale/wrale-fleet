// Package logger provides a stage-aware logging infrastructure for the wfdevice command.
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
// and proper capability gating. The stage value is constrained to be between
// MinStage and MaxStage inclusive.
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
//
// The current stage is determined by the stage field in the logger's context.
// If no stage is explicitly set, MinStage (1) is assumed.
func StageCheck(logger *zap.Logger, requiredStage int, operation string) bool {
	// The stage should already be set in the logger's fields during creation
	// or via WithStage(). We keep using the same logger to maintain the stage.

	if requiredStage > MaxStage {
		logger.Error("invalid required stage",
			zap.String("operation", operation),
			zap.Int("required_stage", requiredStage),
			zap.Int("max_stage", MaxStage),
		)
		return false
	}

	// No need to check stages for operations supported in Stage 1
	if requiredStage <= MinStage {
		return true
	}

	// For operations requiring Stage 2+, warn if attempted at a lower stage
	logger.Warn("operation requires higher stage capability",
		zap.String("operation", operation),
		zap.Int("required_stage", requiredStage),
	)
	return false
}

// StageField adds stage information as a structured field.
// This is useful when adding stage context to individual log entries.
// The stage value is constrained to be between MinStage and MaxStage inclusive.
func StageField(stage int) zap.Field {
	if stage < MinStage {
		stage = MinStage
	}
	if stage > MaxStage {
		stage = MaxStage
	}
	return zap.Int(stageKey, stage)
}
