package device

import (
	"context"
	"fmt"
)

// contextKey is a private type for context keys to prevent collisions
type contextKey int

const (
	// tenantIDKey is the context key for tenant ID
	tenantIDKey contextKey = iota
)

// ContextWithTenant adds tenant ID to the context
func ContextWithTenant(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// TenantFromContext extracts tenant ID from context
func TenantFromContext(ctx context.Context) (string, error) {
	tenantID, ok := ctx.Value(tenantIDKey).(string)
	if !ok || tenantID == "" {
		return "", E("device.TenantFromContext", ErrCodeUnauthorized, "tenant ID not found in context", nil)
	}
	return tenantID, nil
}

// ValidateTenantAccess checks if the context tenant matches the device tenant
func ValidateTenantAccess(ctx context.Context, d *Device) error {
	tenantID, err := TenantFromContext(ctx)
	if err != nil {
		return err
	}

	if d.TenantID != tenantID {
		return E("device.ValidateTenantAccess", ErrCodeUnauthorized, "unauthorized access to device", nil).
			WithField("device_tenant", d.TenantID).
			WithField("context_tenant", tenantID)
	}

	return nil
}

// ValidateTenantMatch ensures two tenant IDs match
func ValidateTenantMatch(tenantID1, tenantID2 string) error {
	if tenantID1 != tenantID2 {
		return E("device.ValidateTenantMatch", ErrCodeUnauthorized, "tenant ID mismatch", nil).
			WithField("tenant1", tenantID1).
			WithField("tenant2", tenantID2)
	}
	return nil
}

// EnsureTenant validates tenant ID presence in context
func EnsureTenant(ctx context.Context) error {
	if _, err := TenantFromContext(ctx); err != nil {
		return fmt.Errorf("tenant validation failed: %w", err)
	}
	return nil
}