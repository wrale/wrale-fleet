package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/device"
	"github.com/wrale/wrale-fleet/fleet/brain/types"
)

// Analyzer implements fleet analysis and decision making
type Analyzer struct {
	inventory *device.Inventory
	topology  *device.TopologyManager
}

// NewAnalyzer creates a new analyzer instance
func NewAnalyzer(inventory *device.Inventory, topology *device.TopologyManager) *Analyzer {
	return &Analyzer{
		inventory: inventory,
		topology:  topology,
	}
}

// AnalyzeState analyzes current fleet state
func (a *Analyzer) AnalyzeState(ctx context.Context) (*types.FleetAnalysis, error) {
	// Get all devices
	devices, err := a.inventory.ListDevices(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	// Initialize analysis
	analysis := &types.FleetAnalysis{
		TotalDevices:   len(devices),
		HealthyDevices: 0,
		ResourceUsage:  make(map[types.ResourceType]float64),
		AnalyzedAt:     time.Now(),
	}

	// Analyze devices
	var totalCPU, totalMemory, totalPower float64
	alerts := make([]types.Alert, 0)
	recommendations := make([]types.Recommendation, 0)

	for _, device := range devices {
		// Track resource usage
		totalCPU += device.Metrics.CPULoad
		totalMemory += device.Metrics.MemoryUsage
		totalPower += device.Metrics.PowerUsage

		// Check health
		if a.isHealthy(device) {
			analysis.HealthyDevices++
		}

		// Generate alerts
		deviceAlerts := a.checkDeviceAlerts(device)
		alerts = append(alerts, deviceAlerts...)

		// Generate recommendations
		deviceRecommendations := a.generateDeviceRecommendations(device)
		recommendations = append(recommendations, deviceRecommendations...)
	}

	// Calculate average resource usage
	if len(devices) > 0 {
		analysis.ResourceUsage[types.ResourceCPU] = totalCPU / float64(len(devices))
		analysis.ResourceUsage[types.ResourceMemory] = totalMemory / float64(len(devices))
		analysis.ResourceUsage[types.ResourcePower] = totalPower / float64(len(devices))
	}

	analysis.Alerts = alerts
	analysis.Recommendations = recommendations

	return analysis, nil
}

// isHealthy determines if a device is healthy
func (a *Analyzer) isHealthy(device types.DeviceState) bool {
	// Basic health checks for v1.0
	if device.Metrics.Temperature > 80 {
		return false
	}
	if device.Metrics.CPULoad > 90 {
		return false
	}
	if device.Metrics.MemoryUsage > 95 {
		return false
	}
	return true
}

// checkDeviceAlerts generates alerts for a device
func (a *Analyzer) checkDeviceAlerts(device types.DeviceState) []types.Alert {
	alerts := make([]types.Alert, 0)
	now := time.Now()

	// Temperature alerts
	if device.Metrics.Temperature > 80 {
		alerts = append(alerts, types.Alert{
			ID:        fmt.Sprintf("temp-high-%s", device.ID),
			Severity:  "critical",
			Message:   fmt.Sprintf("High temperature detected: %.1f°C", device.Metrics.Temperature),
			DeviceID:  device.ID,
			CreatedAt: now,
		})
	} else if device.Metrics.Temperature > 70 {
		alerts = append(alerts, types.Alert{
			ID:        fmt.Sprintf("temp-warn-%s", device.ID),
			Severity:  "warning",
			Message:   fmt.Sprintf("Elevated temperature: %.1f°C", device.Metrics.Temperature),
			DeviceID:  device.ID,
			CreatedAt: now,
		})
	}

	// Resource alerts
	if device.Metrics.CPULoad > 90 {
		alerts = append(alerts, types.Alert{
			ID:        fmt.Sprintf("cpu-high-%s", device.ID),
			Severity:  "critical",
			Message:   fmt.Sprintf("High CPU usage: %.1f%%", device.Metrics.CPULoad),
			DeviceID:  device.ID,
			CreatedAt: now,
		})
	}

	if device.Metrics.MemoryUsage > 95 {
		alerts = append(alerts, types.Alert{
			ID:        fmt.Sprintf("mem-high-%s", device.ID),
			Severity:  "critical",
			Message:   fmt.Sprintf("High memory usage: %.1f%%", device.Metrics.MemoryUsage),
			DeviceID:  device.ID,
			CreatedAt: now,
		})
	}

	if device.Metrics.PowerUsage > 800 { // Example threshold
		alerts = append(alerts, types.Alert{
			ID:        fmt.Sprintf("power-high-%s", device.ID),
			Severity:  "warning",
			Message:   fmt.Sprintf("High power consumption: %.1fW", device.Metrics.PowerUsage),
			DeviceID:  device.ID,
			CreatedAt: now,
		})
	}

	return alerts
}

// generateDeviceRecommendations creates optimization recommendations
func (a *Analyzer) generateDeviceRecommendations(device types.DeviceState) []types.Recommendation {
	recommendations := make([]types.Recommendation, 0)
	now := time.Now()

	// Temperature optimization
	if device.Metrics.Temperature > 75 {
		recommendations = append(recommendations, types.Recommendation{
			ID:       fmt.Sprintf("cooling-optimize-%s", device.ID),
			Priority: 1,
			Action:   "optimize_cooling",
			Reason:   fmt.Sprintf("Temperature above optimal range: %.1f°C", device.Metrics.Temperature),
			DeviceIDs: []types.DeviceID{device.ID},
			CreatedAt: now,
		})
	}

	// Resource optimization
	if device.Metrics.CPULoad > 85 {
		recommendations = append(recommendations, types.Recommendation{
			ID:       fmt.Sprintf("workload-balance-%s", device.ID),
			Priority: 2,
			Action:   "balance_workload",
			Reason:   fmt.Sprintf("High CPU utilization: %.1f%%", device.Metrics.CPULoad),
			DeviceIDs: []types.DeviceID{device.ID},
			CreatedAt: now,
		})
	}

	// Power optimization
	if device.Metrics.PowerUsage > 750 {
		recommendations = append(recommendations, types.Recommendation{
			ID:       fmt.Sprintf("power-optimize-%s", device.ID),
			Priority: 3,
			Action:   "optimize_power",
			Reason:   fmt.Sprintf("Power usage above optimal: %.1fW", device.Metrics.PowerUsage),
			DeviceIDs: []types.DeviceID{device.ID},
			CreatedAt: now,
		})
	}

	return recommendations
}

// GetAlerts returns current alerts
func (a *Analyzer) GetAlerts(ctx context.Context) ([]types.Alert, error) {
	analysis, err := a.AnalyzeState(ctx)
	if err != nil {
		return nil, err
	}
	return analysis.Alerts, nil
}

// GetRecommendations returns current recommendations
func (a *Analyzer) GetRecommendations(ctx context.Context) ([]types.Recommendation, error) {
	analysis, err := a.AnalyzeState(ctx)
	if err != nil {
		return nil, err
	}
	return analysis.Recommendations, nil
}