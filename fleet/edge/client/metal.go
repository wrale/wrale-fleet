package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/wrale/wrale-fleet/metal/hw/diag"
	"github.com/wrale/wrale-fleet/metal/hw/power"
	"github.com/wrale/wrale-fleet/metal/hw/thermal"
)

// MetalClient represents a client connection to the metal layer
type MetalClient struct {
	baseURL    string
	httpClient *http.Client
	powerMgr   power.Manager // For direct hardware access
}

// MetricsResponse represents system metrics from the metal layer
type MetricsResponse struct {
	Temperature float64   `json:"temperature"`
	PowerUsage  float64   `json:"power_usage"`
	CPULoad     float64   `json:"cpu_load"`
	MemoryUsage float64   `json:"memory_usage"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewMetalClient creates a new metal client with the given base URL and optional power manager
func NewMetalClient(baseURL string, powerMgr power.Manager) *MetalClient {
	return &MetalClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
		powerMgr:   powerMgr,
	}
}

// GetPowerState retrieves current power system state
func (c *MetalClient) GetPowerState() (*power.PowerState, error) {
	if c.powerMgr != nil {
		// Direct hardware access
		state := c.powerMgr.GetState()
		return &state, nil
	}

	// HTTP fallback
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/power/state", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get power state: %v", err)
	}
	defer resp.Body.Close()

	var state power.PowerState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("failed to decode power state: %v", err)
	}
	return &state, nil
}

// GetMetrics retrieves current system metrics
func (c *MetalClient) GetMetrics() (*MetricsResponse, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/metrics", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %v", err)
	}
	defer resp.Body.Close()

	var metrics MetricsResponse
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, fmt.Errorf("failed to decode metrics response: %v", err)
	}
	return &metrics, nil
}

// UpdatePowerState updates the device power state
func (c *MetalClient) UpdatePowerState(powerState *power.PowerState) error {
	payload, err := json.Marshal(powerState)
	if err != nil {
		return fmt.Errorf("failed to marshal power state: %v", err)
	}

	url := fmt.Sprintf("%s/power/state", c.baseURL)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update power state: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update power state: got status %d", resp.StatusCode)
	}
	return nil
}

// GetHealthStatus retrieves the current health status of the device
func (c *MetalClient) GetHealthStatus() (bool, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/health", c.baseURL))
	if err != nil {
		return false, fmt.Errorf("failed to get health status: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}
	return true, nil
}

// RunDiagnostics runs system diagnostics and returns results
func (c *MetalClient) RunDiagnostics() (*diag.TestResult, error) {
	resp, err := c.httpClient.Post(fmt.Sprintf("%s/diagnostics/run", c.baseURL), "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to run diagnostics: %v", err)
	}
	defer resp.Body.Close()

	var results diag.TestResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode diagnostics results: %v", err)
	}
	return &results, nil
}

// GetThermalState retrieves the current thermal state
func (c *MetalClient) GetThermalState() (*thermal.ThermalState, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/thermal/state", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get thermal state: %v", err)
	}
	defer resp.Body.Close()

	var state thermal.ThermalState
	if err := json.NewDecoder(resp.Body).Decode(&state); err != nil {
		return nil, fmt.Errorf("failed to decode thermal state: %v", err)
	}
	return &state, nil
}

// ExecuteOperation executes a generic metal operation
func (c *MetalClient) ExecuteOperation(operation string, params map[string]interface{}) error {
	if params == nil {
		params = make(map[string]interface{})
	}

	payload, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("failed to marshal operation parameters: %v", err)
	}

	url := fmt.Sprintf("%s/operations/%s", c.baseURL, operation)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to execute operation %s: %v", operation, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("operation %s failed with status %d", operation, resp.StatusCode)
		}
		return fmt.Errorf("operation %s failed: %s", operation, errResp.Error)
	}

	return nil
}
