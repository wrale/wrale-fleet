package device

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Status represents the current state of a device
type Status string

const (
	StatusUnknown     Status = "unknown"
	StatusOnline      Status = "online"
	StatusOffline     Status = "offline"
	StatusError       Status = "error"
	StatusMaintenance Status = "maintenance"
)

// DiscoveryMethod represents how a device was discovered
type DiscoveryMethod string

const (
	DiscoveryManual    DiscoveryMethod = "manual"
	DiscoveryAutomatic DiscoveryMethod = "automatic"
	DiscoveryMDNS      DiscoveryMethod = "mdns"
	DiscoveryScan      DiscoveryMethod = "network_scan"
)

// NetworkInfo contains device network-related information
type NetworkInfo struct {
	IPAddress  string            `json:"ip_address,omitempty"`
	MACAddress string            `json:"mac_address,omitempty"`
	Hostname   string            `json:"hostname,omitempty"`
	Port       int               `json:"port,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// Device represents a managed Raspberry Pi device in the fleet
type Device struct {
	ID              string            `json:"id"`
	TenantID        string            `json:"tenant_id"`
	Name            string            `json:"name"`
	Status          Status            `json:"status"`
	Config          json.RawMessage   `json:"config,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	LastDiscovered  time.Time         `json:"last_discovered,omitempty"`
	DiscoveryMethod DiscoveryMethod   `json:"discovery_method,omitempty"`
	NetworkInfo     *NetworkInfo      `json:"network_info,omitempty"`
}

// New creates a new Device with generated ID and timestamps
func New(tenantID, name string) *Device {
	now := time.Now().UTC()
	return &Device{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Name:      name,
		Status:    StatusUnknown,
		Tags:      make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Validate checks if the device data is valid
func (d *Device) Validate() error {
	if d.ID == "" {
		return fmt.Errorf("device id cannot be empty")
	}
	if d.TenantID == "" {
		return fmt.Errorf("tenant id cannot be empty")
	}
	if d.Name == "" {
		return fmt.Errorf("device name cannot be empty")
	}

	// Validate NetworkInfo if present
	if d.NetworkInfo != nil {
		if d.NetworkInfo.Port < 0 || d.NetworkInfo.Port > 65535 {
			return fmt.Errorf("invalid port number")
		}
		// Note: IP and MAC validation could be added here if needed
	}

	return nil
}

// SetStatus updates the device status and updated timestamp
func (d *Device) SetStatus(status Status) {
	d.Status = status
	d.UpdatedAt = time.Now().UTC()
}

// SetConfig updates the device configuration and updated timestamp
func (d *Device) SetConfig(config json.RawMessage) {
	d.Config = config
	d.UpdatedAt = time.Now().UTC()
}

// UpdateDiscoveryInfo updates the device's discovery-related information
func (d *Device) UpdateDiscoveryInfo(method DiscoveryMethod, networkInfo *NetworkInfo) {
	now := time.Now().UTC()
	d.LastDiscovered = now
	d.DiscoveryMethod = method
	d.NetworkInfo = networkInfo
	d.UpdatedAt = now
}

// AddTag adds or updates a tag value
func (d *Device) AddTag(key, value string) {
	if d.Tags == nil {
		d.Tags = make(map[string]string)
	}
	d.Tags[key] = value
	d.UpdatedAt = time.Now().UTC()
}

// RemoveTag removes a tag if it exists
func (d *Device) RemoveTag(key string) {
	if d.Tags != nil {
		delete(d.Tags, key)
		d.UpdatedAt = time.Now().UTC()
	}
}