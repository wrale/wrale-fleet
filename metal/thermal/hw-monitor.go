package thermal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/metal/gpio"
)

// Monitor handles thermal hardware monitoring and control
type Monitor struct {
	mux   sync.RWMutex
	state ThermalState

	// Hardware interface
	gpio        *gpio.Controller
	fanPin      string
	throttlePin string

	// Temperature paths
	cpuTemp     string
	gpuTemp     string
	ambientTemp string

	// Configuration
	monitorInterval time.Duration
	onStateChange   func(ThermalState)
}

// New creates a new thermal monitor
func New(cfg Config) (*Monitor, error) {
	if cfg.GPIO == nil {
		return nil, fmt.Errorf("GPIO controller is required")
	}

	// Set defaults
	if cfg.MonitorInterval == 0 {
		cfg.MonitorInterval = defaultMonitorInterval
	}

	m := &Monitor{
		gpio:            cfg.GPIO,
		fanPin:          cfg.FanControlPin,
		throttlePin:     cfg.ThrottlePin,
		cpuTemp:         cfg.CPUTempPath,
		gpuTemp:         cfg.GPUTempPath,
		ambientTemp:     cfg.AmbientTempPath,
		monitorInterval: cfg.MonitorInterval,
		onStateChange:   cfg.OnStateChange,
	}

	if m.fanPin != "" {
		if err := m.InitializeFanControl(); err != nil {
			return nil, fmt.Errorf("failed to initialize fan: %w", err)
		}
	}

	return m, nil
}

// GetState returns the current thermal state
func (m *Monitor) GetState() ThermalState {
	m.mux.RLock()
	defer m.mux.RUnlock()
	return m.state
}

// Monitor starts continuous hardware monitoring
func (m *Monitor) Monitor(ctx context.Context) error {
	ticker := time.NewTicker(m.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := m.updateThermalState(); err != nil {
				return fmt.Errorf("failed to update thermal state: %w", err)
			}
		}
	}
}