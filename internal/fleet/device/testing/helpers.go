package testing

import (
	"context"
	"fmt"
	"testing"

	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/fleet/device/store/memory"
	"go.uber.org/zap/zaptest"
)

// NewTestService creates a new device.Service configured for testing
func NewTestService(t *testing.T) *device.Service {
	logger := zaptest.NewLogger(t)
	store := memory.New()
	return device.NewService(store, logger)
}

// CreateTestDevice creates a new device for testing with the given tenant ID
func CreateTestDevice(ctx context.Context, s *device.Service, tenantID string, name string) (*device.Device, error) {
	return s.Register(ctx, tenantID, name)
}

// ContextWithTestTenant creates a new context with the specified tenant ID
func ContextWithTestTenant(ctx context.Context, tenantID string) context.Context {
	return device.ContextWithTenant(ctx, tenantID)
}

// CreateTestTenantDevices creates a specified number of test devices for a tenant
func CreateTestTenantDevices(ctx context.Context, s *device.Service, tenantID string, count int) ([]*device.Device, error) {
	devices := make([]*device.Device, 0, count)

	for i := 0; i < count; i++ {
		d, err := CreateTestDevice(ctx, s, tenantID, fmt.Sprintf("test-device-%d", i))
		if err != nil {
			return nil, fmt.Errorf("failed to create test device %d: %w", i, err)
		}
		devices = append(devices, d)
	}

	return devices, nil
}

// SetupMultiTenantTest creates test devices across multiple tenants
func SetupMultiTenantTest(t *testing.T, s *device.Service, tenants []string, devicesPerTenant int) map[string][]*device.Device {
	result := make(map[string][]*device.Device)

	for _, tenantID := range tenants {
		ctx := ContextWithTestTenant(context.Background(), tenantID)
		devices, err := CreateTestTenantDevices(ctx, s, tenantID, devicesPerTenant)
		if err != nil {
			t.Fatalf("failed to create test devices for tenant %s: %v", tenantID, err)
		}
		result[tenantID] = devices
	}

	return result
}
