package client

import (
	"fmt"
	"github.com/wrale/wrale-fleet/metal/hw/power"
)

// MetalClient provides interface to metal layer functionality
type MetalClient struct {
	powerManager *power.Manager
}

// NewMetalClient creates a new metal layer client
func NewMetalClient(powerMgr *power.Manager) *MetalClient {
	return &MetalClient{
		powerManager: powerMgr,
	}
}

// GetPowerState retrieves current power system state
func (c *MetalClient) GetPowerState() (*power.State, error) {
	if c.powerManager == nil {
		return nil, fmt.Errorf("power manager not initialized")
	}
	return c.powerManager.GetState(), nil
}
