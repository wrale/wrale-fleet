package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/fleet/internal/fleet/device"
)

func TestNew(t *testing.T) {
	store := New()
	// Test store initialization by creating and retrieving a device
	dev := &device.Device{
		ID:       "test-init",
		TenantID: "tenant-init",
		Name:     "Test Init Device",
	}
	ctx := context.Background()
	err := store.Create(ctx, dev)
	require.NoError(t, err)

	retrieved, err := store.Get(ctx, dev.TenantID, dev.ID)
	require.NoError(t, err)
	assert.Equal(t, dev.ID, retrieved.ID)
}

func TestStore_Create(t *testing.T) {
	tests := []struct {
		name    string
		device  *device.Device
		wantErr bool
	}{
		{
			name: "valid device",
			device: &device.Device{
				ID:       "test-1",
				TenantID: "tenant-1",
				Name:     "Test Device",
			},
			wantErr: false,
		},
		{
			name: "duplicate device",
			device: &device.Device{
				ID:       "test-1",
				TenantID: "tenant-1",
				Name:     "Test Device",
			},
			wantErr: true,
		},
		{
			name: "missing required fields",
			device: &device.Device{
				Name: "Invalid Device",
			},
			wantErr: true,
		},
	}

	store := New()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Create(ctx, tt.device)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify device was stored correctly
			stored, err := store.Get(ctx, tt.device.TenantID, tt.device.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.device.Name, stored.Name)
		})
	}
}

func TestStore_Get(t *testing.T) {
	store := New()
	ctx := context.Background()

	// Create test device
	device := &device.Device{
		ID:       "test-1",
		TenantID: "tenant-1",
		Name:     "Test Device",
	}
	require.NoError(t, store.Create(ctx, device))

	tests := []struct {
		name     string
		tenantID string
		deviceID string
		wantErr  bool
	}{
		{
			name:     "existing device",
			tenantID: "tenant-1",
			deviceID: "test-1",
			wantErr:  false,
		},
		{
			name:     "wrong tenant",
			tenantID: "wrong-tenant",
			deviceID: "test-1",
			wantErr:  true,
		},
		{
			name:     "non-existent device",
			tenantID: "tenant-1",
			deviceID: "missing",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.Get(ctx, tt.tenantID, tt.deviceID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.deviceID, got.ID)
			assert.Equal(t, tt.tenantID, got.TenantID)
		})
	}
}

func TestStore_Update(t *testing.T) {
	store := New()
	ctx := context.Background()

	// Create initial device
	initial := &device.Device{
		ID:       "test-1",
		TenantID: "tenant-1",
		Name:     "Initial Name",
	}
	require.NoError(t, store.Create(ctx, initial))

	tests := []struct {
		name    string
		device  *device.Device
		wantErr bool
	}{
		{
			name: "valid update",
			device: &device.Device{
				ID:       "test-1",
				TenantID: "tenant-1",
				Name:     "Updated Name",
			},
			wantErr: false,
		},
		{
			name: "non-existent device",
			device: &device.Device{
				ID:       "missing",
				TenantID: "tenant-1",
				Name:     "Missing Device",
			},
			wantErr: true,
		},
		{
			name: "wrong tenant",
			device: &device.Device{
				ID:       "test-1",
				TenantID: "wrong-tenant",
				Name:     "Wrong Tenant",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Update(ctx, tt.device)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			// Verify update
			updated, err := store.Get(ctx, tt.device.TenantID, tt.device.ID)
			require.NoError(t, err)
			assert.Equal(t, tt.device.Name, updated.Name)
		})
	}
}

