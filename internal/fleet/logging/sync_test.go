package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestSync(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Test normal sync
	assert.NoError(t, Sync(logger))

	// Test sync with nil logger
	assert.NoError(t, Sync(nil))
}

func TestMustSync(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// Test normal must sync
	assert.NotPanics(t, func() {
		MustSync(logger)
	})

	// Test must sync with nil logger
	assert.NotPanics(t, func() {
		MustSync(nil)
	})
}
