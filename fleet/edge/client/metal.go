package client

import (
	"github.com/wrale/wrale-fleet/metal/hw/thermal"
)

// MetalClient represents a client connection to the metal layer
type MetalClient struct {
	// ... existing fields
}

func (c *MetalClient) GetThermalState() (thermal.ThermalState, error) {
	var state thermal.ThermalState
	// Implementation
	return state, nil
}