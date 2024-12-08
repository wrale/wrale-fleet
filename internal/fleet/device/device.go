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

// ComplianceStatus represents regulatory compliance state
type ComplianceStatus struct {
	IsCompliant    bool                 `json:"is_compliant"`
	LastCheck      time.Time            `json:"last_check"`
	Requirements   []string             `json:"requirements"`
	Violations     []string             `json:"violations,omitempty"`
	Certifications map[string]time.Time `json:"certifications,omitempty"`
}

// NetworkInfo contains device network-related information
type NetworkInfo struct {
	IPAddress  string            `json:"ip_address,omitempty"`
	MACAddress string            `json:"mac_address,omitempty"`
	Hostname   string            `json:"hostname,omitempty"`
	Port       int               `json:"port,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// ConfigVersion represents a versioned configuration
type ConfigVersion struct {
	Version     int             `json:"version"`
	Config      json.RawMessage `json:"config"`
	Hash        string          `json:"hash"`
	AppliedAt   time.Time       `json:"applied_at"`
	AppliedBy   string          `json:"applied_by"`
	ValidatedAt *time.Time      `json:"validated_at,omitempty"`
}

// OfflineCapabilities represents device airgap features
type OfflineCapabilities struct {
	SupportsAirgap    bool          `json:"supports_airgap"`
	LastSyncTime      time.Time     `json:"last_sync_time,omitempty"`
	OfflineOperations []string      `json:"offline_operations,omitempty"`
	SyncInterval      time.Duration `json:"sync_interval,omitempty"`
	LocalBufferSize   int64         `json:"local_buffer_size,omitempty"`
}

// Device represents a managed Raspberry Pi device in the fleet
type Device struct {
	ID              string            `json:"id"`
	TenantID        string            `json:"tenant_id"`
	Name            string            `json:"name"`
	Status          Status            `json:"status"`
	Config          json.RawMessage   `json:"config,omitempty"`
	ConfigHistory   []ConfigVersion   `json:"config_history,omitempty"`
	LastConfigHash  string            `json:"last_config_hash,omitempty"`
	Tags            map[string]string `json:"tags,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	LastDiscovered  time.Time         `json:"last_discovered,omitempty"`
	DiscoveryMethod DiscoveryMethod   `json:"discovery_method,omitempty"`
	NetworkInfo     *NetworkInfo      `json:"network_info,omitempty"`

	// Security and compliance fields
	SecureBootEnabled bool              `json:"secure_boot_enabled"`
	SecurityVersion   string            `json:"security_version,omitempty"`
	ComplianceStatus  *ComplianceStatus `json:"compliance_status,omitempty"`

	// Airgapped operation support
	OfflineCapabilities *OfflineCapabilities `json:"offline_capabilities,omitempty"`
}

// New creates a new Device with generated ID and timestamps
func New(tenantID, name string) *Device {
	now := time.Now().UTC()
	return &Device{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		Name:          name,
		Status:        StatusUnknown,
		Tags:          make(map[string]string),
		CreatedAt:     now,
		UpdatedAt:     now,
		ConfigHistory: make([]ConfigVersion, 0),
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
	}

	// Validate OfflineCapabilities if present
	if d.OfflineCapabilities != nil {
		if d.OfflineCapabilities.SyncInterval < 0 {
			return fmt.Errorf("sync interval cannot be negative")
		}
		if d.OfflineCapabilities.LocalBufferSize < 0 {
			return fmt.Errorf("local buffer size cannot be negative")
		}
	}

	return nil
}

// SetStatus updates the device status and updated timestamp
func (d *Device) SetStatus(status Status) {
	d.Status = status
	d.UpdatedAt = time.Now().UTC()
}

// SetConfig updates the device configuration with versioning
func (d *Device) SetConfig(config json.RawMessage, appliedBy string) {
	now := time.Now().UTC()

	// Create new config version
	version := len(d.ConfigHistory) + 1
	hash := calculateConfigHash(config) // Implementation needed

	configVersion := ConfigVersion{
		Version:   version,
		Config:    config,
		Hash:      hash,
		AppliedAt: now,
		AppliedBy: appliedBy,
	}

	// Add to history and update current config
	d.ConfigHistory = append(d.ConfigHistory, configVersion)
	d.Config = config
	d.LastConfigHash = hash
	d.UpdatedAt = now
}

// ValidateConfig marks the current config as validated
func (d *Device) ValidateConfig() {
	if len(d.ConfigHistory) > 0 {
		now := time.Now().UTC()
		latest := &d.ConfigHistory[len(d.ConfigHistory)-1]
		latest.ValidatedAt = &now
	}
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

// UpdateComplianceStatus updates the device's compliance information
func (d *Device) UpdateComplianceStatus(status *ComplianceStatus) {
	d.ComplianceStatus = status
	d.UpdatedAt = time.Now().UTC()
}

// UpdateOfflineCapabilities updates the device's airgap support information
func (d *Device) UpdateOfflineCapabilities(capabilities *OfflineCapabilities) {
	d.OfflineCapabilities = capabilities
	d.UpdatedAt = time.Now().UTC()
}

// calculateConfigHash generates a hash of the configuration
// Implementation needed - placeholder for now
func calculateConfigHash(config json.RawMessage) string {
	return "hash-placeholder" // TODO: Implement proper hashing
}
