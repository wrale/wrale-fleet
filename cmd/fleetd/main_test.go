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

const (
	testInitTimeout     = 500 * time.Millisecond
	testShutdownTimeout = 5 * time.Second
)

// TestRunDemo tests the demo workflow functionality
func TestRunDemo(t *testing.T) {
	// Use test logger with proper cleanup
	logger := zaptest.NewLogger(t)
	defer func() {
		_ = safeSync(logger)
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
		})
	}
}

// TestMainSignalHandling tests proper shutdown signal handling
func TestMainSignalHandling(t *testing.T) {
	// Create coordination channels
	ready := make(chan struct{})
	done := make(chan struct{})

	// Setup exit capture
	var exitCode int
	var exitMu sync.Mutex
	origExit := osExit
	defer func() { osExit = origExit }()
	osExit = func(code int) {
		exitMu.Lock()
		exitCode = code
		exitMu.Unlock()
		close(done)
	}()

	// Start main in a goroutine with initialization signal
	go func() {
		// Signal when initialization is complete
		time.AfterFunc(testInitTimeout/2, func() {
			close(ready)
		})
		main()
	}()

	// Wait for initialization with timeout
	select {
	case <-ready:
		// Initialized successfully
	case <-time.After(testInitTimeout):
		t.Fatal("program did not initialize within timeout")
	}

	// Send interrupt signal
	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, p.Signal(os.Interrupt))

	// Wait for shutdown with timeout
	select {
	case <-done:
		exitMu.Lock()
		code := exitCode
		exitMu.Unlock()
		assert.Equal(t, 0, code)
	case <-time.After(testShutdownTimeout):
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
				_ = safeSync(logger)
			}()

			assert.Equal(t, tt.wantLevel, getLoggerLevel(logger))
		})
	}
}

// Mock os.Exit for testing
var osExit = os.Exit
