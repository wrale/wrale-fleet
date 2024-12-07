package resolver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wrale/wrale-fleet/fleet/brain/types"
	synctypes "github.com/wrale/wrale-fleet/fleet/sync/types"
)

func TestResolver(t *testing.T) {
	t.Run("test state resolution", func(t *testing.T) {
		resolver := NewResolver(60) // 60 second timeout
		assert.NotNil(t, resolver)

		now := time.Now()
		// Create test states
		deviceState1 := types.DeviceState{
			Status: "running",
			Metrics: types.DeviceMetrics{
				Temperature: 45.0,
				PowerUsage: 100,
			},
		}
		
		deviceState2 := types.DeviceState{
			Status: "running",
			Metrics: types.DeviceMetrics{
				Temperature: 46.0,
				PowerUsage: 102,
			},
		}

		states := []synctypes.VersionedState{
			{
				Version:   synctypes.StateVersion(1),
				Timestamp: now.Add(-30 * time.Second),
				State:     deviceState1,
			},
			{
				Version:   synctypes.StateVersion(2),
				Timestamp: now,
				State:     deviceState2,
			},
		}

		// Test resolution
		resolved, err := resolver.ResolveStateConflict(states)
		assert.NoError(t, err)
		assert.NotNil(t, resolved)
		assert.Equal(t, synctypes.StateVersion(2), resolved.Version)

		// Test validation
		assert.True(t, resolver.ValidateState(resolved))

		// Test old state validation
		oldState := &synctypes.VersionedState{
			Version:   synctypes.StateVersion(1),
			Timestamp: now.Add(-2 * time.Hour),
			State:     deviceState1,
		}
		assert.False(t, resolver.ValidateState(oldState))
	})

	t.Run("test empty states", func(t *testing.T) {
		resolver := NewResolver(60)
		resolved, err := resolver.ResolveStateConflict([]synctypes.VersionedState{})
		assert.NoError(t, err)
		assert.Nil(t, resolved)
	})
}
