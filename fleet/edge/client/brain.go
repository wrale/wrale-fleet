package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wrale/wrale-fleet/fleet/types"
)

// BrainClient represents a client connection to the fleet brain
type BrainClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewBrainClient creates a new brain client with the given base URL
func NewBrainClient(baseURL string) *BrainClient {
	return &BrainClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

// RegisterDevices registers devices with the brain
func (c *BrainClient) RegisterDevices(devices []types.DeviceID) error {
	payload, err := json.Marshal(devices)
	if err != nil {
		return fmt.Errorf("failed to marshal devices: %v", err)
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/devices/register", c.baseURL),
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return fmt.Errorf("failed to register devices: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register devices: got status %d", resp.StatusCode)
	}
	return nil
}

// SyncDevices synchronizes device state with the brain
func (c *BrainClient) SyncDevices(devices []types.DeviceID) error {
	payload, err := json.Marshal(devices)
	if err != nil {
		return fmt.Errorf("failed to marshal devices: %v", err)
	}

	resp, err := c.httpClient.Post(
		fmt.Sprintf("%s/devices/sync", c.baseURL),
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return fmt.Errorf("failed to sync devices: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to sync devices: got status %d", resp.StatusCode)
	}
	return nil
}
