package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wrale/wrale-fleet/fleet/edge/agent"
)

// FileStore implements persistent state storage using the filesystem
type FileStore struct {
	baseDir string
	mu      sync.RWMutex
}

const (
	stateFile    = "state.json"
	commandsFile = "commands.json"
	configFile   = "config.json"
)

// NewFileStore creates a new file-based store
func NewFileStore(baseDir string) (*FileStore, error) {
	// Ensure directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create store directory: %w", err)
	}

	return &FileStore{
		baseDir: baseDir,
	}, nil
}

// GetState retrieves the current agent state
func (s *FileStore) GetState() (agent.AgentState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var state agent.AgentState
	if err := s.readJSON(stateFile, &state); err != nil {
		if os.IsNotExist(err) {
			// Return empty state if file doesn't exist
			return state, nil
		}
		return state, fmt.Errorf("failed to read state: %w", err)
	}

	return state, nil
}

// UpdateState stores the current agent state
func (s *FileStore) UpdateState(state agent.AgentState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.writeJSON(stateFile, state); err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	return nil
}

// GetCommandHistory retrieves the command execution history
func (s *FileStore) GetCommandHistory() ([]agent.CommandResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []agent.CommandResult
	if err := s.readJSON(commandsFile, &results); err != nil {
		if os.IsNotExist(err) {
			// Return empty slice if file doesn't exist
			return []agent.CommandResult{}, nil
		}
		return nil, fmt.Errorf("failed to read command history: %w", err)
	}

	return results, nil
}

// AddCommandResult adds a command result to the history
func (s *FileStore) AddCommandResult(result agent.CommandResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Read existing history
	var results []agent.CommandResult
	if err := s.readJSON(commandsFile, &results); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read command history: %w", err)
	}

	// Add new result
	results = append(results, result)

	// Limit history size (keep last 1000 commands)
	if len(results) > 1000 {
		results = results[len(results)-1000:]
	}

	// Write updated history
	if err := s.writeJSON(commandsFile, results); err != nil {
		return fmt.Errorf("failed to write command history: %w", err)
	}

	return nil
}

// GetConfig retrieves the current configuration
func (s *FileStore) GetConfig() (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var config map[string]interface{}
	if err := s.readJSON(configFile, &config); err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	return config, nil
}

// UpdateConfig stores the current configuration
func (s *FileStore) UpdateConfig(config map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.writeJSON(configFile, config); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// readJSON reads and unmarshals JSON from a file
func (s *FileStore) readJSON(filename string, v interface{}) error {
	path := filepath.Join(s.baseDir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// writeJSON marshals and writes JSON to a file
func (s *FileStore) writeJSON(filename string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	path := filepath.Join(s.baseDir, filename)
	tempPath := path + ".tmp"

	// Write to temp file first
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}

	// Atomically replace the old file
	return os.Rename(tempPath, path)
}

// Cleanup removes old command history entries
func (s *FileStore) Cleanup(maxAge time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var results []agent.CommandResult
	if err := s.readJSON(commandsFile, &results); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read command history: %w", err)
	}

	cutoff := time.Now().Add(-maxAge)
	filtered := make([]agent.CommandResult, 0, len(results))

	for _, result := range results {
		if result.CompletedAt.After(cutoff) {
			filtered = append(filtered, result)
		}
	}

	if err := s.writeJSON(commandsFile, filtered); err != nil {
		return fmt.Errorf("failed to write filtered command history: %w", err)
	}

	return nil
}
