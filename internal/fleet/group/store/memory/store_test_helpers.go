package memory

import (
	"testing"

	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/fleet/device/store/memory"
	"github.com/wrale/fleet/internal/fleet/group"
)

// testSetup encapsulates test dependencies
type testSetup struct {
	store       *Store
	deviceStore device.Store
}

// setupTest creates a new test environment
func setupTest(t *testing.T) *testSetup {
	deviceStore := memory.New()
	groupStore := New(deviceStore)
	return &testSetup{
		store:       groupStore,
		deviceStore: deviceStore,
	}
}

// createTestGroup creates a group for testing
func (ts *testSetup) createTestGroup(id, tenantID, name string, groupType group.Type) *group.Group {
	return &group.Group{
		ID:       id,
		TenantID: tenantID,
		Name:     name,
		Type:     groupType,
		Properties: group.Properties{
			Metadata: make(map[string]string),
		},
	}
}

// createTestDevice creates a device for testing
func (ts *testSetup) createTestDevice(id, tenantID, name string) *device.Device {
	return &device.Device{
		ID:       id,
		TenantID: tenantID,
		Name:     name,
		Tags:     make(map[string]string),
	}
}

// createTestDeviceWithTags creates a device with specified tags for testing
func (ts *testSetup) createTestDeviceWithTags(id, tenantID, name string, tags map[string]string) *device.Device {
	d := ts.createTestDevice(id, tenantID, name)
	d.Tags = tags
	return d
}

// createTestDynamicGroup creates a dynamic group with query for testing
func (ts *testSetup) createTestDynamicGroup(id, tenantID, name string, query *group.MembershipQuery) *group.Group {
	g := ts.createTestGroup(id, tenantID, name, group.TypeDynamic)
	g.Query = query
	return g
}
