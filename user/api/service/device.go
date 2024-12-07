// Package service implements concrete API services
package service

import (
    "context"
    "fmt"
    "time"

    "github.com/wrale/wrale-fleet/fleet/brain/service"
    "github.com/wrale/wrale-fleet/fleet/brain/coordinator"
    "github.com/wrale/wrale-fleet/fleet/brain/types"
    apitypes "github.com/wrale/wrale-fleet/user/api/types"
)

// DeviceService implements device operations
type DeviceService struct {
    brain *service.Service
}

// NewDeviceService creates a new device service
func NewDeviceService(brain *service.Service) *DeviceService {
    return &DeviceService{
        brain: brain,
    }
}

// CreateDevice registers a new device
func (s *DeviceService) CreateDevice(req *apitypes.DeviceCreateRequest) (*apitypes.DeviceResponse, error) {
    ctx := context.Background()

    // Create device state
    state := types.DeviceState{
        ID:       req.ID,
        Status:   "initializing",
        Location: req.Location,
        Resources: map[types.ResourceType]float64{
            types.ResourceCPU:    100.0,
            types.ResourceMemory: 100.0,
        },
    }

    // Register with brain
    if err := s.brain.RegisterDevice(ctx, state); err != nil {
        return nil, fmt.Errorf("failed to register device: %w", err)
    }

    // Get device response which includes full state
    return s.getDeviceResponse(ctx, req.ID)
}

// GetDevice retrieves device information
func (s *DeviceService) GetDevice(id types.DeviceID) (*apitypes.DeviceResponse, error) {
    return s.getDeviceResponse(context.Background(), id)
}

// UpdateDevice updates device state
func (s *DeviceService) UpdateDevice(id types.DeviceID, req *apitypes.DeviceUpdateRequest) (*apitypes.DeviceResponse, error) {
    ctx := context.Background()

    // Get current state
    state, err := s.brain.GetDeviceState(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get device state: %w", err)
    }

    // Apply updates
    if req.Status != "" {
        state.Status = req.Status
    }
    if req.Location != nil {
        state.Location = *req.Location
    }

    // Update state
    if err := s.brain.UpdateDeviceState(ctx, *state); err != nil {
        return nil, fmt.Errorf("failed to update device state: %w", err)
    }

    // Get updated device state
    return s.getDeviceResponse(ctx, id)
}

// ListDevices returns all registered devices
func (s *DeviceService) ListDevices() ([]*apitypes.DeviceResponse, error) {
    ctx := context.Background()

    // Get all devices from brain
    states, err := s.brain.ListDevices(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to list devices: %w", err)
    }

    // Convert to API responses
    responses := make([]*apitypes.DeviceResponse, len(states))
    for i, state := range states {
        resp, err := s.deviceStateToResponse(ctx, state)
        if err != nil {
            return nil, err
        }
        responses[i] = resp
    }

    return responses, nil
}

// DeleteDevice unregisters a device
func (s *DeviceService) DeleteDevice(id types.DeviceID) error {
    ctx := context.Background()

    if err := s.brain.UnregisterDevice(ctx, id); err != nil {
        return fmt.Errorf("failed to unregister device: %w", err)
    }

    return nil
}

// ExecuteCommand executes a device operation
func (s *DeviceService) ExecuteCommand(id types.DeviceID, req *apitypes.DeviceCommandRequest) (*apitypes.CommandResponse, error) {
    ctx := context.Background()

    // Create task
    taskID := types.TaskID(fmt.Sprintf("cmd-%d", time.Now().UnixNano()))
    task := types.Task{
        ID:        taskID,
        Type:      types.TaskType(req.Operation),
        DeviceIDs: []types.DeviceID{id},
        Operation: req.Operation,
        Priority:  1,
        CreatedAt: time.Now(),
        Payload:   req.Payload,
    }

    // Schedule and execute task
    if err := s.brain.ScheduleTask(ctx, task); err != nil {
        return nil, fmt.Errorf("failed to schedule task: %w", err)
    }

    if err := s.brain.ExecuteTask(ctx, task); err != nil {
        return nil, fmt.Errorf("failed to execute task: %w", err)
    }

    // Get result task
    entry, err := s.brain.GetTask(ctx, taskID)
    if err != nil {
        return nil, fmt.Errorf("failed to get task result: %w", err)
    }

    // Convert to API response
    resp := &apitypes.CommandResponse{
        ID:        taskID,
        Status:    string(entry.State),
    }
    if entry.StartedAt != nil {
        resp.StartTime = *entry.StartedAt
    }
    if entry.EndedAt != nil {
        resp.EndTime = entry.EndedAt
    }
    if entry.Error != nil {
        resp.Error = entry.Error.Error()
    }

    return resp, nil
}

// Helper methods

func (s *DeviceService) getDeviceResponse(ctx context.Context, id types.DeviceID) (*apitypes.DeviceResponse, error) {
    // Get device state
    state, err := s.brain.GetDeviceState(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get device state: %w", err)
    }

    return s.deviceStateToResponse(ctx, *state)
}

func (s *DeviceService) deviceStateToResponse(ctx context.Context, state types.DeviceState) (*apitypes.DeviceResponse, error) {
    return &apitypes.DeviceResponse{
        ID:         state.ID,
        Status:     state.Status,
        Location:   state.Location,
        Metrics:    &state.Metrics,
        Config:     nil, // TODO: Implement config retrieval
        LastUpdate: state.LastUpdated,
    }, nil
}