package memory

import (
	"context"
	"sync"

	"github.com/wrale/fleet/internal/fleet/device"
)

// Store provides an in-memory implementation of device.Store interface.
// It is primarily used for testing and demonstration purposes.
type Store struct {
	mu      sync.RWMutex
	devices map[string]*device.Device // key: tenantID:deviceID
}

// New creates a new in-memory device store
func New() device.Store {
	return &Store{
		devices: make(map[string]*device.Device),
	}
}

// Create stores a new device
func (s *Store) Create(ctx context.Context, d *device.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.deviceKey(d.TenantID, d.ID)
	if _, exists := s.devices[key]; exists {
		return device.E("Store.Create", device.ErrCodeDeviceExists, "device already exists", nil)
	}

	s.devices[key] = d
	return nil
}

// Get retrieves a device by ID
func (s *Store) Get(ctx context.Context, tenantID, deviceID string) (*device.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.deviceKey(tenantID, deviceID)
	d, exists := s.devices[key]
	if !exists {
		return nil, device.E("Store.Get", device.ErrCodeDeviceNotFound, "device not found", nil)
	}

	return d, nil
}

// Update modifies an existing device
func (s *Store) Update(ctx context.Context, d *device.Device) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.deviceKey(d.TenantID, d.ID)
	if _, exists := s.devices[key]; !exists {
		return device.E("Store.Update", device.ErrCodeDeviceNotFound, "device not found", nil)
	}

	s.devices[key] = d
	return nil
}

// Delete removes a device
func (s *Store) Delete(ctx context.Context, tenantID, deviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.deviceKey(tenantID, deviceID)
	if _, exists := s.devices[key]; !exists {
		return device.E("Store.Delete", device.ErrCodeDeviceNotFound, "device not found", nil)
	}

	delete(s.devices, key)
	return nil
}

// List retrieves devices matching the given options
func (s *Store) List(ctx context.Context, opts device.ListOptions) ([]*device.Device, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*device.Device

	for _, d := range s.devices {
		if opts.TenantID != "" && d.TenantID != opts.TenantID {
			continue
		}

		if opts.Status != "" && d.Status != opts.Status {
			continue
		}

		if len(opts.Tags) > 0 {
			matches := true
			for k, v := range opts.Tags {
				if d.Tags[k] != v {
					matches = false
					break
				}
			}
			if !matches {
				continue
			}
		}

		result = append(result, d)
	}

	// Apply pagination
	if opts.Offset >= len(result) {
		return []*device.Device{}, nil
	}

	end := opts.Offset + opts.Limit
	if end > len(result) {
		end = len(result)
	}

	return result[opts.Offset:end], nil
}

// deviceKey generates a composite key for storing devices
func (s *Store) deviceKey(tenantID, deviceID string) string {
	return tenantID + ":" + deviceID
}