func TestStore_Delete(t *testing.T) {
	store := New()
	ctx := context.Background()

	// Create test device
	device := &device.Device{
		ID:       "test-1",
		TenantID: "tenant-1",
		Name:     "Test Device",
	}
	require.NoError(t, store.Create(ctx, device))

	tests := []struct {
		name     string
		tenantID string
		deviceID string
		wantErr  bool
	}{
		{
			name:     "existing device",
			tenantID: "tenant-1",
			deviceID: "test-1",
			wantErr:  false,
		},
		{
			name:     "non-existent device",
			tenantID: "tenant-1",
			deviceID: "missing",
			wantErr:  true,
		},
		{
			name:     "wrong tenant",
			tenantID: "wrong-tenant",
			deviceID: "test-1",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := store.Delete(ctx, tt.tenantID, tt.deviceID)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			// Verify deletion
			_, err = store.Get(ctx, tt.tenantID, tt.deviceID)
			require.Error(t, err)
		})
	}
}

func TestStore_List(t *testing.T) {
	store := New()
	ctx := context.Background()

	// Create test devices
	devices := []*device.Device{
		{
			ID:       "dev-1",
			TenantID: "tenant-1",
			Name:     "Device 1",
			Status:   device.StatusOnline,
			Tags:     map[string]string{"env": "prod"},
		},
		{
			ID:       "dev-2",
			TenantID: "tenant-1",
			Name:     "Device 2",
			Status:   device.StatusOffline,
			Tags:     map[string]string{"env": "staging"},
		},
		{
			ID:       "dev-3",
			TenantID: "tenant-2",
			Name:     "Device 3",
			Status:   device.StatusOnline,
		},
	}

	for _, d := range devices {
		require.NoError(t, store.Create(ctx, d))
	}

	tests := []struct {
		name    string
		opts    device.ListOptions
		want    int
		wantIDs []string
	}{
		{
			name:    "list all devices",
			opts:    device.ListOptions{},
			want:    3,
			wantIDs: []string{"dev-1", "dev-2", "dev-3"},
		},
		{
			name: "filter by tenant",
			opts: device.ListOptions{
				TenantID: "tenant-1",
			},
			want:    2,
			wantIDs: []string{"dev-1", "dev-2"},
		},
		{
			name: "filter by status",
			opts: device.ListOptions{
				Status: device.StatusOnline,
			},
			want:    2,
			wantIDs: []string{"dev-1", "dev-3"},
		},
		{
			name: "filter by tags",
			opts: device.ListOptions{
				Tags: map[string]string{"env": "prod"},
			},
			want:    1,
			wantIDs: []string{"dev-1"},
		},
		{
			name: "pagination",
			opts: device.ListOptions{
				Offset: 1,
				Limit:  1,
			},
			want:    1,
			wantIDs: []string{"dev-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.List(ctx, tt.opts)
			require.NoError(t, err)
			assert.Len(t, got, tt.want)

			if tt.wantIDs != nil {
				var gotIDs []string
				for _, d := range got {
					gotIDs = append(gotIDs, d.ID)
				}
				assert.ElementsMatch(t, tt.wantIDs, gotIDs)
			}
		})
	}
}

func TestStore_Concurrency(t *testing.T) {
	store := New()
	ctx := context.Background()

	// Create initial device
	dev := &device.Device{
		ID:       "test-1",
		TenantID: "tenant-1",
		Name:     "Test Device",
	}
	require.NoError(t, store.Create(ctx, dev))

	// Test concurrent operations
	var wg sync.WaitGroup
	concurrentOps := 100

	// Test concurrent reads
	wg.Add(concurrentOps)
	for i := 0; i < concurrentOps; i++ {
		go func() {
			defer wg.Done()
			_, _ = store.Get(ctx, "tenant-1", "test-1")
		}()
	}

	// Test concurrent updates
	wg.Add(concurrentOps)
	for i := 0; i < concurrentOps; i++ {
		go func(i int) {
			defer wg.Done()
			device := &device.Device{
				ID:       "test-1",
				TenantID: "tenant-1",
				Name:     fmt.Sprintf("Updated Name %d", i),
			}
			_ = store.Update(ctx, device)
		}(i)
	}

	// Add timeout to prevent test hanging
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - no deadlocks or panics
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for concurrent operations")
	}

	// Verify device is still accessible
	stored, err := store.Get(ctx, "tenant-1", "test-1")
	require.NoError(t, err)
	assert.NotNil(t, stored)
}
