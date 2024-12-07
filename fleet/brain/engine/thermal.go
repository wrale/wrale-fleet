package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/device"
	"github.com/wrale/wrale-fleet/fleet/brain/types"
	metalThermal "github.com/wrale/wrale-fleet/metal/core/thermal"
)

// ThermalManager implements fleet-wide thermal management
type ThermalManager struct {
	inventory *device.Inventory
	topology  *device.TopologyManager
	analyzer  *Analyzer

	// Cache thermal policies
	policyCache     map[types.DeviceID]*metalThermal.ThermalPolicy
	zonePolicyCache map[string]*metalThermal.ThermalPolicy
	cacheMutex      sync.RWMutex

	// Track thermal events
	recentEvents []metalThermal.ThermalEvent
	eventsMutex  sync.RWMutex
	maxEvents    int
}

// NewThermalManager creates a new thermal manager instance
func NewThermalManager(inventory *device.Inventory, topology *device.TopologyManager, analyzer *Analyzer) *ThermalManager {
	return &ThermalManager{
		inventory:       inventory,
		topology:       topology,
		analyzer:       analyzer,
		policyCache:    make(map[types.DeviceID]*metalThermal.ThermalPolicy),
		zonePolicyCache: make(map[string]*metalThermal.ThermalPolicy),
		maxEvents:      1000, // Keep last 1000 events
	}
}

// UpdateDeviceThermal processes updated thermal metrics from a device
func (tm *ThermalManager) UpdateDeviceThermal(ctx context.Context, deviceID types.DeviceID, metrics *metalThermal.ThermalMetrics) error {
	// Get device state
	device, err := tm.inventory.GetDevice(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}

	// Update metrics
	device.Metrics.ThermalMetrics = metrics

	// Apply device policy
	if err := tm.applyDevicePolicy(ctx, device); err != nil {
		return fmt.Errorf("failed to apply device policy: %w", err)
	}

	// Update state
	if err := tm.inventory.UpdateState(ctx, *device); err != nil {
		return fmt.Errorf("failed to update device state: %w", err)
	}

	return nil
}

// GetDeviceThermal retrieves current thermal metrics for a device
func (tm *ThermalManager) GetDeviceThermal(ctx context.Context, deviceID types.DeviceID) (*metalThermal.ThermalMetrics, error) {
	device, err := tm.inventory.GetDevice(ctx, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	return device.Metrics.ThermalMetrics, nil
}

// SetDevicePolicy updates thermal policy for a specific device
func (tm *ThermalManager) SetDevicePolicy(ctx context.Context, deviceID types.DeviceID, policy *metalThermal.ThermalPolicy) error {
	tm.cacheMutex.Lock()
	tm.policyCache[deviceID] = policy
	tm.cacheMutex.Unlock()

	// Apply policy immediately
	device, err := tm.inventory.GetDevice(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}

	return tm.applyDevicePolicy(ctx, device)
}

// GetDevicePolicy retrieves thermal policy for a device
func (tm *ThermalManager) GetDevicePolicy(ctx context.Context, deviceID types.DeviceID) (*metalThermal.ThermalPolicy, error) {
	tm.cacheMutex.RLock()
	policy, ok := tm.policyCache[deviceID]
	tm.cacheMutex.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no policy found for device %s", deviceID)
	}
	return policy, nil
}

// SetZonePolicy updates thermal policy for all devices in a zone
func (tm *ThermalManager) SetZonePolicy(ctx context.Context, zone string, policy *metalThermal.ThermalPolicy) error {
	tm.cacheMutex.Lock()
	tm.zonePolicyCache[zone] = policy
	tm.cacheMutex.Unlock()

	// Apply to all devices in zone
	devices, err := tm.topology.GetDevicesInZone(ctx, zone)
	if err != nil {
		return fmt.Errorf("failed to get zone devices: %w", err)
	}

	for _, device := range devices {
		if err := tm.applyDevicePolicy(ctx, &device); err != nil {
			return fmt.Errorf("failed to apply policy to device %s: %w", device.ID, err)
		}
	}

	return nil
}

// GetZonePolicy retrieves thermal policy for a zone
func (tm *ThermalManager) GetZonePolicy(ctx context.Context, zone string) (*metalThermal.ThermalPolicy, error) {
	tm.cacheMutex.RLock()
	policy, ok := tm.zonePolicyCache[zone]
	tm.cacheMutex.RUnlock()

	if !ok {
		return nil, fmt.Errorf("no policy found for zone %s", zone)
	}
	return policy, nil
}

