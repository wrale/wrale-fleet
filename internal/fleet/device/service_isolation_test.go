package device_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/fleet/internal/fleet/device"
	devicetesting "github.com/wrale/fleet/internal/fleet/device/testing"
)

func TestTenantIsolation(t *testing.T) {
	service := devicetesting.NewTestService(t)

	// Create test tenants
	tenants := []string{"tenant-production", "tenant-staging", "tenant-dev"}
	devices := make(map[string]*device.Device)

	// Setup: Create a device for each tenant
	for _, tenantID := range tenants {
		ctx := devicetesting.ContextWithTestTenant(context.Background(), tenantID)
		device, err := devicetesting.CreateTestDevice(ctx, service, tenantID, "Test Device")
		require.NoError(t, err)
		devices[tenantID] = device

		// Verify device was created with correct tenant
		assert.Equal(t, tenantID, device.TenantID)
		assert.NotEmpty(t, device.ID)
	}

	// Test cases focusing on cross-tenant access attempts
	tests := []struct {
		name          string
		sourceCtx     context.Context // Context with source tenant
		targetTenant  string          // Target tenant we're trying to access
		targetDevice  *device.Device  // Device we're trying to access
		operation     string          // Operation we're attempting
		expectError   bool            // Whether we expect an error
		errorContains string          // Expected error message substring
	}{
		{
			name:         "same tenant access allowed",
			sourceCtx:    devicetesting.ContextWithTestTenant(context.Background(), "tenant-production"),
			targetTenant: "tenant-production",
			targetDevice: devices["tenant-production"],
			operation:    "status_update",
			expectError:  false,
		},
		{
			name:          "cross tenant status update blocked",
			sourceCtx:     devicetesting.ContextWithTestTenant(context.Background(), "tenant-staging"),
			targetTenant:  "tenant-production",
			targetDevice:  devices["tenant-production"],
			operation:     "status_update",
			expectError:   true,
			errorContains: "unauthorized",
		},
		{
			name:          "cross tenant device retrieval blocked",
			sourceCtx:     devicetesting.ContextWithTestTenant(context.Background(), "tenant-dev"),
			targetTenant:  "tenant-production",
			targetDevice:  devices["tenant-production"],
			operation:     "get",
			expectError:   true,
			errorContains: "unauthorized",
		},
		{
			name:          "cross tenant device deletion blocked",
			sourceCtx:     devicetesting.ContextWithTestTenant(context.Background(), "tenant-staging"),
			targetTenant:  "tenant-production",
			targetDevice:  devices["tenant-production"],
			operation:     "delete",
			expectError:   true,
			errorContains: "unauthorized",
		},
		{
			name:          "cross tenant device update blocked",
			sourceCtx:     devicetesting.ContextWithTestTenant(context.Background(), "tenant-dev"),
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
				err = service.UpdateStatus(tt.sourceCtx, tt.targetTenant, tt.targetDevice.ID, device.StatusOnline)
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
						devicetesting.ContextWithTestTenant(context.Background(), tt.targetTenant),
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
	service := devicetesting.NewTestService(t)

	// Create devices for multiple tenants
	tenantDevices := devicetesting.SetupMultiTenantTest(t, service, []string{
		"tenant-production",
		"tenant-staging",
	}, 3)

	// Verify each tenant only sees their own devices
	for tenantID, expectedDevices := range tenantDevices {
		ctx := devicetesting.ContextWithTestTenant(context.Background(), tenantID)

		// Test listing with explicit tenant filter
		devices, err := service.List(ctx, device.ListOptions{TenantID: tenantID})
		assert.NoError(t, err)
		assert.Len(t, devices, len(expectedDevices))
		for _, d := range devices {
			assert.Equal(t, tenantID, d.TenantID)
		}

		// Test listing without tenant filter (should still be isolated)
		devices, err = service.List(ctx, device.ListOptions{})
		assert.NoError(t, err)
		for _, d := range devices {
			assert.Equal(t, tenantID, d.TenantID)
		}

		// Verify we can't see other tenant's devices
		for otherTenant := range tenantDevices {
			if otherTenant == tenantID {
				continue
			}
			devices, err = service.List(ctx, device.ListOptions{TenantID: otherTenant})
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unauthorized")
		}
	}
}
