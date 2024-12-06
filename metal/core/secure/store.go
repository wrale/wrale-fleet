package secure

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	hw "github.com/wrale/wrale-fleet/metal/hw/secure"
)

// FileStore implements StateStore using the filesystem
type FileStore struct {
	sync.RWMutex
	basePath string
}

// NewFileStore creates a new file-based state store
func NewFileStore(basePath string) (*FileStore, error) {
	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	return &FileStore{
		basePath: basePath,
	}, nil
}

// devicePath returns the path for a device's state file
func (s *FileStore) devicePath(deviceID string) string {
	return filepath.Join(s.basePath, fmt.Sprintf("device_%s.json", deviceID))
}

// eventPath returns the path for a device's event log
func (s *FileStore) eventPath(deviceID string) string {
	return filepath.Join(s.basePath, fmt.Sprintf("events_%s.jsonl", deviceID))
}

// SaveState persists the current security state
func (s *FileStore) SaveState(ctx context.Context, deviceID string, state hw.TamperState) error {
	s.Lock()
	defer s.Unlock()

	// Add timestamp if not set
	if state.LastCheck.IsZero() {
		state.LastCheck = time.Now()
	}

	// Marshal state to JSON
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to file
	if err := os.WriteFile(s.devicePath(deviceID), data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// LoadState retrieves the last known security state
func (s *FileStore) LoadState(ctx context.Context, deviceID string) (hw.TamperState, error) {
	s.RLock()
	defer s.RUnlock()

	// Read state file
	data, err := os.ReadFile(s.devicePath(deviceID))
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty state if file doesn't exist
			return hw.TamperState{LastCheck: time.Now()}, nil
		}
		return hw.TamperState{}, fmt.Errorf("failed to read state file: %w", err)
	}

	// Unmarshal state
	var state hw.TamperState
	if err := json.Unmarshal(data, &state); err != nil {
		return hw.TamperState{}, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return state, nil
}

// LogEvent records a security event
func (s *FileStore) LogEvent(ctx context.Context, deviceID string, eventType string, details interface{}) error {
	s.Lock()
	defer s.Unlock()

	// Create event record
	event := Event{
		DeviceID:  deviceID,
		Type:      eventType,
		Timestamp: time.Now(),
		Details:   details,
	}

	// Marshal event
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Append to event log file
	f, err := os.OpenFile(s.eventPath(deviceID), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open event log: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}

	return nil
}

// GetEvents retrieves recent security events
func (s *FileStore) GetEvents(ctx context.Context, deviceID string, since time.Time) ([]Event, error) {
	s.RLock()
	defer s.RUnlock()

	// Read event log file
	data, err := os.ReadFile(s.eventPath(deviceID))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read event log: %w", err)
	}

	// Parse events
	var events []Event
	for _, line := range splitLines(data) {
		if len(line) == 0 {
			continue
		}

		var event Event
		if err := json.Unmarshal(line, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal event: %w", err)
		}

		if event.Timestamp.After(since) {
			events = append(events, event)
		}
	}

	return events, nil
}

// Helper to split byte slice into lines
func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
