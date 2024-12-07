package service

import (
    "context"
    "fmt"
    "time"

    "github.com/wrale/wrale-fleet/fleet/brain/service"
    "github.com/wrale/wrale-fleet/fleet/brain/types"
    apitypes "github.com/wrale/wrale-fleet/user/api/types"
)

// FleetService implements fleet-wide operations
type FleetService struct {
    brainSvc *service.Service
}

// NewFleetService creates a new fleet service
func NewFleetService(brainSvc *service.Service) *FleetService {
    return &FleetService{
        brainSvc: brainSvc,
    }
}

// ExecuteFleetCommand executes an operation across multiple devices
func (s *FleetService) ExecuteFleetCommand(req *apitypes.FleetCommandRequest) (*apitypes.CommandResponse, error) {
    ctx := context.Background()

    // Create fleet task
    task := types.Task{
        ID:        fmt.Sprintf("fleet-%d", time.Now().UnixNano()),
        DeviceIDs: req.Devices,
        Operation: req.Operation,
        Priority:  1,
        CreatedAt: time.Now(),
    }

    // Schedule task
    if err := s.brainSvc.ScheduleTask(ctx, task); err != nil {
        return nil, fmt.Errorf("failed to schedule fleet task: %w", err)
    }

    // Execute task
    if err := s.brainSvc.ExecuteTask(ctx, task); err != nil {
        return nil, fmt.Errorf("failed to execute fleet task: %w", err)
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

// GetFleetMetrics returns aggregated fleet metrics
func (s *FleetService) GetFleetMetrics() (map[string]interface{}, error) {
    ctx := context.Background()

    // Get all devices
    devices, err := s.brainSvc.ListDevices(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to list devices: %w", err)
    }

    metrics := map[string]interface{}{
        "total_devices":   len(devices),
        "active_devices": 0,
        "total_power":    0.0,
        "avg_temp":       0.0,
        "avg_cpu":        0.0,
        "avg_memory":     0.0,
    }

    if len(devices) > 0 {
        var totalTemp, totalCPU, totalMem, totalPower float64
        for _, device := range devices {
            if device.Status == "active" {
                metrics["active_devices"] = metrics["active_devices"].(int) + 1
            }
            totalTemp += device.Metrics.Temperature
            totalCPU += device.Metrics.CPULoad
            totalMem += device.Metrics.MemoryUsage
            totalPower += device.Metrics.PowerUsage
        }

        count := float64(len(devices))
        metrics["avg_temp"] = totalTemp / count
        metrics["avg_cpu"] = totalCPU / count
        metrics["avg_memory"] = totalMem / count
        metrics["total_power"] = totalPower
    }

    // Get fleet analysis
    analysis, err := s.brainSvc.AnalyzeFleet(ctx)
    if err == nil { // Add analysis metrics if available
        metrics["resource_usage"] = analysis.ResourceUsage
        metrics["healthy_devices"] = analysis.HealthyDevices
        metrics["alert_count"] = len(analysis.Alerts)
        metrics["recommendation_count"] = len(analysis.Recommendations)
    }

    return metrics, nil
}

// UpdateConfig updates configuration for specified devices
func (s *FleetService) UpdateConfig(req *apitypes.ConfigUpdateRequest) error {
    ctx := context.Background()

    // Handle all devices if none specified
    deviceIDs := req.Devices
    if len(deviceIDs) == 0 {
        devices, err := s.brainSvc.ListDevices(ctx)
        if err != nil {
            return fmt.Errorf("failed to list devices: %w", err)
        }
        for _, device := range devices {
            deviceIDs = append(deviceIDs, device.ID)
        }
    }

    // Update each device's configuration
    for _, deviceID := range deviceIDs {
        if err := s.brainSvc.UpdateDeviceConfig(ctx, deviceID, req.Config); err != nil {
            return fmt.Errorf("failed to update config for device %s: %w", deviceID, err)
        }
    }

    return nil
}

// GetConfig retrieves configurations for specified devices
func (s *FleetService) GetConfig(deviceIDs []types.DeviceID) (map[types.DeviceID]map[string]interface{}, error) {
    ctx := context.Background()
    configs := make(map[types.DeviceID]map[string]interface{})

    // Get all devices if none specified
    if len(deviceIDs) == 0 {
        devices, err := s.brainSvc.ListDevices(ctx)
        if err != nil {
            return nil, fmt.Errorf("failed to list devices: %w", err)
        }
        for _, device := range devices {
            deviceIDs = append(deviceIDs, device.ID)
        }
    }

    // Get config for each device
    for _, deviceID := range deviceIDs {
        config, err := s.brainSvc.GetDeviceConfig(ctx, deviceID)
        if err != nil {
            return nil, fmt.Errorf("failed to get config for device %s: %w", deviceID, err)
        }
        configs[deviceID] = config
    }

    return configs, nil
}
