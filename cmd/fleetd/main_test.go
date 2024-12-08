package main

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/fleet/device/store/memory"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
)

// Test constants define reasonable timeouts for different operations.
// These values are long enough to allow for normal operation but short
// enough to catch deadlocks or hanging operations quickly.
const (
	testInitTimeout     = 2 * time.Second // Time allowed for initialization
	testRunTime         = 3 * time.Second // Time to let the demo run
	testShutdownTimeout = 2 * time.Second // Time allowed for clean shutdown
)

// TestDemoManager verifies the continuous demo operation capabilities,
// including startup, running state, and graceful shutdown.
func TestDemoManager(t *testing.T) {
	// Create a test logger that integrates with the testing framework
	logger := zaptest.NewLogger(t)
	defer func() {
		_ = safeSync(logger)
	}()

	// Create the basic infrastructure needed for the demo
	store := memory.New()
	service := device.NewService(store, logger)

	// Initialize the demo manager
	dm := NewDemoManager(service, logger)
	require.NotNil(t, dm, "Demo manager should be created successfully")

	// Start the demo and verify initialization
	err := dm.Start()
	require.NoError(t, err, "Demo should start without errors")

	// Let the demo run for a while to verify continuous operation
	time.Sleep(testRunTime)

	// Verify that the demo device exists and is being maintained
	devices, err := service.List(context.Background(), device.ListOptions{
		TenantID: "demo-tenant",
	})
	require.NoError(t, err, "Should be able to list devices")
	require.Len(t, devices, 1, "Should have exactly one demo device")
	assert.Equal(t, "Demo Raspberry Pi", devices[0].Name)
	assert.Equal(t, device.StatusOnline, devices[0].Status)

	// Verify graceful shutdown
	err = dm.Stop()
	assert.NoError(t, err, "Demo should stop gracefully")
}

// TestMainSignalHandling verifies that the main program handles
// system signals appropriately, including proper initialization
// and graceful shutdown.
func TestMainSignalHandling(t *testing.T) {
	// Create channels for test coordination
	ready := make(chan struct{})
	done := make(chan struct{})

	// Capture exit codes for verification
	var exitCode int
	var exitMu sync.Mutex
	origExit := osExit
	defer func() {
		osExit = origExit
	}()

	// Mock the exit function to capture exit codes instead of terminating
	osExit = func(code int) {
		exitMu.Lock()
		exitCode = code
		exitMu.Unlock()
		close(done)
	}

	// Start the main program in a goroutine
	go func() {
		// Signal test framework when initialization is done
		defer close(ready)
		main()
	}()

	// Wait for program initialization with timeout
	select {
	case <-ready:
		// Program initialized successfully
	case <-time.After(testInitTimeout):
		t.Fatal("Program failed to initialize within timeout")
	}

	// Let it run briefly to ensure stable operation
	time.Sleep(testRunTime)

	// Send termination signal
	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err, "Should be able to find current process")
	require.NoError(t, p.Signal(os.Interrupt), "Should be able to send interrupt signal")

	// Wait for clean shutdown
	select {
	case <-done:
		exitMu.Lock()
		code := exitCode
		exitMu.Unlock()
		assert.Equal(t, 0, code, "Program should exit with success code")
	case <-time.After(testShutdownTimeout):
		t.Fatal("Program failed to shut down within timeout")
	}
}

// TestLogger verifies that the logger is properly configured based
// on the environment setting and handles synchronization gracefully.
func TestLogger(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		wantLevel   zapcore.Level
	}{
		{
			name:        "development logger",
			environment: "development",
			wantLevel:   zap.DebugLevel,
		},
		{
			name:        "production logger",
			environment: "production",
			wantLevel:   zap.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment state
			prevEnv := os.Getenv("ENVIRONMENT")
			os.Setenv("ENVIRONMENT", tt.environment)
			defer os.Setenv("ENVIRONMENT", prevEnv)

			// Create and verify logger
			logger, err := setupLogger()
			require.NoError(t, err, "Logger setup should succeed")
			defer func() {
				_ = safeSync(logger)
			}()

			assert.Equal(t, tt.wantLevel, getLoggerLevel(logger),
				"Logger should have correct level for environment")

			// Verify sync behavior
			err = safeSync(logger)
			assert.NoError(t, err, "Logger sync should handle common issues gracefully")
		})
	}
}

// Mock os.Exit for testing purposes
var osExit = os.Exit
