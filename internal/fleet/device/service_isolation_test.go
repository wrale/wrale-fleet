package device

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/fleet/internal/fleet/device/store/memory"
	"go.uber.org/zap/zaptest"
)

func TestTenantIsolation(t *testing.T) {
	// Create logger that integrates with testing framework
	logger := zaptest.NewLogger(t)
	
	// Initialize service with memory store
	store := memory.New()
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
		
		// Verify device was created with correct tenant
		assert.Equal(t, tenantID, device.TenantID)
		assert.NotEmpty(t, device.ID)
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
		{
			// This test specifically targets the scenario seen in make run
			name:          "staging to production status update blocked",
			sourceCtx:     ContextWithTenant(context.Background(), "tenant-staging"),
			targetTenant:  "tenant-production",
			targetDevice:  devices["tenant-production"],
			operation:     "status_update",
			expectError:   true,
			errorContains: "unauthorized",
		},
		{
			// Additional test for the reverse direction
			name:          "production to staging status update blocked",
			sourceCtx:     ContextWithTenant(context.Background(), "tenant-production"),
			targetTenant:  "tenant-staging",
			targetDevice:  devices["tenant-staging"],
			operation:     "status_update",
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
				deviceCopy := *tt.targetDevice // Create copy to avoid modifying original
				deviceCopy.UpdatedAt = time.Now()
				err = service.Update(tt.sourceCtx, &deviceCopy)
			}

			// Verify expected outcome
			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				
				// Additional verification that device state didn't change
				if tt.operation == "status_update" {
					device, getErr := service.Get(
						ContextWithTenant(context.Background(), tt.targetTenant),
						tt.targetTenant,
						tt.targetDevice.ID,
					)
					require.NoError(t, getErr)
					assert.Equal(t, tt.targetDevice.Status, device.Status)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTenantListIsolation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	store := memory.New()
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

		// Verify we can't see other tenant's devices
		otherTenants := make([]string, 0)
		for _, t := range tenants {
			if t != tenantID {
				otherTenants = append(otherTenants, t)
			}
		}
		for _, otherTenant := range otherTenants {
			devices, err = service.List(ctx, ListOptions{TenantID: otherTenant})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unauthorized")
		}
	}
}