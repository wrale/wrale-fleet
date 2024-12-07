package resolver

import (
	"fmt"
	"time"

	"github.com/wrale/wrale-fleet/fleet/sync/types"
)

// ConflictResolver implements state conflict resolution
type ConflictResolver struct {
	history    []*ResolutionRecord
	maxHistory int
}

// ResolutionRecord tracks conflict resolution history
type ResolutionRecord struct {
	States     []*types.VersionedState
	Changes    []*types.StateChange
	Result     *types.VersionedState
	ResolvedAt time.Time
}

// NewResolver creates a new conflict resolver
func NewResolver(maxHistory int) *ConflictResolver {
	return &ConflictResolver{
		history:    make([]*ResolutionRecord, 0),
		maxHistory: maxHistory,
	}
}

// DetectConflicts identifies conflicts between states
func (r *ConflictResolver) DetectConflicts(states []*types.VersionedState) ([]*types.StateChange, error) {
	if len(states) < 2 {
		return nil, nil
	}

	changes := make([]*types.StateChange, 0)

	// Compare each pair of states
	for i := 0; i < len(states)-1; i++ {
		for j := i + 1; j < len(states); j++ {
			s1, s2 := states[i], states[j]

			// Check for divergent updates
			if s1.UpdatedAt.Equal(s2.UpdatedAt) {
				// Simultaneous updates
				changes = append(changes, &types.StateChange{
					PrevVersion: s1.Version,
					NewVersion:  s2.Version,
					Changes:     detectStateChanges(s1, s2),
					Timestamp:   time.Now(),
					Source:      "conflict_detection",
				})
			} else if hasConflictingChanges(s1, s2) {
				// Changes that affect same properties
				changes = append(changes, &types.StateChange{
					PrevVersion: s1.Version,
					NewVersion:  s2.Version,
					Changes:     detectStateChanges(s1, s2),
					Timestamp:   time.Now(),
					Source:      "conflict_detection",
				})
			}
		}
	}

	return changes, nil
}

// ResolveConflicts resolves detected conflicts
func (r *ConflictResolver) ResolveConflicts(changes []*types.StateChange) (*types.VersionedState, error) {
	if len(changes) == 0 {
		return nil, fmt.Errorf("no changes to resolve")
	}

	// For v1.0, implement last-writer-wins with validation
	var latest *types.StateChange
	for _, change := range changes {
		if latest == nil || change.Timestamp.After(latest.Timestamp) {
			latest = change
		}
	}

	// Create new state version
	newState := &types.VersionedState{
		Version:   types.StateVersion(fmt.Sprintf("v-%d", time.Now().UnixNano())),
		UpdatedAt: time.Now(),
		UpdatedBy: "conflict_resolver",
	}

	// Apply changes from winning version
	if err := applyChanges(newState, latest.Changes); err != nil {
		return nil, fmt.Errorf("failed to apply changes: %w", err)
	}

	// Record resolution
	record := &ResolutionRecord{
		Changes:    changes,
		Result:     newState,
		ResolvedAt: time.Now(),
	}
	r.addRecord(record)

	return newState, nil
}

// ValidateResolution validates a resolved state
func (r *ConflictResolver) ValidateResolution(state *types.VersionedState) error {
	// For v1.0, implement basic validation:

	// Check required fields
	if state.Version == "" {
		return fmt.Errorf("missing state version")
	}
	if state.UpdatedAt.IsZero() {
		return fmt.Errorf("missing update timestamp")
	}
	if state.UpdatedBy == "" {
		return fmt.Errorf("missing update source")
	}

	// Validate device state fields
	if state.State.ID == "" {
		return fmt.Errorf("missing device ID")
	}
	if state.State.Status == "" {
		return fmt.Errorf("missing device status")
	}

	// Validate metrics
	metrics := state.State.Metrics
	if metrics.Temperature < 0 || metrics.Temperature > 100 {
		return fmt.Errorf("invalid temperature range")
	}
	if metrics.CPULoad < 0 || metrics.CPULoad > 100 {
		return fmt.Errorf("invalid CPU load range")
	}

	return nil
}

// GetResolutionHistory returns conflict resolution history
func (r *ConflictResolver) GetResolutionHistory() []*ResolutionRecord {
	return r.history
}

// addRecord adds a resolution record to history
func (r *ConflictResolver) addRecord(record *ResolutionRecord) {
	r.history = append(r.history, record)

	// Trim history if needed
	if len(r.history) > r.maxHistory {
		r.history = r.history[len(r.history)-r.maxHistory:]
	}
}

// hasConflictingChanges checks if states have conflicting changes
func hasConflictingChanges(s1, s2 *types.VersionedState) bool {
	// Check for changes to same properties
	changes := detectStateChanges(s1, s2)
	return len(changes) > 0
}

// detectStateChanges identifies changes between states
func detectStateChanges(s1, s2 *types.VersionedState) map[string]interface{} {
	changes := make(map[string]interface{})

	// Compare basic fields
	if s1.State.Status != s2.State.Status {
		changes["status"] = s2.State.Status
	}

	// Compare metrics
	if s1.State.Metrics.Temperature != s2.State.Metrics.Temperature {
		changes["temperature"] = s2.State.Metrics.Temperature
	}
	if s1.State.Metrics.PowerUsage != s2.State.Metrics.PowerUsage {
		changes["power_usage"] = s2.State.Metrics.PowerUsage
	}
	if s1.State.Metrics.CPULoad != s2.State.Metrics.CPULoad {
		changes["cpu_load"] = s2.State.Metrics.CPULoad
	}
	if s1.State.Metrics.MemoryUsage != s2.State.Metrics.MemoryUsage {
		changes["memory_usage"] = s2.State.Metrics.MemoryUsage
	}

	return changes
}

// applyChanges applies a set of changes to a state
func applyChanges(state *types.VersionedState, changes map[string]interface{}) error {
	for key, value := range changes {
		switch key {
		case "status":
			if status, ok := value.(string); ok {
				state.State.Status = status
			}
		case "temperature":
			if temp, ok := value.(float64); ok {
				state.State.Metrics.Temperature = temp
			}
		case "power_usage":
			if power, ok := value.(float64); ok {
				state.State.Metrics.PowerUsage = power
			}
		case "cpu_load":
			if cpu, ok := value.(float64); ok {
				state.State.Metrics.CPULoad = cpu
			}
		case "memory_usage":
			if mem, ok := value.(float64); ok {
				state.State.Metrics.MemoryUsage = mem
			}
		}
	}
	return nil
}
