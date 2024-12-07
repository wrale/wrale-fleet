package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/wrale/wrale-fleet/fleet/brain/types"
)

// MetalClient implements communication with the metal layer
type MetalClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewMetalClient creates a new metal client instance
func NewMetalClient(baseURL string) *MetalClient {
	return &MetalClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Second * 10, // Shorter timeout for local operations
		},
	}
}

// GetMetrics retrieves current device metrics
func (c *MetalClient) GetMetrics() (types.DeviceMetrics, error) {
	var metrics types.DeviceMetrics
	url := fmt.Sprintf("%s/api/v1/metrics", c.baseURL)

	if err := c.doRequest("GET", url, nil, &metrics); err != nil {
		return metrics, fmt.Errorf("failed to get metrics: %w", err)
	}

	return metrics, nil
}

// GetThermalState retrieves current thermal state
func (c *MetalClient) GetThermalState() (types.ThermalState, error) {
	var state types.ThermalState
	url := fmt.Sprintf("%s/api/v1/thermal/state", c.baseURL)

	if err := c.doRequest("GET", url, nil, &state); err != nil {
		return state, fmt.Errorf("failed to get thermal state: %w", err)
	}

	return state, nil
}

// UpdateThermalPolicy updates the thermal management policy
func (c *MetalClient) UpdateThermalPolicy(policy types.ThermalPolicy) error {
	url := fmt.Sprintf("%s/api/v1/thermal/policy", c.baseURL)

	if err := c.doRequest("PUT", url, policy, nil); err != nil {
		return fmt.Errorf("failed to update thermal policy: %w", err)
	}

	return nil
}

// SetFanSpeed sets the cooling fan speed
func (c *MetalClient) SetFanSpeed(speed uint32) error {
	url := fmt.Sprintf("%s/api/v1/thermal/fan", c.baseURL)
	payload := map[string]uint32{"speed": speed}

	if err := c.doRequest("PUT", url, payload, nil); err != nil {
		return fmt.Errorf("failed to set fan speed: %w", err)
	}

	return nil
}

// SetThrottling enables or disables thermal throttling
func (c *MetalClient) SetThrottling(enabled bool) error {
	url := fmt.Sprintf("%s/api/v1/thermal/throttle", c.baseURL)
	payload := map[string]bool{"enabled": enabled}

	if err := c.doRequest("PUT", url, payload, nil); err != nil {
		return fmt.Errorf("failed to set throttling: %w", err)
	}

	return nil
}

// UpdatePowerState updates the device power state
func (c *MetalClient) UpdatePowerState(state string) error {
	url := fmt.Sprintf("%s/api/v1/power/state", c.baseURL)
	payload := map[string]string{"state": state}

	if err := c.doRequest("PUT", url, payload, nil); err != nil {
		return fmt.Errorf("failed to update power state: %w", err)
	}

	return nil
}

// UpdateThermalConfig updates thermal management configuration
func (c *MetalClient) UpdateThermalConfig(config map[string]interface{}) error {
	url := fmt.Sprintf("%s/api/v1/thermal/config", c.baseURL)

	if err := c.doRequest("PUT", url, config, nil); err != nil {
		return fmt.Errorf("failed to update thermal config: %w", err)
	}

	return nil
}

// ExecuteOperation executes a hardware operation
func (c *MetalClient) ExecuteOperation(operation string) error {
	url := fmt.Sprintf("%s/api/v1/operations", c.baseURL)
	payload := map[string]string{"operation": operation}

	if err := c.doRequest("POST", url, payload, nil); err != nil {
		return fmt.Errorf("failed to execute operation: %w", err)
	}

	return nil
}

// GetOperationStatus retrieves the status of an operation
func (c *MetalClient) GetOperationStatus(operationID string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/operations/%s/status", c.baseURL, operationID)
	var result struct {
		Status string `json:"status"`
	}

	if err := c.doRequest("GET", url, nil, &result); err != nil {
		return "", fmt.Errorf("failed to get operation status: %w", err)
	}

	return result.Status, nil
}

// GetHealthStatus checks device health
func (c *MetalClient) GetHealthStatus() (bool, error) {
	url := fmt.Sprintf("%s/api/v1/health", c.baseURL)
	var result struct {
		Healthy bool `json:"healthy"`
	}

	if err := c.doRequest("GET", url, nil, &result); err != nil {
		return false, fmt.Errorf("failed to get health status: %w", err)
	}

	return result.Healthy, nil
}

// RunDiagnostics executes hardware diagnostics
func (c *MetalClient) RunDiagnostics() (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v1/diagnostics", c.baseURL)
	var result map[string]interface{}

	if err := c.doRequest("POST", url, nil, &result); err != nil {
		return nil, fmt.Errorf("failed to run diagnostics: %w", err)
	}

	return result, nil
}

// doRequest performs an HTTP request to the metal API
func (c *MetalClient) doRequest(method, url string, payload, response interface{}) error {
	var req *http.Request
	var err error

	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		req, err = http.NewRequest(method, url, bytes.NewReader(data))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Perform request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode >= 400 {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	// Parse response if needed
	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
