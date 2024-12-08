package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/fleet/internal/fleet/group"
)

func TestStore_AddDevice(t *testing.T) {
	ts := setupTest(t)
	ctx := context.Background()

	// Create test group
	testGroup := ts.createTestGroup("test-group", "tenant-1", "Test Group", group.TypeStatic)
	require.NoError(t, ts.store.Create(ctx, testGroup))

	// Create test device
	testDevice := ts.createTestDevice("test-device", "tenant-1", "Test Device")
	require.NoError(t, ts.deviceStore.Create(ctx, testDevice))

	tests := []struct {
		name      string
		tenantID  string
		groupID   string
		device    *device.Device
		wantErr   bool
		errCode   string
		deviceCnt int
	}{
		{
			name:      "add device to static group",
			tenantID:  "tenant-1",
			groupID:   "test-group",
			device:    testDevice,
			wantErr:   false,
			deviceCnt: 1,
		},
		{
			name:     "add device to non-existent group",
			tenantID: "tenant-1",
			groupID:  "missing-group",
			device:   testDevice,
			wantErr:  true,
		},
		{
			name:     "add device with wrong tenant",
			tenantID: "tenant-1",
			groupID:  "test-group",
			device:   ts.createTestDevice("wrong-tenant-device", "tenant-2", "Wrong Tenant Device"),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ts.store.AddDevice(ctx, tt.tenantID, tt.groupID, tt.device)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					assert.Equal(t, tt.errCode, err.(*group.Error).Code)
				}
				return
			}

			require.NoError(t, err)

			// Verify device count
			g, err := ts.store.Get(ctx, tt.tenantID, tt.groupID)
			require.NoError(t, err)
			assert.Equal(t, tt.deviceCnt, g.DeviceCount)

			// Verify device membership
			devices, err := ts.store.ListDevices(ctx, tt.tenantID, tt.groupID)
			require.NoError(t, err)
			assert.Len(t, devices, tt.deviceCnt)
		})
	}
}

func TestStore_RemoveDevice(t *testing.T) {
	ts := setupTest(t)
	ctx := context.Background()

	// Create and set up test group with device
	testGroup := ts.createTestGroup("test-group", "tenant-1", "Test Group", group.TypeStatic)
	require.NoError(t, ts.store.Create(ctx, testGroup))

	testDevice := ts.createTestDevice("test-device", "tenant-1", "Test Device")
	require.NoError(t, ts.deviceStore.Create(ctx, testDevice))
	require.NoError(t, ts.store.AddDevice(ctx, "tenant-1", "test-group", testDevice))

	tests := []struct {
		name      string
		tenantID  string
		groupID   string
		deviceID  string
		wantErr   bool
		deviceCnt int
	}{
		{
			name:      "remove existing device",
			tenantID:  "tenant-1",
			groupID:   "test-group",
			deviceID:  "test-device",
			wantErr:   false,
			deviceCnt: 0,
		},
		{
			name:     "remove from non-existent group",
			tenantID: "tenant-1",
			groupID:  "missing-group",
			deviceID: "test-device",
			wantErr:  true,
		},
		{
			name:     "remove non-existent device",
			tenantID: "tenant-1",
			groupID:  "test-group",
			deviceID: "missing-device",
			wantErr:  false, // Should not error as it's idempotent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ts.store.RemoveDevice(ctx, tt.tenantID, tt.groupID, tt.deviceID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify device count
			if !tt.wantErr {
				g, err := ts.store.Get(ctx, tt.tenantID, tt.groupID)
				require.NoError(t, err)
				assert.Equal(t, tt.deviceCnt, g.DeviceCount)
			}
		})
	}
}

