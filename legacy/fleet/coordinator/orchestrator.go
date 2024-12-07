package coordinator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/fleet/types"
)

// MetalClient defines the interface for interacting with the metal layer
type MetalClient interface {
	ExecuteOperation(ctx context.Context, deviceID types.DeviceID, operation string) error
	GetDeviceMetrics(ctx context.Context, deviceID types.DeviceID) (*types.DeviceMetrics, error)
}

// Orchestrator coordinates fleet-wide operations
type Orchestrator struct {
	scheduler    *Scheduler
	stateManager types.StateManager
	metalClient  MetalClient
	mu           sync.RWMutex
}

// NewOrchestrator creates a new orchestrator instance
func NewOrchestrator(scheduler *Scheduler, stateManager types.StateManager, metalClient MetalClient) *Orchestrator {
	return &Orchestrator{
		scheduler:    scheduler,
		stateManager: stateManager,
		metalClient:  metalClient,
	}
}

// ExecuteTask executes a scheduled task across devices
func (o *Orchestrator) ExecuteTask(ctx context.Context, task types.Task) error {
	// Start task execution
	if err := o.scheduler.StartTask(ctx, task.ID); err != nil {
		return fmt.Errorf("failed to start task: %w", err)
	}

	// Execute on all devices
	var executeError error
	for _, deviceID := range task.DeviceIDs {
		// Get current device state
		state, err := o.stateManager.GetDeviceState(ctx, deviceID)
		if err != nil {
			executeError = fmt.Errorf("failed to get device state: %w", err)
			break
		}
		if state == nil {
			executeError = fmt.Errorf("device not found: %s", deviceID)
			break
		}

		// Execute operation on device
		if err := o.metalClient.ExecuteOperation(ctx, deviceID, task.Operation); err != nil {
			executeError = fmt.Errorf("operation failed on device %s: %w", deviceID, err)
			break
		}

		// Update device metrics
		metrics, err := o.metalClient.GetDeviceMetrics(ctx, deviceID)
		if err != nil {
			executeError = fmt.Errorf("failed to get device metrics: %w", err)
			break
		}

		// Update device state
		state.Metrics = *metrics
		state.LastUpdated = time.Now()
		if err := o.stateManager.UpdateDeviceState(ctx, *state); err != nil {
			executeError = fmt.Errorf("failed to update device state: %w", err)
			break
		}
	}

	// Complete task
	if err := o.scheduler.CompleteTask(ctx, task.ID, executeError); err != nil {
		return fmt.Errorf("failed to complete task: %w", err)
	}

	return executeError
}

// GetDeviceState retrieves current device state
func (o *Orchestrator) GetDeviceState(ctx context.Context, deviceID types.DeviceID) (*types.DeviceState, error) {
	return o.stateManager.GetDeviceState(ctx, deviceID)
}

// UpdateDeviceState updates device state
func (o *Orchestrator) UpdateDeviceState(ctx context.Context, state types.DeviceState) error {
	return o.stateManager.UpdateDeviceState(ctx, state)
}

// ListDevices returns all known devices
func (o *Orchestrator) ListDevices(ctx context.Context) ([]types.DeviceState, error) {
	return o.stateManager.ListDevices(ctx)
}