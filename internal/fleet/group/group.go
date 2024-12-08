package group

import (
	"encoding/json"
	"strings"
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
	Tags    map[string]string `json:"tags,omitempty"`            // Tag-based matching
	Status  device.Status     `json:"status,omitempty"`          // Status-based matching
	Regions []string          `json:"regions,omitempty"`         // Region-based matching
	Custom  json.RawMessage   `json:"custom_criteria,omitempty"` // Custom query criteria
}

// Properties represents group configuration properties
type Properties struct {
	ConfigTemplate  json.RawMessage            `json:"config_template,omitempty"`  // Base configuration for group devices
	PolicyOverrides map[string]json.RawMessage `json:"policy_overrides,omitempty"` // Policy overrides for the group
	Metadata        map[string]string          `json:"metadata,omitempty"`         // Additional group metadata
}

// AncestryInfo contains information about a group's position in the hierarchy
type AncestryInfo struct {
	Path      string   `json:"path"`       // Full path in group hierarchy (e.g., "/root/parent/group")
	PathParts []string `json:"path_parts"` // Path components for efficient traversal
	Depth     int      `json:"depth"`      // Depth in the hierarchy (0 for root groups)
	Children  []string `json:"children"`   // Direct child group IDs
}

// Group represents a collection of devices with shared management properties
type Group struct {
	ID          string           `json:"id"`
	TenantID    string           `json:"tenant_id"`
	Name        string           `json:"name"`
	Description string           `json:"description,omitempty"`
	Type        Type             `json:"type"`
	ParentID    string           `json:"parent_id,omitempty"` // ID of parent group for inheritance
	Ancestry    AncestryInfo     `json:"ancestry"`            // Hierarchical relationship information
	Query       *MembershipQuery `json:"query,omitempty"`     // Criteria for dynamic membership
	Properties  Properties       `json:"properties"`          // Group configuration and policies
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	DeviceCount int              `json:"device_count"` // Count of member devices
}

// DeepCopy creates a deep copy of a Group and all its nested structures
func (g *Group) DeepCopy() *Group {
	if g == nil {
		return nil
	}

	result := &Group{
		ID:          g.ID,
		TenantID:    g.TenantID,
		Name:        g.Name,
		Description: g.Description,
		Type:        g.Type,
		ParentID:    g.ParentID,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
		DeviceCount: g.DeviceCount,
	}

	// Deep copy Ancestry with proper slice initialization
	result.Ancestry = AncestryInfo{
		Path:      g.Ancestry.Path,
		Depth:     g.Ancestry.Depth,
		PathParts: make([]string, len(g.Ancestry.PathParts)),
		Children:  make([]string, len(g.Ancestry.Children)),
	}

	copy(result.Ancestry.PathParts, g.Ancestry.PathParts)
	copy(result.Ancestry.Children, g.Ancestry.Children)

	// Deep copy Query if present
	if g.Query != nil {
		result.Query = &MembershipQuery{
			Status: g.Query.Status,
		}

		if g.Query.Tags != nil {
			result.Query.Tags = make(map[string]string, len(g.Query.Tags))
			for k, v := range g.Query.Tags {
				result.Query.Tags[k] = v
			}
		}

		if g.Query.Regions != nil {
			result.Query.Regions = make([]string, len(g.Query.Regions))
			copy(result.Query.Regions, g.Query.Regions)
		}

		if g.Query.Custom != nil {
			result.Query.Custom = make(json.RawMessage, len(g.Query.Custom))
			copy(result.Query.Custom, g.Query.Custom)
		}
	}

	// Deep copy Properties
	result.Properties = Properties{
		Metadata: make(map[string]string, len(g.Properties.Metadata)),
	}

	if g.Properties.ConfigTemplate != nil {
		result.Properties.ConfigTemplate = make(json.RawMessage, len(g.Properties.ConfigTemplate))
		copy(result.Properties.ConfigTemplate, g.Properties.ConfigTemplate)
	}

	if g.Properties.PolicyOverrides != nil {
		result.Properties.PolicyOverrides = make(map[string]json.RawMessage, len(g.Properties.PolicyOverrides))
		for k, v := range g.Properties.PolicyOverrides {
			newValue := make(json.RawMessage, len(v))
			copy(newValue, v)
			result.Properties.PolicyOverrides[k] = newValue
		}
	}

	for k, v := range g.Properties.Metadata {
		result.Properties.Metadata[k] = v
	}

	return result
}

