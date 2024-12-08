package device

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Test creating a new device
	tenantID := "test-tenant"
	name := "test-device"

	device := New(tenantID, name)

	require.NotEmpty(t, device.ID, "device ID should be generated")
	require.Equal(t, tenantID, device.TenantID, "tenant ID should match")
	require.Equal(t, name, device.Name, "device name should match")
	assert.Equal(t, StatusUnknown, device.Status, "initial status should be unknown")
	assert.NotNil(t, device.Tags, "tags map should be initialized")
	assert.False(t, device.CreatedAt.IsZero(), "created timestamp should be set")
	assert.False(t, device.UpdatedAt.IsZero(), "updated timestamp should be set")
}

func TestDevice_Validate(t *testing.T) {
	tests := []struct {
		name    string
		device  *Device
		wantErr bool
	}{
		{
			name: "valid device",
			device: &Device{
				ID:       "test-id",
				TenantID: "test-tenant",
				Name:     "test-device",
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			device: &Device{
				TenantID: "test-tenant",
				Name:     "test-device",
			},
			wantErr: true,
		},
		{
			name: "missing tenant ID",
			device: &Device{
				ID:   "test-id",
				Name: "test-device",
			},
			wantErr: true,
		},
		{
			name: "missing name",
			device: &Device{
				ID:       "test-id",
				TenantID: "test-tenant",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.device.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDevice_SetStatus(t *testing.T) {
	device := New("test-tenant", "test-device")
	originalUpdate := device.UpdatedAt

	// Wait briefly to ensure timestamp changes
	time.Sleep(time.Millisecond)

	device.SetStatus(StatusOnline)

	assert.Equal(t, StatusOnline, device.Status, "status should be updated")
	assert.True(t, device.UpdatedAt.After(originalUpdate), "updated timestamp should be newer")
}

func TestDevice_SetConfig(t *testing.T) {
	device := New("test-tenant", "test-device")
	originalUpdate := device.UpdatedAt

	config := json.RawMessage(`{"key": "value"}`)
	device.SetConfig(config)

	assert.Equal(t, config, device.Config, "config should be updated")
	assert.True(t, device.UpdatedAt.After(originalUpdate), "updated timestamp should be newer")
}

func TestDevice_Tags(t *testing.T) {
	device := New("test-tenant", "test-device")

	// Test adding tags
	device.AddTag("env", "prod")
	assert.Equal(t, "prod", device.Tags["env"], "tag should be added")

	// Test updating existing tag
	device.AddTag("env", "dev")
	assert.Equal(t, "dev", device.Tags["env"], "tag should be updated")

	// Test removing tag
	device.RemoveTag("env")
	_, exists := device.Tags["env"]
	assert.False(t, exists, "tag should be removed")
}
