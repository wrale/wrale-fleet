package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wrale/wrale-fleet/fleet/types"
)

func TestBrainClient(t *testing.T) {
	t.Run("test device registration", func(t *testing.T) {
		client := NewBrainClient("http://localhost:8080")
		devices := []types.DeviceID{"device-1"}
		err := client.RegisterDevices(devices)
		assert.NoError(t, err)
	})

	t.Run("test device sync", func(t *testing.T) {
		client := NewBrainClient("http://localhost:8080")
		devices := []types.DeviceID{"device-1"}
		err := client.SyncDevices(devices)
		assert.NoError(t, err)
	})
}
