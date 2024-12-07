package store

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	t.Run("test store operations", func(t *testing.T) {
		store := NewStore()
		assert.NotNil(t, store)

		err := store.Set("test-key", "test-value")
		assert.NoError(t, err)

		value, err := store.Get("test-key")
		assert.NoError(t, err)
		assert.Equal(t, "test-value", value)

		fmt.Printf("Store test completed successfully\n")
	})
}
