// Original content...
package group

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	devmem "github.com/wrale/fleet/internal/fleet/device/store/memory"
	"github.com/wrale/fleet/internal/fleet/group/store/memory"
)

func newTestStore() Store {
	deviceStore := devmem.New()
	return memory.New(deviceStore)
}

// Rest of the file unchanged...
