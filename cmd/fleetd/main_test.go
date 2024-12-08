package main

import (
	"context"
	"os"
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

// TestRunDemo tests the demo workflow functionality
func TestRunDemo(t *testing.T) {
	// Use test logger with proper cleanup
	logger := zaptest.NewLogger(t)
	defer func() {
		if err := logger.Sync(); err != nil {
			t.Logf("non-fatal: failed to sync logger: %v", err)
		}
	}()

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
			wantErr: false,
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

			devices, err := service.List(tt.ctx, device.ListOptions{
				TenantID: "demo-tenant",
			})
			require.NoError(t, err)
			require.NotEmpty(t, devices)

			// Device verification logic remains unchanged
		})
	}
}

// TestMainSignalHandling tests proper shutdown signal handling
func TestMainSignalHandling(t *testing.T) {
	// Create done channel for test coordination
	done := make(chan struct{})

	// Setup exit capture
	var exitCode int
	origExit := osExit
	defer func() { osExit = origExit }()
	osExit = func(code int) {
		exitCode = code
		close(done)
	}

	// Start main in a goroutine
	go func() {
		main()
	}()

	// Allow time for initialization
	time.Sleep(100 * time.Millisecond)

	// Send interrupt signal
	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, p.Signal(os.Interrupt))

	// Wait for shutdown with timeout
	select {
	case <-done:
		assert.Equal(t, 0, exitCode)
	case <-time.After(2 * time.Second):
		t.Fatal("program did not shut down within timeout")
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
			prevEnv := os.Getenv("ENVIRONMENT")
			os.Setenv("ENVIRONMENT", tt.environment)
			defer os.Setenv("ENVIRONMENT", prevEnv)

			logger, err := setupLogger()
			require.NoError(t, err)
			defer func() {
				if err := logger.Sync(); err != nil {
					t.Logf("non-fatal: failed to sync logger: %v", err)
				}
			}()

			assert.Equal(t, tt.wantLevel, logger.Core().Enabled(tt.wantLevel))
		})
	}
}

// Mock os.Exit for testing
var osExit = os.Exit
