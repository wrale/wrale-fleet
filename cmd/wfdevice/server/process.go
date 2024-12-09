package server

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"go.uber.org/zap"
)

const (
	// serverPIDFile is the name of the file storing the running server's PID
	serverPIDFile = "wfdevice.pid"
)

// GetRunningPID returns the PID of the running server, if any.
// Returns 0 if no server is running or the PID file doesn't exist.
func GetRunningPID(dataDir string) (int, error) {
	pidFile := filepath.Join(dataDir, serverPIDFile)
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
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

// writePIDFile writes the current process ID to the PID file.
// Creates the data directory if it doesn't exist.
func (s *Server) writePIDFile() error {
	// Ensure data directory exists
	if err := os.MkdirAll(s.cfg.DataDir, 0755); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}

	pid := os.Getpid()
	if err := os.WriteFile(s.pidFilePath(), []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		return fmt.Errorf("writing pid file: %w", err)
	}

	return nil
}

// removePIDFile removes the PID file.
// Logs a warning but doesn't return an error if removal fails.
func (s *Server) removePIDFile() {
	if err := os.Remove(s.pidFilePath()); err != nil && !os.IsNotExist(err) {
		s.logger.Warn("failed to remove pid file", zap.Error(err))
	}
}
