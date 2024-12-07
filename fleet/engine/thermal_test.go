package engine

import (
	"context"
	"testing"
	"time"

	"github.com/wrale/wrale-fleet/fleet/device"
	"github.com/wrale/wrale-fleet/fleet/types"
)

func TestThermalManager(t *testing.T) {
	ctx := context.Background()
	inventory := device.NewInventory()
	topology := device.NewTopologyManager(inventory)
	analyzer := NewAnalyzer(inventory, topology)
	thermalMgr := NewThermalManager(inventory, topology, analyzer)

	// Register test rack and cooling zone
	err := topology.RegisterRack(ctx, "rack-1", device.RackConfig{
		MaxUnits:    42,
		PowerLimit:  5000.0,
		CoolingZone: "zone-1",
	})
	if err != nil {
		t.Fatalf("Failed to register test rack: %v", err)
	}

	// Create test zone policy
	zonePolicy := &types.ThermalPolicy{
		Profile:         types.ProfileBalance,
		CPUWarning:      70.0,
		CPUCritical:     80.0,
		GPUWarning:      75.0,
		GPUCritical:     85.0,
		AmbientWarning:  35.0,
		AmbientCritical: 40.0,
		MonitoringInterval: time.Second * 5,
		AlertInterval:      time.Minute,
		AutoThrottle:       true,
		MaxDevicesThrottled: 2,
		ZonePriority:       1,
	}

	// Initialize test devices with different thermal states
	devices := []types.DeviceState{
		{
			ID: "cool-device",
			Status: "active",
			Location: types.PhysicalLocation{
				Rack:     "rack-1",
				Position: 1,
				Zone:     "zone-1",
			},
			Metrics: types.DeviceMetrics{
				ThermalMetrics: &types.ThermalMetrics{
					CPUTemp:     45.0,
					GPUTemp:     40.0,
					AmbientTemp: 25.0,
					FanSpeed:    30,
					IsThrottled: false,
					LastUpdate:  time.Now(),
				},
			},
		},
		{
			ID: "warm-device",
			Status: "active",
			Location: types.PhysicalLocation{
				Rack:     "rack-1",
				Position: 2,
				Zone:     "zone-1",
			},
			Metrics: types.DeviceMetrics{
				ThermalMetrics: &types.ThermalMetrics{
					CPUTemp:     72.0,
					GPUTemp:     68.0,
					AmbientTemp: 32.0,
					FanSpeed:    60,
					IsThrottled: false,
					LastUpdate:  time.Now(),
				},
			},
		},
		{
			ID: "hot-device",
			Status: "active",
			Location: types.PhysicalLocation{
				Rack:     "rack-1",
				Position: 3,
				Zone:     "zone-1",
			},
			Metrics: types.DeviceMetrics{
				ThermalMetrics: &types.ThermalMetrics{
					CPUTemp:     82.0,
					GPUTemp:     78.0,
					AmbientTemp: 36.0,
					FanSpeed:    100,
					IsThrottled: false,
					LastUpdate:  time.Now(),
				},
			},
		},
	}

	// Register test devices
	for _, d := range devices {
		err := inventory.RegisterDevice(ctx, d)
		if err != nil {
			t.Fatalf("Failed to register device %s: %v", d.ID, err)
		}
	}

	t.Run("Set Zone Policy", func(t *testing.T) {
		err := thermalMgr.SetZonePolicy(ctx, "zone-1", zonePolicy)
		if err != nil {
			t.Fatalf("Failed to set zone policy: %v", err)
		}

		retrievedPolicy, err := thermalMgr.GetZonePolicy(ctx, "zone-1")
		if err != nil {
			t.Fatalf("Failed to get zone policy: %v", err)
		}

		if retrievedPolicy.CPUCritical != zonePolicy.CPUCritical {
			t.Errorf("Expected CPU critical threshold %.1f, got %.1f",
				zonePolicy.CPUCritical, retrievedPolicy.CPUCritical)
		}
	})

	t.Run("Device Policy Override", func(t *testing.T) {
		devicePolicy := *zonePolicy // Copy zone policy
		devicePolicy.CPUCritical = 85.0 // Higher threshold for specific device
		devicePolicy.AutoThrottle = false // Disable auto-throttling

		err := thermalMgr.SetDevicePolicy(ctx, "hot-device", &devicePolicy)
		if err != nil {
			t.Fatalf("Failed to set device policy: %v", err)
		}

		retrievedPolicy, err := thermalMgr.GetDevicePolicy(ctx, "hot-device")
		if err != nil {
			t.Fatalf("Failed to get device policy: %v", err)
		}

		if retrievedPolicy.CPUCritical != devicePolicy.CPUCritical {
			t.Errorf("Expected device CPU critical threshold %.1f, got %.1f",
				devicePolicy.CPUCritical, retrievedPolicy.CPUCritical)
		}
	})

	t.Run("Auto Throttling", func(t *testing.T) {
		// Update hot device to trigger throttling
		hotMetrics := &types.ThermalMetrics{
			CPUTemp:     85.0,
			GPUTemp:     80.0,
			AmbientTemp: 38.0,
			FanSpeed:    100,
			IsThrottled: false,
			LastUpdate:  time.Now(),
		}

		err := thermalMgr.UpdateDeviceThermal(ctx, "hot-device", hotMetrics)
		if err != nil {
			t.Fatalf("Failed to update device thermal: %v", err)
		}

		// Verify device was throttled
		updatedMetrics, err := thermalMgr.GetDeviceThermal(ctx, "hot-device")
		if err != nil {
			t.Fatalf("Failed to get device thermal: %v", err)
		}

		if !updatedMetrics.IsThrottled {
			t.Error("Expected device to be throttled")
		}
	})

	t.Run("Zone Metrics", func(t *testing.T) {
		metrics, err := thermalMgr.GetZoneMetrics(ctx, "zone-1")
		if err != nil {
			t.Fatalf("Failed to get zone metrics: %v", err)
		}

		if metrics.TotalDevices != 3 {
			t.Errorf("Expected 3 devices in zone, got %d", metrics.TotalDevices)
		}

		if metrics.MaxTemp != 85.0 {
			t.Errorf("Expected max temperature 85.0, got %.1f", metrics.MaxTemp)
		}

		expectedAvg := (45.0 + 72.0 + 85.0) / 3.0
		if metrics.AverageTemp != expectedAvg {
			t.Errorf("Expected average temperature %.1f, got %.1f",
				expectedAvg, metrics.AverageTemp)
		}

		if metrics.DevicesOverTemp != 1 {
			t.Errorf("Expected 1 device over temperature, got %d",
				metrics.DevicesOverTemp)
		}

		if metrics.DevicesThrottled != 1 {
			t.Errorf("Expected 1 device throttled, got %d",
				metrics.DevicesThrottled)
		}
	})

	t.Run("Thermal Events", func(t *testing.T) {
		events, err := thermalMgr.GetThermalEvents(ctx)
		if err != nil {
			t.Fatalf("Failed to get thermal events: %v", err)
		}

		if len(events) == 0 {
			t.Error("Expected at least one thermal event")
		}

		foundCritical := false
		for _, event := range events {
			if event.Type == "critical" && event.DeviceID == "hot-device" {
				foundCritical = true
				if !event.Throttled {
					t.Error("Expected critical event to indicate throttling")
				}
				break
			}
		}

		if !foundCritical {
			t.Error("Expected critical temperature event for hot device")
		}
	})

	t.Run("Max Throttled Devices", func(t *testing.T) {
		// Update warm device to also trigger throttling
		warmMetrics := &types.ThermalMetrics{
			CPUTemp:     82.0,
			GPUTemp:     76.0,
			AmbientTemp: 35.0,
			FanSpeed:    90,
			IsThrottled: false,
			LastUpdate:  time.Now(),
		}

		err := thermalMgr.UpdateDeviceThermal(ctx, "warm-device", warmMetrics)
		if err != nil {
			t.Fatalf("Failed to update device thermal: %v", err)
		}

		// Update cool device to also exceed threshold
		coolMetrics := &types.ThermalMetrics{
			CPUTemp:     81.0,
			GPUTemp:     75.0,
			AmbientTemp: 34.0,
			FanSpeed:    85,
			IsThrottled: false,
			LastUpdate:  time.Now(),
		}

		err = thermalMgr.UpdateDeviceThermal(ctx, "cool-device", coolMetrics)
		if err != nil {
			t.Fatalf("Failed to update device thermal: %v", err)
		}

		// Check zone metrics
		metrics, err := thermalMgr.GetZoneMetrics(ctx, "zone-1")
		if err != nil {
			t.Fatalf("Failed to get zone metrics: %v", err)
		}

		// Should not exceed MaxDevicesThrottled
		if metrics.DevicesThrottled > zonePolicy.MaxDevicesThrottled {
			t.Errorf("Zone has %d throttled devices, exceeding limit of %d",
				metrics.DevicesThrottled, zonePolicy.MaxDevicesThrottled)
		}

		// Should have violation for device that couldn't be throttled
		hasViolation := false
		for _, violation := range metrics.PolicyViolations {
			if violation == "Device cool-device: CPU temperature 81.0°C exceeds critical threshold 80.0°C but zone throttle limit reached" {
				hasViolation = true
				break
			}
		}
		if !hasViolation {
			t.Error("Expected policy violation for device that couldn't be throttled")
		}
	})
}