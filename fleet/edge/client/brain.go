package client

import (
	"github.com/wrale/wrale-fleet/metal/hw/thermal"
)

// BrainClient represents a client connection to the brain
type BrainClient struct {
	// ... existing fields
}

func (c *BrainClient) SyncThermalState(state thermal.ThermalState) error {
	// Implementation
	return nil
}
