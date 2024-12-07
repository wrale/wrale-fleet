package engine

import (
	"context"
	"fmt"
	"sort"

	"github.com/wrale/wrale-fleet/fleet/brain/device"
	"github.com/wrale/wrale-fleet/fleet/brain/types"
)

// Optimizer implements resource optimization logic
type Optimizer struct {
	inventory *device.Inventory
	topology  *device.TopologyManager
	analyzer  *Analyzer
}

// NewOptimizer creates a new optimizer instance
func NewOptimizer(inventory *device.Inventory, topology *device.TopologyManager, analyzer *Analyzer) *Optimizer {
	return &Optimizer{
		inventory: inventory,
		topology:  topology,
		analyzer:  analyzer,
	}
}

// OptimizeResources suggests optimizations for resource allocation
func (o *Optimizer) OptimizeResources(ctx context.Context, devices []types.DeviceState) ([]types.DeviceState, error) {
	// For v1.0, implement basic load balancing
	if len(devices) == 0 {
		return devices, nil
	}

	// Sort devices by CPU load
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].Metrics.CPULoad > devices[j].Metrics.CPULoad
	})

	// Basic load balancing - suggest moving workload from high to low utilization
	optimized := make([]types.DeviceState, len(devices))
	copy(optimized, devices)

	// Find imbalances
	avgCPU := 0.0
	for _, d := range devices {
		avgCPU += d.Metrics.CPULoad
	}
	avgCPU /= float64(len(devices))

	// Adjust suggested resource allocation
	for i := range optimized {
		if optimized[i].Metrics.CPULoad > avgCPU*1.2 { // 20% above average
			optimized[i].Resources[types.ResourceCPU] *= 0.8 // Reduce allocation
		} else if optimized[i].Metrics.CPULoad < avgCPU*0.8 { // 20% below average
			optimized[i].Resources[types.ResourceCPU] *= 1.2 // Increase allocation
		}
	}

	return optimized, nil
}

// GetResourceUtilization returns current resource utilization
func (o *Optimizer) GetResourceUtilization(ctx context.Context) (map[types.ResourceType]float64, error) {
	devices, err := o.inventory.ListDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	utilization := make(map[types.ResourceType]float64)
	if len(devices) == 0 {
		return utilization, nil
	}

	// Calculate average utilization
	for _, device := range devices {
		utilization[types.ResourceCPU] += device.Metrics.CPULoad
		utilization[types.ResourceMemory] += device.Metrics.MemoryUsage
		utilization[types.ResourcePower] += device.Metrics.PowerUsage
	}

	// Convert to averages
	count := float64(len(devices))
	for resource := range utilization {
		utilization[resource] /= count
	}

	return utilization, nil
}

// SuggestPlacements suggests optimal device placements for a task
func (o *Optimizer) SuggestPlacements(ctx context.Context, task types.Task) ([]types.DeviceID, error) {
	devices, err := o.inventory.ListDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	// Filter devices that can handle the task's resource requirements
	suitable := make([]types.DeviceState, 0)
	for _, device := range devices {
		if o.canHandleTask(device, task) {
			suitable = append(suitable, device)
		}
	}

	// Sort by available capacity (inverse of current load)
	sort.Slice(suitable, func(i, j int) bool {
		return suitable[i].Metrics.CPULoad < suitable[j].Metrics.CPULoad
	})

	// Select best candidates
	result := make([]types.DeviceID, 0)
	for _, device := range suitable {
		result = append(result, device.ID)
		if len(result) >= len(task.DeviceIDs) {
			break
		}
	}

	return result, nil
}

// canHandleTask checks if a device can handle the task's requirements
func (o *Optimizer) canHandleTask(device types.DeviceState, task types.Task) bool {
	// Check CPU capacity
	if cpuReq := task.Resources[types.ResourceCPU]; cpuReq > 0 {
		if device.Metrics.CPULoad+cpuReq > 95 { // Leave 5% buffer
			return false
		}
	}

	// Check memory capacity
	if memReq := task.Resources[types.ResourceMemory]; memReq > 0 {
		if device.Metrics.MemoryUsage+memReq > 95 { // Leave 5% buffer
			return false
		}
	}

	// Check power capacity
	if powerReq := task.Resources[types.ResourcePower]; powerReq > 0 {
		if device.Metrics.PowerUsage+powerReq > 800 { // Example threshold
			return false
		}
	}

	return true
}

// OptimizeZone suggests optimizations for devices in a cooling zone
func (o *Optimizer) OptimizeZone(ctx context.Context, zone string) ([]types.Recommendation, error) {
	devices, err := o.topology.GetDevicesInZone(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone devices: %w", err)
	}

	recommendations := make([]types.Recommendation, 0)

	// Check thermal distribution
	var totalTemp float64
	for _, device := range devices {
		totalTemp += device.Metrics.Temperature
	}
	avgTemp := totalTemp / float64(len(devices))

	// Look for thermal hotspots
	for _, device := range devices {
		if device.Metrics.Temperature > avgTemp*1.2 { // 20% above zone average
			recommendations = append(recommendations, types.Recommendation{
				ID:        fmt.Sprintf("thermal-balance-%s", device.ID),
				Priority:  1,
				Action:    "redistribute_workload",
				Reason:    fmt.Sprintf("Device temperature %.1f°C significantly above zone average %.1f°C", device.Metrics.Temperature, avgTemp),
				DeviceIDs: []types.DeviceID{device.ID},
			})
		}
	}

	return recommendations, nil
}

// OptimizePowerDistribution suggests power optimization across racks
func (o *Optimizer) OptimizePowerDistribution(ctx context.Context) ([]types.Recommendation, error) {
	devices, err := o.inventory.ListDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	// Group devices by rack
	rackPower := make(map[string]float64)
	for _, device := range devices {
		rackPower[device.Location.Rack] += device.Metrics.PowerUsage
	}

	recommendations := make([]types.Recommendation, 0)

	// Check each rack's power distribution
	for rack, power := range rackPower {
		// Example threshold - should be configurable in production
		if power > 5000 { // 5kW per rack limit
			rackDevices := o.getHighPowerDevices(devices, rack)
			if len(rackDevices) > 0 {
				recommendations = append(recommendations, types.Recommendation{
					ID:        fmt.Sprintf("power-optimize-%s", rack),
					Priority:  2,
					Action:    "rebalance_power",
					Reason:    fmt.Sprintf("Rack %s power usage (%.0fW) exceeds recommended limit", rack, power),
					DeviceIDs: rackDevices,
				})
			}
		}
	}

	return recommendations, nil
}

// getHighPowerDevices returns IDs of devices with high power usage in a rack
func (o *Optimizer) getHighPowerDevices(devices []types.DeviceState, rack string) []types.DeviceID {
	highPowerDevices := make([]types.DeviceID, 0)
	for _, device := range devices {
		if device.Location.Rack == rack && device.Metrics.PowerUsage > 500 { // 500W threshold
			highPowerDevices = append(highPowerDevices, device.ID)
		}
	}
	return highPowerDevices
}