func TestStore_ListDevices(t *testing.T) {
	ts := setupTest(t)
	ctx := context.Background()

	// Create and set up static group
	staticGroup := ts.createTestGroup("static-group", "tenant-1", "Static Group", group.TypeStatic)
	require.NoError(t, ts.store.Create(ctx, staticGroup))

	// Create dynamic group with tag query
	dynamicGroup := ts.createTestDynamicGroup("dynamic-group", "tenant-1", "Dynamic Group",
		&group.MembershipQuery{
			Tags: map[string]string{"env": "prod"},
		})
	require.NoError(t, ts.store.Create(ctx, dynamicGroup))

	// Create test devices
	devices := []*device.Device{
		ts.createTestDeviceWithTags("device-1", "tenant-1", "Device 1", map[string]string{"env": "prod"}),
		ts.createTestDeviceWithTags("device-2", "tenant-1", "Device 2", map[string]string{"env": "staging"}),
	}

	for _, d := range devices {
		require.NoError(t, ts.deviceStore.Create(ctx, d))
	}

	// Add device to static group
	require.NoError(t, ts.store.AddDevice(ctx, "tenant-1", "static-group", devices[0]))

	tests := []struct {
		name     string
		tenantID string
		groupID  string
		want     int
		wantIDs  []string
		wantErr  bool
	}{
		{
			name:     "list devices in static group",
			tenantID: "tenant-1",
			groupID:  "static-group",
			want:     1,
			wantIDs:  []string{"device-1"},
		},
		{
			name:     "list devices in dynamic group",
			tenantID: "tenant-1",
			groupID:  "dynamic-group",
			want:     1,
			wantIDs:  []string{"device-1"},
		},
		{
			name:     "list devices in non-existent group",
			tenantID: "tenant-1",
			groupID:  "missing",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devices, err := ts.store.ListDevices(ctx, tt.tenantID, tt.groupID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, devices, tt.want)

			if tt.wantIDs != nil {
				var gotIDs []string
				for _, d := range devices {
					gotIDs = append(gotIDs, d.ID)
				}
				assert.ElementsMatch(t, tt.wantIDs, gotIDs)
			}
		})
	}
}

func TestDynamicGroupMembership(t *testing.T) {
	ts := setupTest(t)
	ctx := context.Background()

	// Create dynamic group that matches devices by multiple criteria
	dynamicGroup := ts.createTestDynamicGroup("dynamic-group", "tenant-1", "Dynamic Group",
		&group.MembershipQuery{
			Tags:    map[string]string{"env": "prod", "region": "us-west"},
			Status:  device.StatusOnline,
			Regions: []string{"us-west-1", "us-west-2"},
		})
	require.NoError(t, ts.store.Create(ctx, dynamicGroup))

	// Create test devices with various attributes
	devices := []*device.Device{
		ts.createTestDeviceWithTags("device-1", "tenant-1", "Device 1",
			map[string]string{"env": "prod", "region": "us-west"}),
		ts.createTestDeviceWithTags("device-2", "tenant-1", "Device 2",
			map[string]string{"env": "prod", "region": "us-east"}),
		ts.createTestDeviceWithTags("device-3", "tenant-1", "Device 3",
			map[string]string{"env": "staging", "region": "us-west"}),
	}

	// Set status for matching device
	devices[0].Status = device.StatusOnline
	devices[1].Status = device.StatusOnline
	devices[2].Status = device.StatusOffline

	for _, d := range devices {
		require.NoError(t, ts.deviceStore.Create(ctx, d))
	}

	// List devices in dynamic group
	groupDevices, err := ts.store.ListDevices(ctx, "tenant-1", "dynamic-group")
	require.NoError(t, err)

	// Should only match device-1 which meets all criteria
	assert.Len(t, groupDevices, 1)
	if len(groupDevices) > 0 {
		assert.Equal(t, "device-1", groupDevices[0].ID)
	}

	// Verify device count was updated
	g, err := ts.store.Get(ctx, "tenant-1", "dynamic-group")
	require.NoError(t, err)
	assert.Equal(t, 1, g.DeviceCount)
}
