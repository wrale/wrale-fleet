package client

import (
	"net/http"
	"github.com/wrale/wrale-fleet/metal/hw/thermal"
)

// MetalClient represents a client connection to the metal layer
type MetalClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewMetalClient creates a new metal client with the given base URL
func NewMetalClient(baseURL string) *MetalClient {
	return &MetalClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *MetalClient) GetThermalState() (thermal.ThermalState, error) {
	var state thermal.ThermalState
	// Implementation
	return state, nil
}
