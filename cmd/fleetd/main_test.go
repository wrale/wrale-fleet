package main

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/fleet/internal/fleet/device"
	"github.com/wrale/fleet/internal/fleet/device/store/memory"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// TestRunDemo tests the demo workflow functionality
func TestRunDemo(t *testing.T) {
	// Use test logger
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
		setup   func(*device.Service) error
	}{
		{
			name:    "successful demo execution",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name: "context cancellation",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Pre-cancel context
				return ctx
			}(),
			wantErr: true,
		},
		{
			name: "device already exists",
			ctx:  context.Background(),
			setup: func(s *device.Service) error {
				// Pre-create device with same tenant ID
				_, err := s.Register(context.Background(), "demo-tenant", "Existing Device")
				return err
			},
			wantErr: false, // Should handle existing device gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := memory.New()
			service := device.NewService(store, logger)

			if tt.setup != nil {
				require.NoError(t, tt.setup(service))
			}

			err := runDemo(tt.ctx, service, logger)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			// Verify device state after demo
			devices, err := service.List(tt.ctx, device.ListOptions{
				TenantID: "demo-tenant",
			})
			require.NoError(t, err)
			require.NotEmpty(t, devices)

			dev := devices[0]
			assert.NotEmpty(t, dev.ID)
			assert.Equal(t, "Demo Raspberry Pi", dev.Name)
			assert.Equal(t, device.StatusOnline, dev.Status)

			// Verify tags
			assert.Equal(t, "production", dev.Tags["environment"])
			assert.Equal(t, "datacenter-1", dev.Tags["location"])

			// Verify config
			var config map[string]interface{}
			require.NoError(t, json.Unmarshal(dev.Config, &config))
			assert.Equal(t, "30s", config["monitoring_interval"])
			assert.Equal(t, "info", config["log_level"])

			features, ok := config["features"].(map[string]interface{})
			require.True(t, ok)
			assert.True(t, features["metrics_enabled"].(bool))
			assert.False(t, features["tracing_enabled"].(bool))
			assert.True(t, features["alerting_enabled"].(bool))

			// Verify network info
			require.NotNil(t, dev.NetworkInfo)
			assert.Equal(t, "192.168.1.100", dev.NetworkInfo.IPAddress)
			assert.Equal(t, "00:11:22:33:44:55", dev.NetworkInfo.MACAddress)
			assert.Equal(t, "demo-device-1", dev.NetworkInfo.Hostname)
			assert.Equal(t, 9100, dev.NetworkInfo.Port)

			// Verify offline capabilities
			require.NotNil(t, dev.OfflineCapabilities)
			assert.True(t, dev.OfflineCapabilities.SupportsAirgap)
			assert.Equal(t, time.Hour, dev.OfflineCapabilities.SyncInterval)
			assert.NotZero(t, dev.OfflineCapabilities.LastSyncTime)
			assert.Contains(t, dev.OfflineCapabilities.OfflineOperations, "status_update")
			assert.Equal(t, int64(104857600), dev.OfflineCapabilities.LocalBufferSize) // 100MB
		})
	}
}

// TestMainSignalHandling tests proper shutdown signal handling
func TestMainSignalHandling(t *testing.T) {
	// Create a channel to coordinate test completion
	done := make(chan struct{})

	// Create a WaitGroup to ensure goroutine completion
	var wg sync.WaitGroup
	wg.Add(1)

	// Replace os.Exit
	origExit := osExit
	defer func() { osExit = origExit }()

	var exitCode int
	osExit = func(code int) {
		exitCode = code
		wg.Done()
	}

	// Start the program in a goroutine
	go func() {
		defer close(done)
		main()
	}()

	// Allow some time for initialization
	time.Sleep(100 * time.Millisecond)

	// Send termination signal
	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, p.Signal(os.Interrupt))

	// Wait for cleanup and exit
	wg.Wait()

	// Verify clean exit
	assert.Equal(t, 0, exitCode)

	// Ensure program terminates
	select {
	case <-done:
		// Success - program terminated
	case <-time.After(5 * time.Second):
		t.Fatal("program did not terminate within timeout")
	}
}

// TestLogger verifies logger configuration
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
			// Set environment
			prevEnv := os.Getenv("ENVIRONMENT")
			os.Setenv("ENVIRONMENT", tt.environment)
			defer os.Setenv("ENVIRONMENT", prevEnv)

			// Create logger
			logger, err := setupLogger()
			require.NoError(t, err)
			defer logger.Sync()

			// Verify logger level
			assert.Equal(t, tt.wantLevel, logger.Core().Enabled(tt.wantLevel))
		})
	}
}

// Mock os.Exit for testing
var osExit = os.Exit
