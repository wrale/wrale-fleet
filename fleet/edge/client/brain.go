package client

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/wrale/wrale-fleet/fleet/brain/types"
    "github.com/wrale/wrale-fleet/fleet/edge/agent"
)

// BrainClient implements communication with the fleet brain
type BrainClient struct {
    baseURL     string
    deviceID    types.DeviceID
    httpClient  *http.Client
    retryConfig RetryConfig
}

// RetryConfig defines retry behavior for brain communication
type RetryConfig struct {
    MaxAttempts  int
    InitialDelay time.Duration
    MaxDelay     time.Duration
}

// DefaultRetryConfig provides sensible default retry settings
var DefaultRetryConfig = RetryConfig{
    MaxAttempts:  5,
    InitialDelay: time.Second,
    MaxDelay:     time.Second * 30,
}

// NewBrainClient creates a new brain client instance
func NewBrainClient(baseURL string, deviceID types.DeviceID, retryConfig *RetryConfig) *BrainClient {
    if retryConfig == nil {
        retryConfig = &DefaultRetryConfig
    }

    return &BrainClient{
        baseURL:  baseURL,
        deviceID: deviceID,
        httpClient: &http.Client{
            Timeout: time.Second * 30,
        },
        retryConfig: *retryConfig,
    }
}

// SyncState synchronizes device state with the brain
func (c *BrainClient) SyncState(state types.DeviceState) error {
    url := fmt.Sprintf("%s/api/v1/devices/%s/state", c.baseURL, c.deviceID)
    return c.retryRequest("POST", url, state, nil)
}

// GetCommands retrieves pending commands from the brain
func (c *BrainClient) GetCommands() ([]agent.Command, error) {
    url := fmt.Sprintf("%s/api/v1/devices/%s/commands", c.baseURL, c.deviceID)
    var commands []agent.Command
    err := c.retryRequest("GET", url, nil, &commands)
    return commands, err
}

// ReportCommandResult reports command execution result to brain
func (c *BrainClient) ReportCommandResult(result agent.CommandResult) error {
    url := fmt.Sprintf("%s/api/v1/devices/%s/commands/%s/result", 
        c.baseURL, c.deviceID, result.CommandID)
    return c.retryRequest("POST", url, result, nil)
}

// ReportHealth reports device health status to brain
func (c *BrainClient) ReportHealth(healthy bool, details map[string]interface{}) error {
    url := fmt.Sprintf("%s/api/v1/devices/%s/health", c.baseURL, c.deviceID)
    payload := map[string]interface{}{
        "healthy": healthy,
        "details": details,
    }
    return c.retryRequest("POST", url, payload, nil)
}

// GetConfig retrieves device configuration from brain
func (c *BrainClient) GetConfig() (map[string]interface{}, error) {
    url := fmt.Sprintf("%s/api/v1/devices/%s/config", c.baseURL, c.deviceID)
    var config map[string]interface{}
    err := c.retryRequest("GET", url, nil, &config)
    return config, err
}

// retryRequest performs an HTTP request with retries
func (c *BrainClient) retryRequest(method, url string, payload, response interface{}) error {
    var delay time.Duration = c.retryConfig.InitialDelay

    for attempt := 1; attempt <= c.retryConfig.MaxAttempts; attempt++ {
        err := c.doRequest(method, url, payload, response)
        if err == nil {
            return nil
        }

        // Don't retry on context cancellation or fatal errors
        if err == context.Canceled || err == context.DeadlineExceeded {
            return err
        }

        // Last attempt failed
        if attempt == c.retryConfig.MaxAttempts {
            return fmt.Errorf("all retry attempts failed: %w", err)
        }

        // Wait before retry
        time.Sleep(delay)

        // Exponential backoff with max delay
        delay *= 2
        if delay > c.retryConfig.MaxDelay {
            delay = c.retryConfig.MaxDelay
        }
    }

    return fmt.Errorf("retry loop completed without return")
}

// doRequest performs a single HTTP request
func (c *BrainClient) doRequest(method, url string, payload, response interface{}) error {
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
    req.Header.Set("X-Device-ID", string(c.deviceID))

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
