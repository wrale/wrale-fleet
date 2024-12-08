package group

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/wrale/fleet/internal/fleet/device"
)

// Type represents the type of device group
type Type string

const (
	// TypeStatic represents a group with explicitly assigned devices
	TypeStatic Type = "static"
	// TypeDynamic represents a group with devices matched by criteria
	TypeDynamic Type = "dynamic"
)

// MembershipQuery defines criteria for dynamic group membership
type MembershipQuery struct {
	Tags    map[string]string `json:"tags,omitempty"`              // Tag-based matching
	Status  device.Status     `json:"status,omitempty"`            // Status-based matching
	Regions []string          `json:"regions,omitempty"`           // Region-based matching
	Custom  json.RawMessage   `json:"custom_criteria,omitempty"`   // Custom query criteria
}

// Properties represents group configuration properties
type Properties struct {
	ConfigTemplate  json.RawMessage            `json:"config_template,omitempty"`  // Base configuration for group devices
	PolicyOverrides map[string]json.RawMessage `json:"policy_overrides,omitempty"` // Policy overrides for the group
	Metadata        map[string]string          `json:"metadata,omitempty"`         // Additional group metadata
}

// Group represents a collection of devices with shared management properties
type Group struct {
	ID           string          `json:"id"`
	TenantID     string          `json:"tenant_id"`
	Name         string          `json:"name"`
	Description  string          `json:"description,omitempty"`
	Type         Type            `json:"type"`
	ParentID     string          `json:"parent_id,omitempty"`     // ID of parent group for inheritance
	Path         string          `json:"path"`                     // Full path in group hierarchy
	Query        *MembershipQuery `json:"query,omitempty"`         // Criteria for dynamic membership
	Properties   Properties      `json:"properties"`              // Group configuration and policies
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeviceCount  int             `json:"device_count"`           // Count of member devices
}

// New creates a new Group with generated ID and timestamps
func New(tenantID, name string, groupType Type) *Group {
	now := time.Now().UTC()
	return &Group{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		Name:       name,
		Type:       groupType,
		Properties: Properties{
			Metadata: make(map[string]string),
		},
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// Validate checks if the group data is valid
func (g *Group) Validate() error {
	const op = "Group.Validate"

	if g.ID == "" {
		return E(op, ErrCodeInvalidGroup, "group id cannot be empty", nil)
	}
	if g.TenantID == "" {
		return E(op, ErrCodeInvalidGroup, "tenant id cannot be empty", nil)
	}
	if g.Name == "" {
		return E(op, ErrCodeInvalidGroup, "group name cannot be empty", nil)
	}
	if g.Type == "" {
		return E(op, ErrCodeInvalidGroup, "group type cannot be empty", nil)
	}

	// Validate dynamic group query if present
	if g.Type == TypeDynamic && g.Query == nil {
		return E(op, ErrCodeInvalidGroup, "dynamic group must have query criteria", nil)
	}

	return nil
}

// SetQuery updates the group's membership query criteria
func (g *Group) SetQuery(query *MembershipQuery) error {
	const op = "Group.SetQuery"

	if g.Type != TypeDynamic {
		return E(op, ErrCodeInvalidOperation, "cannot set query on non-dynamic group", nil)
	}
	if query == nil {
		return E(op, ErrCodeInvalidOperation, "query cannot be nil", nil)
	}

	g.Query = query
	g.UpdatedAt = time.Now().UTC()
	return nil
}

// SetParent updates the group's parent ID and path
func (g *Group) SetParent(parentID, parentPath string) error {
	const op = "Group.SetParent"

	if parentID != "" {
		if parentPath == "" {
			return E(op, ErrCodeInvalidOperation, "parent path cannot be empty when parent ID is set", nil)
		}
		g.Path = parentPath + "/" + g.ID
	} else {
		g.Path = g.ID
	}

	g.ParentID = parentID
	g.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateProperties updates the group's configuration properties
func (g *Group) UpdateProperties(properties Properties) error {
	const op = "Group.UpdateProperties"

	g.Properties = properties
	g.UpdatedAt = time.Now().UTC()
	return nil
}

// IsAncestor checks if the given group ID is an ancestor of this group
func (g *Group) IsAncestor(groupID string) bool {
	// Check each component of the path
	current := g.ParentID
	for current != "" {
		if current == groupID {
			return true
		}
		// In practice, we'd load the parent group to check its parent
		// This is just a placeholder implementation
		break
	}
	return false
}