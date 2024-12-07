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
    brain *service.Service
}

// NewFleetService creates a new fleet service
func NewFleetService(brain *service.Service) *FleetService {
    return &FleetService{
        brain: brain,
    }
}

// GetFleetMetrics returns fleet-wide metrics
func (s *FleetService) GetFleetMetrics() (*apitypes.FleetMetrics, error) {
    ctx := context.Background()

    // Get analysis from brain
    analysis, err := s.brain.AnalyzeFleet(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get fleet metrics: %w", err)
    }

    // Convert to API metrics
    metrics := &apitypes.FleetMetrics{
        TotalDevices:  analysis.TotalDevices,
        ActiveDevices: analysis.HealthyDevices,
    }

    // Extract resource usage
    if usage, ok := analysis.ResourceUsage[types.ResourceCPU]; ok {
        metrics.CPUUsage = usage
    }
    if usage, ok := analysis.ResourceUsage[types.ResourceMemory]; ok {
        metrics.MemoryUsage = usage
    }
    if usage, ok := analysis.ResourceUsage[types.ResourcePower]; ok {
        metrics.PowerUsage = usage
    }

    return metrics, nil
}

// ExecuteFleetCommand executes a fleet-wide operation
func (s *FleetService) ExecuteFleetCommand(req *apitypes.FleetCommandRequest) error {
    ctx := context.Background()

    // Create task with no specific devices
    task := types.Task{
        ID:        types.TaskID(fmt.Sprintf("fleet-%d", time.Now().UnixNano())),
        Operation: req.Operation,
        Priority:  1,
        CreatedAt: time.Now(),
    }

    // If device selector is provided, get matching devices
    if req.DeviceSelector != nil {
        // Extract devices matching location (only location supported for now)
        if req.DeviceSelector.Location != "" {
            devices, err := s.brain.GetDevicesInZone(ctx, req.DeviceSelector.Location)
            if err != nil {
                return fmt.Errorf("failed to get devices in zone: %w", err)
            }
            for _, device := range devices {
                task.DeviceIDs = append(task.DeviceIDs, device.ID)
            }
        }
    }

    // Schedule and execute task
    if err := s.brain.ScheduleTask(ctx, task); err != nil {
        return fmt.Errorf("failed to schedule fleet task: %w", err)
    }

    if err := s.brain.ExecuteTask(ctx, task); err != nil {
        return fmt.Errorf("failed to execute fleet task: %w", err)
    }

    return nil
}

// GetFleetConfig gets fleet-wide configuration
func (s *FleetService) GetFleetConfig() (map[string]interface{}, error) {
    // TODO: Implement fleet-wide configuration in brain service
    return map[string]interface{}{}, nil
}

// UpdateFleetConfig updates fleet-wide configuration
func (s *FleetService) UpdateFleetConfig(config map[string]interface{}) error {
    // TODO: Implement fleet-wide configuration in brain service
    return nil
}