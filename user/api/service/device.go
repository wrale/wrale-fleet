// Package service implements concrete API services
package service

import (
    "context"
    "fmt"
    "time"

    "github.com/wrale/wrale-fleet/fleet/brain/service"
    "github.com/wrale/wrale-fleet/fleet/brain/types"
    apitypes "github.com/wrale/wrale-fleet/user/api/types"
)

// DeviceService implements device operations
type DeviceService struct {
    brainSvc *service.Service
}

// NewDeviceService creates a new device service
func NewDeviceService(brainSvc *service.Service) *DeviceService {
    return &DeviceService{
        brainSvc: brainSvc,
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
    if err := s.brainSvc.RegisterDevice(ctx, state); err != nil {
        return nil, fmt.Errorf("failed to register device: %w", err)
    }

    // Update initial config if provided
    if len(req.Config) > 0 {
        if err := s.brainSvc.UpdateDeviceConfig(ctx, req.ID, req.Config); err != nil {
            return nil, fmt.Errorf("failed to set initial config: %w", err)
        }
    }

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
    state, err := s.brainSvc.GetDeviceState(ctx, id)
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
    if err := s.brainSvc.UpdateDeviceState(ctx, id, *state); err != nil {
        return nil, fmt.Errorf("failed to update device state: %w", err)
    }

    // Update config if provided
    if len(req.Config) > 0 {
        if err := s.brainSvc.UpdateDeviceConfig(ctx, id, req.Config); err != nil {
            return nil, fmt.Errorf("failed to update config: %w", err)
        }
    }

    return s.getDeviceResponse(ctx, id)
}

// ListDevices returns all registered devices
func (s *DeviceService) ListDevices() ([]*apitypes.DeviceResponse, error) {
    ctx := context.Background()

    // Get all devices from brain
    states, err := s.brainSvc.ListDevices(ctx)
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

    if err := s.brainSvc.UnregisterDevice(ctx, id); err != nil {
        return fmt.Errorf("failed to unregister device: %w", err)
    }

    return nil
}

// ExecuteCommand executes a device operation
func (s *DeviceService) ExecuteCommand(id types.DeviceID, req *apitypes.DeviceCommandRequest) (*apitypes.CommandResponse, error) {
    ctx := context.Background()

    // Create task
    task := types.Task{
        ID:        fmt.Sprintf("cmd-%d", time.Now().UnixNano()),
        DeviceIDs: []types.DeviceID{id},
        Operation: req.Operation,
        Priority:  1,
        CreatedAt: time.Now(),
    }

    // Schedule task
    if err := s.brainSvc.ScheduleTask(ctx, task); err != nil {
        return nil, fmt.Errorf("failed to schedule task: %w", err)
    }

    // Execute task
    if err := s.brainSvc.ExecuteTask(ctx, task); err != nil {
        return nil, fmt.Errorf("failed to execute task: %w", err)
    }

    // Get task result
    taskEntry, err := s.brainSvc.GetTask(ctx, task.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to get task result: %w", err)
    }

    // Convert to API response
    resp := &apitypes.CommandResponse{
        ID:        task.ID,
        Status:    taskEntry.Status,
        StartTime: taskEntry.StartedAt.Time(),
    }
    if taskEntry.EndedAt != nil {
        endTime := taskEntry.EndedAt.Time()
        resp.EndTime = &endTime
    }
    if taskEntry.Error != nil {
        resp.Error = taskEntry.Error.Error()
    }

    return resp, nil
}

// Helper methods

func (s *DeviceService) getDeviceResponse(ctx context.Context, id types.DeviceID) (*apitypes.DeviceResponse, error) {
    // Get device state
    state, err := s.brainSvc.GetDeviceState(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get device state: %w", err)
    }

    return s.deviceStateToResponse(ctx, *state)
}

func (s *DeviceService) deviceStateToResponse(ctx context.Context, state types.DeviceState) (*apitypes.DeviceResponse, error) {
    // Get device config
    config, err := s.brainSvc.GetDeviceConfig(ctx, state.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to get device config: %w", err)
    }

    return &apitypes.DeviceResponse{
        ID:         state.ID,
        Status:     state.Status,
        Location:   state.Location,
        Metrics:    state.Metrics,
        Config:     config,
        LastUpdate: state.LastUpdated,
    }, nil
}
