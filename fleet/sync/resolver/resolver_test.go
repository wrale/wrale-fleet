package resolver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wrale/wrale-fleet/fleet/sync/types"
)

func TestResolver(t *testing.T) {
	t.Run("test state resolution", func(t *testing.T) {
		resolver := NewResolver(60) // 60 second timeout
		assert.NotNil(t, resolver)

		// Create test states
		states := []types.VersionedState{
			{
				Version:   1,
				Timestamp: time.Now().Add(-30 * time.Second).Unix(),
				State:     map[string]interface{}{"key": "value1"},
			},
			{
				Version:   2,
				Timestamp: time.Now().Unix(),
				State:     map[string]interface{}{"key": "value2"},
			},
		}

		// Test resolution
		resolved, err := resolver.ResolveStateConflict(states)
		assert.NoError(t, err)
		assert.NotNil(t, resolved)
		assert.Equal(t, int64(2), resolved.Version)

		// Test validation
		assert.True(t, resolver.ValidateState(resolved))

		// Test old state validation
		oldState := &types.VersionedState{
			Version:   1,
			Timestamp: time.Now().Add(-2 * time.Hour).Unix(),
			State:     map[string]interface{}{"key": "old"},
		}
		assert.False(t, resolver.ValidateState(oldState))
	})

	t.Run("test empty states", func(t *testing.T) {
		resolver := NewResolver(60)
		resolved, err := resolver.ResolveStateConflict([]types.VersionedState{})
		assert.NoError(t, err)
		assert.Nil(t, resolved)
	})
}