// GetZoneMetrics calculates thermal metrics for a zone
func (tm *ThermalManager) GetZoneMetrics(ctx context.Context, zone string) (*types.ZoneThermalMetrics, error) {
	devices, err := tm.topology.GetDevicesInZone(ctx, zone)
	if err != nil {
		return nil, fmt.Errorf("failed to get zone devices: %w", err)
	}

	metrics := &types.ZoneThermalMetrics{
		Zone:         zone,
		TotalDevices: len(devices),
		UpdatedAt:    time.Now(),
	}

	if len(devices) == 0 {
		return metrics, nil
	}

	// Calculate zone metrics
	var totalTemp float64
	metrics.MaxTemp = -1000 // Initialize to impossible low
	metrics.MinTemp = 1000  // Initialize to impossible high

	for _, device := range devices {
		if device.Metrics.ThermalMetrics == nil {
			continue
		}

		temp := device.Metrics.ThermalMetrics.CPUTemp
		totalTemp += temp

		if temp > metrics.MaxTemp {
			metrics.MaxTemp = temp
		}
		if temp < metrics.MinTemp {
			metrics.MinTemp = temp
		}

		// Check against zone policy
		if err := tm.checkZonePolicy(ctx, zone, device); err != nil {
			metrics.PolicyViolations = append(metrics.PolicyViolations, 
				fmt.Sprintf("Device %s: %v", device.ID, err))
		}

		// Count devices over temperature
		if temp > device.Metrics.ThermalMetrics.CPUTemp {
			metrics.DevicesOverTemp++
		}
	}

	metrics.AverageTemp = totalTemp / float64(metrics.TotalDevices)
	return metrics, nil
}

// GetThermalEvents returns recent thermal events
func (tm *ThermalManager) GetThermalEvents(ctx context.Context) ([]metalThermal.ThermalEvent, error) {
	tm.eventsMutex.RLock()
	events := make([]metalThermal.ThermalEvent, len(tm.recentEvents))
	copy(events, tm.recentEvents)
	tm.eventsMutex.RUnlock()
	return events, nil
}

// applyDevicePolicy applies and validates thermal policy for a device
func (tm *ThermalManager) applyDevicePolicy(ctx context.Context, device *types.DeviceState) error {
	// Get effective policy (device-specific overrides zone policy)
	tm.cacheMutex.RLock()
	policy, ok := tm.policyCache[device.ID]
	if !ok {
		policy, ok = tm.zonePolicyCache[device.Location.Zone]
	}
	tm.cacheMutex.RUnlock()

	if !ok {
		return nil // No policy to apply
	}

	if device.Metrics.ThermalMetrics == nil {
		return nil // No thermal metrics to check
	}

	metrics := device.Metrics.ThermalMetrics

	// Check thresholds and generate events
	if metrics.CPUTemp > policy.CPUCritical {
		tm.addThermalEvent(metalThermal.ThermalEvent{
			DeviceID:    string(device.ID),
			Zone:        device.Location.Zone,
			Type:        "cpu_critical",
			Temperature: metrics.CPUTemp,
			Threshold:   policy.CPUCritical,
			State:       tm.getHWState(metrics),
			Timestamp:   time.Now(),
		})
	} else if metrics.CPUTemp > policy.CPUWarning {
		tm.addThermalEvent(metalThermal.ThermalEvent{
			DeviceID:    string(device.ID),
			Zone:        device.Location.Zone,
			Type:        "cpu_warning",
			Temperature: metrics.CPUTemp,
			Threshold:   policy.CPUWarning,
			State:       tm.getHWState(metrics),
			Timestamp:   time.Now(),
		})
	}

	return nil
}

// checkZonePolicy validates a device against zone policy
func (tm *ThermalManager) checkZonePolicy(ctx context.Context, zone string, device types.DeviceState) error {
	tm.cacheMutex.RLock()
	policy, ok := tm.zonePolicyCache[zone]
	tm.cacheMutex.RUnlock()

	if !ok {
		return nil // No policy to check
	}

	if device.Metrics.ThermalMetrics == nil {
		return nil
	}

	metrics := device.Metrics.ThermalMetrics

	if metrics.CPUTemp > policy.CPUCritical {
		return fmt.Errorf("CPU temperature %.1f°C exceeds critical threshold %.1f°C",
			metrics.CPUTemp, policy.CPUCritical)
	}

	return nil
}

// addThermalEvent adds a new thermal event to the history
func (tm *ThermalManager) addThermalEvent(event metalThermal.ThermalEvent) {
	tm.eventsMutex.Lock()
	defer tm.eventsMutex.Unlock()

	tm.recentEvents = append(tm.recentEvents, event)
	if len(tm.recentEvents) > tm.maxEvents {
		// Remove oldest events when limit reached
		tm.recentEvents = tm.recentEvents[len(tm.recentEvents)-tm.maxEvents:]
	}
}

// getHWState converts metrics to hardware state
func (tm *ThermalManager) getHWState(metrics *metalThermal.ThermalMetrics) metalThermal.ThermalState {
	return metalThermal.ThermalState{
		CPUTemp:     metrics.CPUTemp,
		GPUTemp:     metrics.GPUTemp,
		AmbientTemp: metrics.AmbientTemp,
		FanSpeed:    metrics.FanSpeed,
		Throttled:   metrics.ThrottleCount > 0,
		UpdatedAt:   metrics.LastUpdate,
	}
}