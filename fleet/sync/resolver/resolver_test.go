package resolver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wrale/wrale-fleet/fleet/sync/types"
)

func TestResolver(t *testing.T) {
	// Test implementation using sync/types.VersionedState and types.StateChange
	t.Run("test state resolution", func(t *testing.T) {
		resolver := NewResolver()
		assert.NotNil(t, resolver)
	})
}