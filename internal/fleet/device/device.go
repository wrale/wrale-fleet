package device

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	const op = "Device.Validate"

	if d.ID == "" {
		return E(op, ErrCodeInvalidDevice, "device id cannot be empty", nil)
	}
	if d.TenantID == "" {
		return E(op, ErrCodeInvalidDevice, "tenant id cannot be empty", nil)
	}
	if d.Name == "" {
		return E(op, ErrCodeInvalidDevice, "device name cannot be empty", nil)
	}

	// Validate NetworkInfo if present
	if d.NetworkInfo != nil {
		if d.NetworkInfo.Port < 0 || d.NetworkInfo.Port > 65535 {
			return E(op, ErrCodeInvalidDevice, "invalid port number", nil)
		}
	}

	// Validate OfflineCapabilities if present
	if d.OfflineCapabilities != nil {
		if d.OfflineCapabilities.SyncInterval < 0 {
			return E(op, ErrCodeInvalidDevice, "sync interval cannot be negative", nil)
		}
		if d.OfflineCapabilities.LocalBufferSize < 0 {
			return E(op, ErrCodeInvalidDevice, "local buffer size cannot be negative", nil)
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
func (d *Device) SetConfig(config json.RawMessage, appliedBy string) error {
	const op = "Device.SetConfig"

	if len(config) == 0 {
		return E(op, ErrCodeInvalidOperation, "config cannot be empty", nil)
	}

	now := time.Now().UTC()
	hash := calculateConfigHash(config)

	// Create new config version
	version := len(d.ConfigHistory) + 1
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

	return nil
}

// ValidateConfig marks the current config as validated
func (d *Device) ValidateConfig() error {
	const op = "Device.ValidateConfig"

	if len(d.ConfigHistory) == 0 {
		return E(op, ErrCodeInvalidOperation, "no configuration history exists", nil)
	}

	now := time.Now().UTC()
	latest := &d.ConfigHistory[len(d.ConfigHistory)-1]
	latest.ValidatedAt = &now
	return nil
}

// UpdateDiscoveryInfo updates the device's discovery-related information
func (d *Device) UpdateDiscoveryInfo(method DiscoveryMethod, networkInfo *NetworkInfo) error {
	const op = "Device.UpdateDiscoveryInfo"

	if networkInfo != nil {
		if networkInfo.Port < 0 || networkInfo.Port > 65535 {
			return E(op, ErrCodeInvalidDevice, "invalid port number", nil)
		}
	}

	now := time.Now().UTC()
	d.LastDiscovered = now
	d.DiscoveryMethod = method
	d.NetworkInfo = networkInfo
	d.UpdatedAt = now

	return nil
}

// AddTag adds or updates a tag value
func (d *Device) AddTag(key, value string) error {
	const op = "Device.AddTag"

	if key == "" {
		return E(op, ErrCodeInvalidOperation, "tag key cannot be empty", nil)
	}

	if d.Tags == nil {
		d.Tags = make(map[string]string)
	}
	d.Tags[key] = value
	d.UpdatedAt = time.Now().UTC()

	return nil
}

// RemoveTag removes a tag if it exists
func (d *Device) RemoveTag(key string) error {
	const op = "Device.RemoveTag"

	if key == "" {
		return E(op, ErrCodeInvalidOperation, "tag key cannot be empty", nil)
	}

	if d.Tags != nil {
		delete(d.Tags, key)
		d.UpdatedAt = time.Now().UTC()
	}
	return nil
}

// UpdateComplianceStatus updates the device's compliance information
func (d *Device) UpdateComplianceStatus(status *ComplianceStatus) error {
	const op = "Device.UpdateComplianceStatus"

	if status == nil {
		return E(op, ErrCodeInvalidOperation, "compliance status cannot be nil", nil)
	}

	d.ComplianceStatus = status
	d.UpdatedAt = time.Now().UTC()
	return nil
}

// GetOfflineCapabilities returns the device's airgap support configuration
func (d *Device) GetOfflineCapabilities() *OfflineCapabilities {
	return d.OfflineCapabilities
}

// UpdateOfflineCapabilities updates the device's airgap support information
func (d *Device) UpdateOfflineCapabilities(capabilities *OfflineCapabilities) error {
	const op = "Device.UpdateOfflineCapabilities"

	if capabilities == nil {
		return E(op, ErrCodeInvalidOperation, "offline capabilities cannot be nil", nil)
	}

	if capabilities.SyncInterval < 0 {
		return E(op, ErrCodeInvalidDevice, "sync interval cannot be negative", nil)
	}
	if capabilities.LocalBufferSize < 0 {
		return E(op, ErrCodeInvalidDevice, "local buffer size cannot be negative", nil)
	}

	d.OfflineCapabilities = capabilities
	d.UpdatedAt = time.Now().UTC()
	return nil
}

// calculateConfigHash generates a hash of the configuration
func calculateConfigHash(config json.RawMessage) string {
	hash := sha256.Sum256(config)
	return hex.EncodeToString(hash[:])
}
