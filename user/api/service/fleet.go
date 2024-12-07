package service

import (
    "context"
    "fmt"
    "log"
    "google.golang.org/grpc"

    "github.com/wrale/wrale-fleet/fleet/brain/service"
    "github.com/wrale/wrale-fleet/fleet/brain/types"
    apitypes "github.com/wrale/wrale-fleet/user/api/types"
)

// FleetService implements fleet-wide operations
type FleetService struct {
    brainSvc *service.Service
    conn     *grpc.ClientConn
}

// NewFleetService creates a new fleet service
func NewFleetService(fleetEndpoint string) *FleetService {
    // Connect to fleet brain service
    conn, err := grpc.Dial(fleetEndpoint, grpc.WithInsecure())
    if err != nil {
        log.Fatalf("Failed to connect to fleet brain: %v", err)
    }

    return &FleetService{
        brainSvc: service.NewClient(conn),
        conn: conn,
    }
}

// Close releases resources
func (s *FleetService) Close() error {
    if s.conn != nil {
        return s.conn.Close()
    }
    return nil
}

// GetFleetMetrics returns system-wide metrics
func (s *FleetService) GetFleetMetrics() (*apitypes.FleetMetrics, error) {
    ctx := context.Background()

    // Get fleet-wide metrics from brain
    metrics, err := s.brainSvc.GetFleetMetrics(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get fleet metrics: %w", err)
    }

    return &apitypes.FleetMetrics{
        TotalDevices:   metrics.TotalDevices,
        ActiveDevices:  metrics.ActiveDevices,
        CPUUsage:       metrics.CPUUsage,
        MemoryUsage:    metrics.MemoryUsage,
        PowerUsage:     metrics.PowerUsage,
        AverageLatency: metrics.AverageLatency,
    }, nil
}

// ExecuteFleetCommand executes a fleet-wide operation
func (s *FleetService) ExecuteFleetCommand(req *apitypes.FleetCommandRequest) error {
    ctx := context.Background()

    // Create fleet task
    task := types.Task{
        Operation: req.Operation,
        Priority:  1, // Fleet operations get high priority
    }

    if req.DeviceSelector != nil {
        // Get matching device IDs
        devices, err := s.brainSvc.QueryDevices(ctx, *req.DeviceSelector)
        if err != nil {
            return fmt.Errorf("failed to query devices: %w", err)
        }
        for _, device := range devices {
            task.DeviceIDs = append(task.DeviceIDs, device.ID)
        }
    }

    // Execute fleet-wide task
    if err := s.brainSvc.ExecuteFleetTask(ctx, task); err != nil {
        return fmt.Errorf("failed to execute fleet command: %w", err)
    }

    return nil
}

// GetFleetConfig returns fleet-wide configuration
func (s *FleetService) GetFleetConfig() (map[string]interface{}, error) {
    ctx := context.Background()
    return s.brainSvc.GetFleetConfig(ctx)
}

// UpdateFleetConfig updates fleet-wide configuration
func (s *FleetService) UpdateFleetConfig(config map[string]interface{}) error {
    ctx := context.Background()
    return s.brainSvc.UpdateFleetConfig(ctx, config)
}