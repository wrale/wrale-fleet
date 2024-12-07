package store

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "sort"
    "sync"
    "time"

    synctypes "github.com/wrale/wrale-fleet/fleet/sync/types"
)

// FileStore implements state persistence using filesystem
type FileStore struct {
    baseDir    string
    mu         sync.RWMutex

    // Cache for performance
    stateCache map[synctypes.StateVersion]*synctypes.VersionedState
    changeLog  []*synctypes.StateChange
}

const (
    statesDir  = "states"
    changesDir = "changes"
)

// NewFileStore creates a new file-based store
func NewFileStore(baseDir string) (*FileStore, error) {
    // Create required directories
    for _, dir := range []string{statesDir, changesDir} {
        path := filepath.Join(baseDir, dir)
        if err := os.MkdirAll(path, 0755); err != nil {
            return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
        }
    }

    store := &FileStore{
        baseDir:    baseDir,
        stateCache: make(map[synctypes.StateVersion]*synctypes.VersionedState),
        changeLog:  make([]*synctypes.StateChange, 0),
    }

    // Load existing states into cache
    if err := store.loadStates(); err != nil {
        return nil, err
    }

    // Load change log
    if err := store.loadChanges(); err != nil {
        return nil, err
    }

    return store, nil
}

// GetState retrieves a state by version
func (s *FileStore) GetState(version synctypes.StateVersion) (*synctypes.VersionedState, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    // Check cache first
    if state, exists := s.stateCache[version]; exists {
        return state, nil
    }

    // Load from file
    path := filepath.Join(s.baseDir, statesDir, string(version)+".json")
    var state synctypes.VersionedState
    if err := readJSON(path, &state); err != nil {
        return nil, fmt.Errorf("failed to read state: %w", err)
    }

    // Update cache
    s.stateCache[version] = &state
    return &state, nil
}

// SaveState persists a state version
func (s *FileStore) SaveState(state *synctypes.VersionedState) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    // Save to file
    path := filepath.Join(s.baseDir, statesDir, string(state.Version)+".json")
    if err := writeJSON(path, state); err != nil {
        return fmt.Errorf("failed to write state: %w", err)
    }

    // Update cache
    s.stateCache[state.Version] = state
    return nil
}

// ListVersions returns all available state versions
func (s *FileStore) ListVersions() ([]synctypes.StateVersion, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    // List state files
    pattern := filepath.Join(s.baseDir, statesDir, "*.json")
    files, err := filepath.Glob(pattern)
    if err != nil {
        return nil, fmt.Errorf("failed to list state files: %w", err)
    }

    // Extract versions from filenames
    versions := make([]synctypes.StateVersion, 0, len(files))
    for _, file := range files {
        base := filepath.Base(file)
        version := synctypes.StateVersion(base[:len(base)-5]) // Remove .json
        versions = append(versions, version)
    }

    // Sort by timestamp (versions should include timestamp)
    sort.Slice(versions, func(i, j int) bool {
        return string(versions[i]) < string(versions[j])
    })

    return versions, nil
}

// TrackChange records a state change
func (s *FileStore) TrackChange(change *synctypes.StateChange) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    // Save change to file
    filename := fmt.Sprintf("%d-%s.json", change.Timestamp.UnixNano(), change.NewVersion)
    path := filepath.Join(s.baseDir, changesDir, filename)
    if err := writeJSON(path, change); err != nil {
        return fmt.Errorf("failed to write change: %w", err)
    }

    // Update log
    s.changeLog = append(s.changeLog, change)

    // Sort by timestamp
    sort.Slice(s.changeLog, func(i, j int) bool {
        return s.changeLog[i].Timestamp.Before(s.changeLog[j].Timestamp)
    })

    return nil
}

// GetChanges retrieves changes since a given time
func (s *FileStore) GetChanges(since time.Time) ([]*synctypes.StateChange, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    changes := make([]*synctypes.StateChange, 0)
    for _, change := range s.changeLog {
        if change.Timestamp.After(since) {
            changes = append(changes, change)
        }
    }

    return changes, nil
}

// loadStates loads existing states into cache
func (s *FileStore) loadStates() error {
    pattern := filepath.Join(s.baseDir, statesDir, "*.json")
    files, err := filepath.Glob(pattern)
    if err != nil {
        return fmt.Errorf("failed to list state files: %w", err)
    }

    for _, file := range files {
        var state synctypes.VersionedState
        if err := readJSON(file, &state); err != nil {
            return fmt.Errorf("failed to read state %s: %w", file, err)
        }
        s.stateCache[state.Version] = &state
    }

    return nil
}

// loadChanges loads the change log
func (s *FileStore) loadChanges() error {
    pattern := filepath.Join(s.baseDir, changesDir, "*.json")
    files, err := filepath.Glob(pattern)
    if err != nil {
        return fmt.Errorf("failed to list change files: %w", err)
    }

    for _, file := range files {
        var change synctypes.StateChange
        if err := readJSON(file, &change); err != nil {
            return fmt.Errorf("failed to read change %s: %w", file, err)
        }
        s.changeLog = append(s.changeLog, &change)
    }

    // Sort by timestamp
    sort.Slice(s.changeLog, func(i, j int) bool {
        return s.changeLog[i].Timestamp.Before(s.changeLog[j].Timestamp)
    })

    return nil
}

// readJSON reads and unmarshals JSON from a file
func readJSON(path string, v interface{}) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return err
    }
    return json.Unmarshal(data, v)
}

// writeJSON marshals and writes JSON to a file
func writeJSON(path string, v interface{}) error {
    data, err := json.Marshal(v)
    if err != nil {
        return err
    }

    // Write to temp file first
    tempPath := path + ".tmp"
    if err := os.WriteFile(tempPath, data, 0644); err != nil {
        return err
    }

    // Atomically replace the old file
    return os.Rename(tempPath, path)
}
