package diag

import (
    "context"
    "fmt"
    "sync"
    "time"
    "runtime"

    "github.com/wrale/wrale-fleet/metal"
)

// Manager handles hardware diagnostics and testing
type Manager struct {
    mux      sync.RWMutex
    cfg      metal.DiagnosticManagerConfig
    running  bool
    results  []metal.TestResult
    testID   int

    // Resource tracking
    resources map[string]float64

    // Test state
    currentTest    string
    onTestStart   func(metal.TestType, string)
    onTestComplete func(metal.TestResult)
}

// New creates a new hardware diagnostics manager
func New(cfg metal.DiagnosticManagerConfig, opts ...metal.Option) (metal.DiagnosticManager, error) {
    if cfg.GPIO == nil {
        return nil, fmt.Errorf("GPIO controller required")
    }

    m := &Manager{
        cfg:           cfg,
        onTestStart:   cfg.OnTestStart,
        onTestComplete: cfg.OnTestComplete,
        results:       make([]metal.TestResult, 0),
        resources:     make(map[string]float64),
    }

    for _, opt := range opts {
        if err := opt(m); err != nil {
            return nil, fmt.Errorf("option error: %w", err)
        }
    }

    return m, nil
}

// GetState implements Monitor interface
func (m *Manager) GetState() interface{} {
    m.mux.RLock()
    defer m.mux.RUnlock()
    
    results := make([]metal.TestResult, len(m.results))
    copy(results, m.results)
    return results
}

// GetResourceUsage returns resource utilization for a component
func (m *Manager) GetResourceUsage(component string) (map[string]float64, error) {
    m.mux.RLock()
    defer m.mux.RUnlock()

    switch component {
    case "system":
        var memStats runtime.MemStats
        runtime.ReadMemStats(&memStats)
        
        return map[string]float64{
            "cpu_usage": m.getCPUUsage(),
            "mem_usage": float64(memStats.Alloc) / float64(memStats.Sys) * 100,
            "goroutines": float64(runtime.NumGoroutine()),
        }, nil

    case "gpio":
        if m.cfg.GPIO == nil {
            return nil, fmt.Errorf("GPIO controller not available")
        }
        return map[string]float64{
            "pins_used": float64(len(m.cfg.GPIO.ListPins())),
        }, nil

    case "power":
        if m.cfg.PowerManager == nil {
            return nil, fmt.Errorf("power manager not available")
        }
        state, err := m.cfg.PowerManager.GetState()
        if err != nil {
            return nil, err
        }
        return map[string]float64{
            "voltage": state.Voltage,
            "current": state.CurrentDraw,
            "power": state.PowerConsumption,
        }, nil

    case "thermal":
        if m.cfg.ThermalManager == nil {
            return nil, fmt.Errorf("thermal manager not available")
        }
        temp, err := m.cfg.ThermalManager.GetTemperature()
        if err != nil {
            return nil, err
        }
        return map[string]float64{
            "temperature": temp,
        }, nil

    default:
        return nil, fmt.Errorf("unknown component: %s", component)
    }
}

// MonitorResources continuously monitors resource usage
func (m *Manager) MonitorResources(ctx context.Context) (<-chan map[string]float64, error) {
    ch := make(chan map[string]float64, 10)
    go func() {
        defer close(ch)
        ticker := time.NewTicker(time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                resources, err := m.GetResourceUsage("system")
                if err != nil {
                    continue
                }
                select {
                case ch <- resources:
                default:
                }
            }
        }
    }()
    return ch, nil
}

// Close implements Monitor interface
func (m *Manager) Close() error {
    m.mux.Lock()
    defer m.mux.Unlock()

    m.running = false
    m.results = nil
    m.currentTest = ""
    return nil
}

// Test execution
[Previous test execution code remains unchanged...]

// Internal helpers

func (m *Manager) getCPUUsage() float64 {
    // Simple CPU usage approximation
    // In a real implementation, this would read from /proc/stat or equivalent
    return 0.0
}

[Rest of the implementation remains unchanged...]
