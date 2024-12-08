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
		return "", NewError(ErrCodeUnauthorized, "tenant ID not found in context", "device.TenantFromContext")
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
		return NewError(ErrCodeUnauthorized, "unauthorized access to device", "device.ValidateTenantAccess").
			WithField("device_tenant", d.TenantID).
			WithField("context_tenant", tenantID)
	}

	return nil
}

// ValidateTenantMatch ensures two tenant IDs match
func ValidateTenantMatch(tenantID1, tenantID2 string) error {
	if tenantID1 != tenantID2 {
		return NewError(ErrCodeUnauthorized, "tenant ID mismatch", "device.ValidateTenantMatch").
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
