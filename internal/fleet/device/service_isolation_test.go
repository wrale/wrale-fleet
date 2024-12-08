package device

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestTenantIsolation(t *testing.T) {
	// Create logger that integrates with testing framework
	logger := zaptest.NewLogger(t)
	
	// Initialize service with memory store
	store := newTestStore(t)
	service := NewService(store, logger)

	// Create test tenants
	tenants := []string{"tenant-production", "tenant-staging", "tenant-dev"}
	devices := make(map[string]*Device)

	// Setup: Create a device for each tenant
	for _, tenantID := range tenants {
		ctx := ContextWithTenant(context.Background(), tenantID)
		device, err := service.Register(ctx, tenantID, "Test Device")
		require.NoError(t, err)
		devices[tenantID] = device
	}

	// Test cases focusing on cross-tenant access attempts
	tests := []struct {
		name          string
		sourceCtx     context.Context  // Context with source tenant
		targetTenant  string          // Target tenant we're trying to access
		targetDevice  *Device         // Device we're trying to access
		operation     string          // Operation we're attempting
		expectError   bool            // Whether we expect an error
		errorContains string          // Expected error message substring
	}{
		{
			name:          "same tenant access allowed",
			sourceCtx:     ContextWithTenant(context.Background(), "tenant-production"),
			targetTenant:  "tenant-production",
			targetDevice:  devices["tenant-production"],
			operation:     "status_update",
			expectError:   false,
		},
		{
			name:          "cross tenant status update blocked",
			sourceCtx:     ContextWithTenant(context.Background(), "tenant-staging"),
			targetTenant:  "tenant-production",
			targetDevice:  devices["tenant-production"],
			operation:     "status_update",
			expectError:   true,
			errorContains: "unauthorized",
		},
		{
			name:          "cross tenant device retrieval blocked",
			sourceCtx:     ContextWithTenant(context.Background(), "tenant-dev"),
			targetTenant:  "tenant-production",
			targetDevice:  devices["tenant-production"],
			operation:     "get",
			expectError:   true,
			errorContains: "unauthorized",
		},
		{
			name:          "cross tenant device deletion blocked",
			sourceCtx:     ContextWithTenant(context.Background(), "tenant-staging"),
			targetTenant:  "tenant-production",
			targetDevice:  devices["tenant-production"],
			operation:     "delete",
			expectError:   true,
			errorContains: "unauthorized",
		},
		{
			name:          "cross tenant device update blocked",
			sourceCtx:     ContextWithTenant(context.Background(), "tenant-dev"),
			targetTenant:  "tenant-production",
			targetDevice:  devices["tenant-production"],
			operation:     "update",
			expectError:   true,
			errorContains: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			// Attempt the cross-tenant operation
			switch tt.operation {
			case "status_update":
				err = service.UpdateStatus(tt.sourceCtx, tt.targetTenant, tt.targetDevice.ID, StatusOnline)
			case "get":
				_, err = service.Get(tt.sourceCtx, tt.targetTenant, tt.targetDevice.ID)
			case "delete":
				err = service.Delete(tt.sourceCtx, tt.targetTenant, tt.targetDevice.ID)
			case "update":
				tt.targetDevice.UpdatedAt = time.Now()
				err = service.Update(tt.sourceCtx, tt.targetDevice)
			}

			// Verify expected outcome
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTenantListIsolation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := newTestStore(t)
	service := NewService(store, logger)

	// Create devices for multiple tenants
	tenants := []string{"tenant-production", "tenant-staging"}
	devicesPerTenant := 3
	
	for _, tenantID := range tenants {
		ctx := ContextWithTenant(context.Background(), tenantID)
		for i := 0; i < devicesPerTenant; i++ {
			_, err := service.Register(ctx, tenantID, "Test Device")
			require.NoError(t, err)
		}
	}

	// Verify each tenant only sees their own devices
	for _, tenantID := range tenants {
		ctx := ContextWithTenant(context.Background(), tenantID)
		
		// Test listing with explicit tenant filter
		devices, err := service.List(ctx, ListOptions{TenantID: tenantID})
		assert.NoError(t, err)
		assert.Len(t, devices, devicesPerTenant)
		for _, device := range devices {
			assert.Equal(t, tenantID, device.TenantID)
		}

		// Test listing without tenant filter (should still be isolated)
		devices, err = service.List(ctx, ListOptions{})
		assert.NoError(t, err)
		for _, device := range devices {
			assert.Equal(t, tenantID, device.TenantID)
		}
	}
}

// Helper to create a test store
func newTestStore(t *testing.T) Store {
	// You could return a memory store implementation here
	// For now we'll return nil to make the test fail and remind us to implement it
	t.Helper()
	return nil
}