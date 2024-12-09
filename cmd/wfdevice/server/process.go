package server

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"go.uber.org/zap"
)

const (
	// serverPIDFile is the name of the file storing the running server's PID
	serverPIDFile = "wfdevice.pid"

	// dirPermissions defines the permissions for the data directory
	dirPermissions = 0750

	// filePermissions defines the permissions for the PID file
	filePermissions = 0600
)

var (
	// ErrInvalidPath indicates an invalid or unsafe path was provided
	ErrInvalidPath = errors.New("invalid path")
)

// GetRunningPID returns the PID of the running server, if any.
// Returns 0 if no server is running or the PID file doesn't exist.
func GetRunningPID(dataDir string) (int, error) {
	if err := validatePath(dataDir); err != nil {
		return 0, fmt.Errorf("invalid data directory: %w", err)
	}

	pidFile := filepath.Join(dataDir, serverPIDFile)
	if err := validatePath(pidFile); err != nil {
		return 0, fmt.Errorf("invalid pid file path: %w", err)
	}

	// #nosec G304 -- path has been validated and normalized
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("reading pid file: %w", err)
	}

	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return 0, fmt.Errorf("invalid pid file content: %w", err)
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return 0, nil
	}

	// Check if the process is actually running
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return 0, nil
	}

	return pid, nil
}

// pidFilePath returns the path to the PID file
func (s *Server) pidFilePath() string {
	return filepath.Join(s.cfg.DataDir, serverPIDFile)
}

// validatePath checks if a path is safe to use
func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("%w: empty path", ErrInvalidPath)
	}

	// Clean the path and ensure it doesn't contain suspicious components
	cleaned := filepath.Clean(path)
	if strings.Contains(cleaned, "..") {
		return fmt.Errorf("%w: path contains parent directory references", ErrInvalidPath)
	}

	// Additional path validation can be added here as needed
	return nil
}

// ensureDataDir ensures the data directory exists with proper permissions
func (s *Server) ensureDataDir() error {
	if err := validatePath(s.cfg.DataDir); err != nil {
		return fmt.Errorf("invalid data directory: %w", err)
	}

	// Create directory with restrictive permissions
	if err := os.MkdirAll(s.cfg.DataDir, dirPermissions); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}

	return nil
}

// writePIDFile writes the current process ID to the PID file.
// Creates the data directory if it doesn't exist.
func (s *Server) writePIDFile() error {
	// Ensure data directory exists with proper permissions
	if err := s.ensureDataDir(); err != nil {
		return err
	}

	// Validate PID file path
	pidFile := s.pidFilePath()
	if err := validatePath(pidFile); err != nil {
		return fmt.Errorf("invalid pid file path: %w", err)
	}

	// Write PID file with restrictive permissions
	pid := os.Getpid()
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), filePermissions); err != nil {
		return fmt.Errorf("writing pid file: %w", err)
	}

	return nil
}

// removePIDFile removes the PID file.
// Logs a warning but doesn't return an error if removal fails.
func (s *Server) removePIDFile() {
	pidFile := s.pidFilePath()
	if err := validatePath(pidFile); err != nil {
		s.logger.Warn("invalid pid file path during removal", zap.Error(err))
		return
	}

	if err := os.Remove(pidFile); err != nil && !os.IsNotExist(err) {
		s.logger.Warn("failed to remove pid file", zap.Error(err))
	}
}