// New creates a new Group with generated ID and timestamps
func New(tenantID, name string, groupType Type) *Group {
	now := time.Now().UTC()
	id := uuid.New().String()

	// Create ancestry with proper path structure
	pathStr := "/" + id
	pathParts := []string{id}

	return &Group{
		ID:       id,
		TenantID: tenantID,
		Name:     name,
		Type:     groupType,
		Ancestry: AncestryInfo{
			Path:      pathStr,
			PathParts: pathParts,
			Depth:     0,
			Children:  make([]string, 0),
		},
		Properties: Properties{
			Metadata:        make(map[string]string),
			PolicyOverrides: make(map[string]json.RawMessage),
		},
		CreatedAt: now,
		UpdatedAt: now,
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

	// Validate ancestry information
	if g.Ancestry.Path == "" {
		return E(op, ErrCodeInvalidGroup, "group ancestry path cannot be empty", nil)
	}
	if len(g.Ancestry.PathParts) == 0 {
		return E(op, ErrCodeInvalidGroup, "group ancestry path parts cannot be empty", nil)
	}
	if g.Ancestry.PathParts[len(g.Ancestry.PathParts)-1] != g.ID {
		return E(op, ErrCodeInvalidGroup, "group ancestry path must end with group ID", nil)
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

// SetParent updates the group's parent ID and ancestry information
func (g *Group) SetParent(parentID string, parentInfo *AncestryInfo) error {
	const op = "Group.SetParent"

	if parentID != "" {
		if parentInfo == nil {
			return E(op, ErrCodeInvalidOperation, "parent ancestry info cannot be nil when parent ID is set", nil)
		}

		// Update ancestry information
		g.ParentID = parentID
		g.Ancestry.Path = parentInfo.Path + "/" + g.ID
		g.Ancestry.PathParts = make([]string, len(parentInfo.PathParts)+1)
		copy(g.Ancestry.PathParts, parentInfo.PathParts)
		g.Ancestry.PathParts[len(parentInfo.PathParts)] = g.ID
		g.Ancestry.Depth = parentInfo.Depth + 1
	} else {
		// Reset to root group
		g.ParentID = ""
		g.Ancestry.Path = "/" + g.ID
		g.Ancestry.PathParts = []string{g.ID}
		g.Ancestry.Depth = 0
	}

	g.UpdatedAt = time.Now().UTC()
	return nil
}

// UpdateProperties updates the group's configuration properties
func (g *Group) UpdateProperties(properties Properties) error {
	g.Properties = properties
	g.UpdatedAt = time.Now().UTC()
	return nil
}

// AddChild adds a child group ID to this group's children list
func (g *Group) AddChild(childID string) {
	// Check if child already exists to prevent duplicates
	for _, id := range g.Ancestry.Children {
		if id == childID {
			return
		}
	}
	g.Ancestry.Children = append(g.Ancestry.Children, childID)
	g.UpdatedAt = time.Now().UTC()
}

// RemoveChild removes a child group ID from this group's children list
func (g *Group) RemoveChild(childID string) {
	children := make([]string, 0, len(g.Ancestry.Children))
	for _, id := range g.Ancestry.Children {
		if id != childID {
			children = append(children, id)
		}
	}
	g.Ancestry.Children = children
	g.UpdatedAt = time.Now().UTC()
}

// IsAncestor checks if the given group ID is an ancestor of this group
func (g *Group) IsAncestor(groupID string) bool {
	// Only check actual ancestors (exclude self from check)
	for _, id := range g.Ancestry.PathParts[:len(g.Ancestry.PathParts)-1] {
		if id == groupID {
			return true
		}
	}
	return false
}

// SharesAncestor checks if this group shares a common ancestor with another group
func (g *Group) SharesAncestor(other *Group) (string, bool) {
	for i := 0; i < len(g.Ancestry.PathParts) && i < len(other.Ancestry.PathParts); i++ {
		if g.Ancestry.PathParts[i] != other.Ancestry.PathParts[i] {
			if i == 0 {
				return "", false
			}
			return g.Ancestry.PathParts[i-1], true
		}
	}
	return "", false
}

// IsDescendant checks if the given group ID is a descendant of this group
func (g *Group) IsDescendant(groupID string) bool {
	for _, child := range g.Ancestry.Children {
		if child == groupID {
			return true
		}
	}
	return false
}

// GetAncestryPath returns the full ancestry path as a string slice
func (g *Group) GetAncestryPath() []string {
	result := make([]string, len(g.Ancestry.PathParts))
	copy(result, g.Ancestry.PathParts)
	return result
}

// GetEffectivePath returns the human-readable path using group names
func (g *Group) GetEffectivePath() string {
	return strings.TrimPrefix(g.Ancestry.Path, "/")
}