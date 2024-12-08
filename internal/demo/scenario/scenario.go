// Package scenario provides the core types and interfaces for demo scenarios
package scenario

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Scenario represents a single demonstration scenario
type Scenario interface {
	// Name returns the scenario's identifier
	Name() string

	// Description returns a human-readable description of what the scenario demonstrates
	Description() string

	// Setup prepares any necessary resources for the scenario
	Setup(ctx context.Context) error

	// Run executes the scenario's demonstration
	Run(ctx context.Context) error

	// Cleanup releases any resources created during setup
	Cleanup(ctx context.Context) error
}

// BaseScenario provides a common implementation of the Scenario interface
type BaseScenario struct {
	name        string
	description string
	logger      *zap.Logger
}

// NewBaseScenario creates a new base scenario with the given name and description
func NewBaseScenario(name, description string, logger *zap.Logger) *BaseScenario {
	return &BaseScenario{
		name:        name,
		description: description,
		logger:      logger,
	}
}

// Name returns the scenario's name
func (b *BaseScenario) Name() string {
	return b.name
}

// Description returns the scenario's description
func (b *BaseScenario) Description() string {
	return b.description
}

// Setup provides a default no-op implementation
func (b *BaseScenario) Setup(ctx context.Context) error {
	return nil
}

// Run must be implemented by concrete scenarios
func (b *BaseScenario) Run(ctx context.Context) error {
	return fmt.Errorf("Run not implemented for scenario %s", b.name)
}

// Cleanup provides a default no-op implementation
func (b *BaseScenario) Cleanup(ctx context.Context) error {
	return nil
}

// Runner executes a series of scenarios while providing progress feedback
type Runner struct {
	logger    *zap.Logger
	scenarios []Scenario
}

// NewRunner creates a new scenario runner
func NewRunner(logger *zap.Logger, scenarios []Scenario) *Runner {
	return &Runner{
		logger:    logger,
		scenarios: scenarios,
	}
}

// Run executes all scenarios in sequence
func (r *Runner) Run(ctx context.Context) error {
	for i, scenario := range r.scenarios {
		r.logger.Info("starting scenario",
			zap.String("name", scenario.Name()),
			zap.Int("number", i+1),
			zap.Int("total", len(r.scenarios)))

		start := time.Now()

		// Setup phase
		r.logger.Info("setting up scenario", zap.String("name", scenario.Name()))
		if err := scenario.Setup(ctx); err != nil {
			return fmt.Errorf("scenario setup failed: %w", err)
		}

		// Execution phase
		r.logger.Info("running scenario", zap.String("name", scenario.Name()))
		if err := scenario.Run(ctx); err != nil {
			return fmt.Errorf("scenario execution failed: %w", err)
		}

		// Cleanup phase
		r.logger.Info("cleaning up scenario", zap.String("name", scenario.Name()))
		if err := scenario.Cleanup(ctx); err != nil {
			// Log cleanup error but don't fail - we want to attempt cleanup of other scenarios
			r.logger.Error("scenario cleanup failed",
				zap.String("name", scenario.Name()),
				zap.Error(err))
		}

		duration := time.Since(start)
		r.logger.Info("completed scenario",
			zap.String("name", scenario.Name()),
			zap.Duration("duration", duration))
	}

	return nil
}
