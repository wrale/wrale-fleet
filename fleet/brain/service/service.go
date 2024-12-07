package service

import (
    "context"

    "github.com/wrale/wrale-fleet/fleet/brain/coordinator"
    "github.com/wrale/wrale-fleet/fleet/brain/device"
    "github.com/wrale/wrale-fleet/fleet/brain/engine"
    "github.com/wrale/wrale-fleet/fleet/brain/types"
)

// Service provides the main brain service functionality
type Service struct {
    inventory   *device.Inventory
    topology    *device.TopologyManager
    scheduler   *coordinator.Scheduler
    orchestrator *coordinator.Orchestrator
    analyzer    *engine.Analyzer
    optimizer   *engine.Optimizer
}

// NewService creates a new brain service instance
func NewService(metalClient coordinator.MetalClient) *Service {
    // Initialize components
    inventory := device.NewInventory()
    topology := device.NewTopologyManager(inventory)
    scheduler := coordinator.NewScheduler()
    orchestrator := coordinator.NewOrchestrator(scheduler, inventory, metalClient)
    analyzer := engine.NewAnalyzer(inventory, topology)
    optimizer := engine.NewOptimizer(inventory, topology, analyzer)

    return &Service{
        inventory:   inventory,
        topology:    topology,
        scheduler:   scheduler,
        orchestrator: orchestrator,
        analyzer:    analyzer,
        optimizer:   optimizer,
    }
}

// Device Management

func (s *Service) RegisterDevice(ctx context.Context, state types.DeviceState) error {
    return s.inventory.RegisterDevice(ctx, state)
}

func (s *Service) UnregisterDevice(ctx context.Context, deviceID types.DeviceID) error {
    return s.inventory.UnregisterDevice(ctx, deviceID)
}

func (s *Service) UpdateDeviceState(ctx context.Context, state types.DeviceState) error {
    return s.inventory.UpdateState(ctx, state)
}

func (s *Service) GetDeviceState(ctx context.Context, deviceID types.DeviceID) (*types.DeviceState, error) {
    return s.inventory.GetDevice(ctx, deviceID)
}

func (s *Service) ListDevices(ctx context.Context) ([]types.DeviceState, error) {
    return s.inventory.ListDevices(ctx)
}

// Task Management

func (s *Service) ScheduleTask(ctx context.Context, task types.Task) error {
    return s.scheduler.Schedule(ctx, task)
}

func (s *Service) CancelTask(ctx context.Context, taskID types.TaskID) error {
    return s.scheduler.Cancel(ctx, taskID)
}

func (s *Service) GetTask(ctx context.Context, taskID types.TaskID) (*coordinator.TaskEntry, error) {
    return s.scheduler.GetTask(ctx, taskID)
}

func (s *Service) ListTasks(ctx context.Context) ([]coordinator.TaskEntry, error) {
    return s.scheduler.ListTasks(ctx)
}

func (s *Service) ExecuteTask(ctx context.Context, task types.Task) error {
    return s.orchestrator.ExecuteTask(ctx, task)
}

// Analysis and Optimization

func (s *Service) AnalyzeFleet(ctx context.Context) (*types.FleetAnalysis, error) {
    return s.analyzer.AnalyzeState(ctx)
}

func (s *Service) GetAlerts(ctx context.Context) ([]types.Alert, error) {
    return s.analyzer.GetAlerts(ctx)
}

func (s *Service) GetRecommendations(ctx context.Context) ([]types.Recommendation, error) {
    return s.analyzer.GetRecommendations(ctx)
}

func (s *Service) OptimizeResources(ctx context.Context, devices []types.DeviceState) ([]types.DeviceState, error) {
    return s.optimizer.OptimizeResources(ctx, devices)
}

func (s *Service) SuggestPlacements(ctx context.Context, task types.Task) ([]types.DeviceID, error) {
    return s.optimizer.SuggestPlacements(ctx, task)
}

// Physical Management

func (s *Service) RegisterRack(ctx context.Context, rackID string, config device.RackConfig) error {
    return s.topology.RegisterRack(ctx, rackID, config)
}

func (s *Service) UnregisterRack(ctx context.Context, rackID string) error {
    return s.topology.UnregisterRack(ctx, rackID)
}

func (s *Service) UpdateDeviceLocation(ctx context.Context, deviceID types.DeviceID, location types.PhysicalLocation) error {
    return s.topology.UpdateLocation(ctx, deviceID, location)
}

func (s *Service) GetDeviceLocation(ctx context.Context, deviceID types.DeviceID) (*types.PhysicalLocation, error) {
    return s.topology.GetLocation(ctx, deviceID)
}

func (s *Service) GetDevicesInZone(ctx context.Context, zone string) ([]types.DeviceState, error) {
    return s.topology.GetDevicesInZone(ctx, zone)
}

func (s *Service) GetDevicesInRack(ctx context.Context, rack string) ([]types.DeviceState, error) {
    return s.topology.GetDevicesInRack(ctx, rack)
}
