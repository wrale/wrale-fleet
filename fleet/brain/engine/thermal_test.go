package engine

import (
	"context"
	"testing"
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/device"
	"github.com/wrale/wrale-fleet/fleet/brain/types"
	metalThermal "github.com/wrale/wrale-fleet/metal/core/thermal"
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

	// Create test thermal policy
	zonePolicy := &metalThermal.ThermalPolicy{
		Profile:         metalThermal.ProfileBalance,
		CPUWarning:      70.0,
		CPUCritical:     80.0,
		GPUWarning:      75.0,
		GPUCritical:     85.0,
		AmbientWarning:  35.0,
		AmbientCritical: 40.0,
		FanStartTemp:    50.0,
		FanMinSpeed:     20,
		FanMaxSpeed:     100,
		FanRampRate:     2.0,
		ThrottleTemp:    78.0,
		ResponseDelay:   time.Second * 5,
		WarningDelay:    time.Minute,
		CriticalDelay:   time.Second * 30,
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
				ThermalMetrics: &metalThermal.ThermalMetrics{
					CPUTemp:     45.0,
					GPUTemp:     40.0,
					AmbientTemp: 25.0,
					FanSpeed:    30,
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
				ThermalMetrics: &metalThermal.ThermalMetrics{
					CPUTemp:     72.0,
					GPUTemp:     68.0,
					AmbientTemp: 32.0,
					FanSpeed:    60,
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
				ThermalMetrics: &metalThermal.ThermalMetrics{
					CPUTemp:       82.0,
					GPUTemp:       78.0,
					AmbientTemp:   36.0,
					FanSpeed:      100,
					ThrottleCount: 1,
					LastUpdate:    time.Now(),
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

	t.Run("Zone Metrics", func(t *testing.T) {
		metrics, err := thermalMgr.GetZoneMetrics(ctx, "zone-1")
		if err != nil {
			t.Fatalf("Failed to get zone metrics: %v", err)
		}

		if metrics.TotalDevices != 3 {
			t.Errorf("Expected 3 devices in zone, got %d", metrics.TotalDevices)
		}

		if metrics.MaxTemp != 82.0 {
			t.Errorf("Expected max temperature 82.0, got %.1f", metrics.MaxTemp)
		}

		expectedAvg := (45.0 + 72.0 + 82.0) / 3.0
		if metrics.AverageTemp != expectedAvg {
			t.Errorf("Expected average temperature %.1f, got %.1f",
				expectedAvg, metrics.AverageTemp)
		}

		if metrics.DevicesOverTemp != 1 {
			t.Errorf("Expected 1 device over temperature, got %d",
				metrics.DevicesOverTemp)
		}
	})

	t.Run("Thermal Events", func(t *testing.T) {
		// Update hot device to trigger event
		hotMetrics := &metalThermal.ThermalMetrics{
			CPUTemp:       85.0,
			GPUTemp:       80.0,
			AmbientTemp:   38.0,
			FanSpeed:      100,
			ThrottleCount: 2,
			LastUpdate:    time.Now(),
		}

		err := thermalMgr.UpdateDeviceThermal(ctx, "hot-device", hotMetrics)
		if err != nil {
			t.Fatalf("Failed to update device thermal: %v", err)
		}

		events, err := thermalMgr.GetThermalEvents(ctx)
		if err != nil {
			t.Fatalf("Failed to get thermal events: %v", err)
		}

		if len(events) == 0 {
			t.Error("Expected at least one thermal event")
		}

		foundCritical := false
		for _, event := range events {
			if event.Type == "cpu_critical" && event.DeviceID == "hot-device" {
				foundCritical = true
				break
			}
		}

		if !foundCritical {
			t.Error("Expected critical CPU temperature event for hot device")
		}
	})

	t.Run("Policy Violations", func(t *testing.T) {
		metrics, err := thermalMgr.GetZoneMetrics(ctx, "zone-1")
		if err != nil {
			t.Fatalf("Failed to get zone metrics: %v", err)
		}

		foundViolation := false
		for _, violation := range metrics.PolicyViolations {
			if violation == "Device hot-device: CPU temperature 85.0°C exceeds critical threshold 80.0°C" {
				foundViolation = true
				break
			}
		}

		if !foundViolation {
			t.Error("Expected policy violation for hot device")
		}
	})

	t.Run("Thermal Updates", func(t *testing.T) {
		// Update cool device with new metrics
		newMetrics := &metalThermal.ThermalMetrics{
			CPUTemp:     55.0,
			GPUTemp:     50.0,
			AmbientTemp: 28.0,
			FanSpeed:    40,
			LastUpdate:  time.Now(),
		}

		err := thermalMgr.UpdateDeviceThermal(ctx, "cool-device", newMetrics)
		if err != nil {
			t.Fatalf("Failed to update device thermal: %v", err)
		}

		retrieved, err := thermalMgr.GetDeviceThermal(ctx, "cool-device")
		if err != nil {
			t.Fatalf("Failed to get device thermal: %v", err)
		}

		if retrieved.CPUTemp != newMetrics.CPUTemp {
			t.Errorf("Expected CPU temperature %.1f, got %.1f",
				newMetrics.CPUTemp, retrieved.CPUTemp)
		}

		if retrieved.FanSpeed != newMetrics.FanSpeed {
			t.Errorf("Expected fan speed %d, got %d",
				newMetrics.FanSpeed, retrieved.FanSpeed)
		}
	})

	t.Run("Invalid Device", func(t *testing.T) {
		_, err := thermalMgr.GetDeviceThermal(ctx, "nonexistent-device")
		if err == nil {
			t.Error("Expected error for nonexistent device")
		}
	})

	t.Run("Invalid Zone", func(t *testing.T) {
		_, err := thermalMgr.GetZonePolicy(ctx, "nonexistent-zone")
		if err == nil {
			t.Error("Expected error for nonexistent zone")
		}
	})

	t.Run("Event Buffer Limit", func(t *testing.T) {
		// Generate many events
		for i := 0; i < 1100; i++ { // More than maxEvents
			hotMetrics := &metalThermal.ThermalMetrics{
				CPUTemp:    85.0,
				LastUpdate: time.Now(),
			}
			err := thermalMgr.UpdateDeviceThermal(ctx, "hot-device", hotMetrics)
			if err != nil {
				t.Fatalf("Failed to update thermal: %v", err)
			}
		}

		events, err := thermalMgr.GetThermalEvents(ctx)
		if err != nil {
			t.Fatalf("Failed to get events: %v", err)
		}

		if len(events) > 1000 { // maxEvents value
			t.Errorf("Expected maximum 1000 events, got %d", len(events))
		}
	})
}