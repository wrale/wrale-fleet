package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/wrale/fleet/internal/fleet/device"
)

// DeviceStore provides an in-memory implementation of device.Store
type DeviceStore struct {
	mu      sync.RWMutex
	devices map[string]*device.Device // key: tenantID:deviceID
}

// NewDeviceStore creates a new in-memory device store
func NewDeviceStore() *DeviceStore {
	return &DeviceStore{
		devices: make(map[string]*device.Device),
	}
}

// key generates the map key for a device
func (s *DeviceStore) key(tenantID, deviceID string) string {
	return fmt.Sprintf("%s:%s", tenantID, deviceID)
}

// Create stores a new device
func (s *DeviceStore) Create(ctx context.Context, d *device.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.key(d.TenantID, d.ID)
	if _, exists := s.devices[key]; exists {
		return device.ErrDeviceExists
	}

	// Store a copy to prevent external modifications
	copy := *d
	s.devices[key] = &copy

	return nil
}

// Get retrieves a device by ID
func (s *DeviceStore) Get(ctx context.Context, tenantID, deviceID string) (*device.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.key(tenantID, deviceID)
	d, exists := s.devices[key]
	if !exists {
		return nil, device.ErrDeviceNotFound
	}

	// Return a copy to prevent external modifications
	copy := *d
	return &copy, nil
}

// Update modifies an existing device
func (s *DeviceStore) Update(ctx context.Context, d *device.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.key(d.TenantID, d.ID)
	if _, exists := s.devices[key]; !exists {
		return device.ErrDeviceNotFound
	}

	// Store a copy to prevent external modifications
	copy := *d
	s.devices[key] = &copy

	return nil
}

// Delete removes a device
func (s *DeviceStore) Delete(ctx context.Context, tenantID, deviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.key(tenantID, deviceID)
	if _, exists := s.devices[key]; !exists {
		return device.ErrDeviceNotFound
	}

	delete(s.devices, key)
	return nil
}

// List retrieves devices matching the given options
func (s *DeviceStore) List(ctx context.Context, opts device.ListOptions) ([]*device.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*device.Device

	for _, d := range s.devices {
		if !s.matchesFilter(d, opts) {
			continue
		}

		// Add a copy to prevent external modifications
		copy := *d
		result = append(result, &copy)
	}

	// Apply pagination if specified
	if opts.Limit > 0 {
		start := opts.Offset
		if start > len(result) {
			start = len(result)
		}
		end := start + opts.Limit
		if end > len(result) {
			end = len(result)
		}
		result = result[start:end]
	}

	return result, nil
}

// matchesFilter checks if a device matches the filter criteria
func (s *DeviceStore) matchesFilter(d *device.Device, opts device.ListOptions) bool {
	if opts.TenantID != "" && d.TenantID != opts.TenantID {
		return false
	}

	if opts.Status != "" && d.Status != opts.Status {
		return false
	}

	// Check if all required tags are present with matching values
	for key, value := range opts.Tags {
		if d.Tags[key] != value {
			return false
		}
	}

	return true
}
