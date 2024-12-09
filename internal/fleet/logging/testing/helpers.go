package testing

import (
	"context"
	"fmt"
	"testing"

	"github.com/wrale/wrale-fleet/internal/fleet/logging"
	"github.com/wrale/wrale-fleet/internal/fleet/logging/store/memory"
	"go.uber.org/zap/zaptest"
)

// NewTestService creates a new logging.Service configured for testing
func NewTestService(t *testing.T) *logging.Service {
	logger := zaptest.NewLogger(t)
	store := memory.New()
	service, err := logging.NewService(store, logger)
	if err != nil {
		t.Fatalf("failed to create test service: %v", err)
	}
	return service
}

// CreateTestEvent creates a new event for testing with the given parameters
func CreateTestEvent(ctx context.Context, s *logging.Service, tenantID string) error {
	return s.Log(ctx, tenantID, logging.EventSystem, logging.LevelInfo, "test event")
}

// CreateTestEvents creates a specified number of test events for a tenant
func CreateTestEvents(ctx context.Context, s *logging.Service, tenantID string, count int) error {
	for i := 0; i < count; i++ {
		if err := s.Log(ctx, tenantID,
			logging.EventSystem,
			logging.LevelInfo,
			fmt.Sprintf("test event %d", i)); err != nil {
			return fmt.Errorf("failed to create test event %d: %w", i, err)
		}
	}
	return nil
}

// SetupMultiTenantTest creates test events across multiple tenants
func SetupMultiTenantTest(ctx context.Context, s *logging.Service, tenants []string, eventsPerTenant int) error {
	for _, tenantID := range tenants {
		if err := CreateTestEvents(ctx, s, tenantID, eventsPerTenant); err != nil {
			return fmt.Errorf("failed to create test events for tenant %s: %w", tenantID, err)
		}
	}
	return nil
}
