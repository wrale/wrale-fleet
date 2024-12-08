package sysadmin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"go.uber.org/zap"
)

// Scenario represents a single demonstration scenario
type Scenario interface {
	// Name returns the scenario's identifier
	Name() string

	// Description returns a human-readable description of what the scenario demonstrates
	Description() string

	// Setup prepares any necessary resources for the scenario
	Setup(ctx context.Context) error

	// Run executes the scenario's demonstration
	Run(ctx context.Context) error

	// Cleanup releases any resources created during setup
	Cleanup(ctx context.Context) error
}

// BaseScenario provides a common implementation of the Scenario interface
type BaseScenario struct {
	name        string
	description string
	logger      *zap.Logger
}

// NewBaseScenario creates a new base scenario with the given name and description
func NewBaseScenario(name, description string, logger *zap.Logger) *BaseScenario {
	return &BaseScenario{
		name:        name,
		description: description,
		logger:      logger,
	}
}

// Name returns the scenario's name
func (b *BaseScenario) Name() string {
	return b.name
}

// Description returns the scenario's description
func (b *BaseScenario) Description() string {
	return b.description
}

// Setup provides a default no-op implementation
func (b *BaseScenario) Setup(ctx context.Context) error {
	return nil
}

// Run must be implemented by concrete scenarios
func (b *BaseScenario) Run(ctx context.Context) error {
	return fmt.Errorf("Run not implemented for scenario %s", b.name)
}

// Cleanup provides a default no-op implementation
func (b *BaseScenario) Cleanup(ctx context.Context) error {
	return nil
}

// DeviceRegistrationScenario demonstrates basic device registration workflow
type DeviceRegistrationScenario struct {
	base          *BaseScenario
	wfcentralPath string
}

// NewDeviceRegistrationScenario creates a new device registration demo
func NewDeviceRegistrationScenario(logger *zap.Logger, wfcentralPath string) *DeviceRegistrationScenario {
	return &DeviceRegistrationScenario{
		base:          NewBaseScenario("device-registration", "Demonstrates registering a new device with wfcentral", logger),
		wfcentralPath: wfcentralPath,
	}
}

func (s *DeviceRegistrationScenario) Name() string                      { return s.base.Name() }
func (s *DeviceRegistrationScenario) Description() string               { return s.base.Description() }
func (s *DeviceRegistrationScenario) Setup(ctx context.Context) error   { return s.base.Setup(ctx) }
func (s *DeviceRegistrationScenario) Cleanup(ctx context.Context) error { return s.base.Cleanup(ctx) }

func (s *DeviceRegistrationScenario) Run(ctx context.Context) error {
	steps := []struct {
		name    string
		command string
		args    []string
	}{
		{
			name:    "Register Device",
			command: s.wfcentralPath,
			args:    []string{"device", "register", "--name", "demo-device-1"},
		},
		{
			name:    "Verify Registration",
			command: s.wfcentralPath,
			args:    []string{"device", "get", "demo-device-1"},
		},
	}

	for _, step := range steps {
		s.base.logger.Info("executing step", zap.String("step", step.name))

		cmd := exec.CommandContext(ctx, step.command, step.args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("step %s failed: %w", step.name, err)
		}

		time.Sleep(time.Second)
	}

	return nil
}

// StatusMonitoringScenario demonstrates device status monitoring
type StatusMonitoringScenario struct {
	base          *BaseScenario
	wfcentralPath string
}

// NewStatusMonitoringScenario creates a new status monitoring demo
func NewStatusMonitoringScenario(logger *zap.Logger, wfcentralPath string) *StatusMonitoringScenario {
	return &StatusMonitoringScenario{
		base:          NewBaseScenario("status-monitoring", "Demonstrates monitoring device status and health", logger),
		wfcentralPath: wfcentralPath,
	}
}

func (s *StatusMonitoringScenario) Name() string                      { return s.base.Name() }
func (s *StatusMonitoringScenario) Description() string               { return s.base.Description() }
func (s *StatusMonitoringScenario) Setup(ctx context.Context) error   { return s.base.Setup(ctx) }
func (s *StatusMonitoringScenario) Cleanup(ctx context.Context) error { return s.base.Cleanup(ctx) }

func (s *StatusMonitoringScenario) Run(ctx context.Context) error {
	steps := []struct {
		name    string
		command string
		args    []string
	}{
		{
			name:    "View Device Status",
			command: s.wfcentralPath,
			args:    []string{"device", "status", "demo-device-1"},
		},
		{
			name:    "Monitor Health Metrics",
			command: s.wfcentralPath,
			args:    []string{"device", "health", "demo-device-1"},
		},
		{
			name:    "Check Alert History",
			command: s.wfcentralPath,
			args:    []string{"device", "alerts", "demo-device-1"},
		},
	}

	for _, step := range steps {
		s.base.logger.Info("executing step", zap.String("step", step.name))

		cmd := exec.CommandContext(ctx, step.command, step.args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("step %s failed: %w", step.name, err)
		}

		time.Sleep(time.Second * 2)
	}

	return nil
}

// ConfigurationScenario demonstrates device configuration management
type ConfigurationScenario struct {
	base          *BaseScenario
	wfcentralPath string
}

// NewConfigurationScenario creates a new configuration management demo
func NewConfigurationScenario(logger *zap.Logger, wfcentralPath string) *ConfigurationScenario {
	return &ConfigurationScenario{
		base:          NewBaseScenario("configuration-management", "Demonstrates device configuration workflows", logger),
		wfcentralPath: wfcentralPath,
	}
}

func (s *ConfigurationScenario) Name() string                      { return s.base.Name() }
func (s *ConfigurationScenario) Description() string               { return s.base.Description() }
func (s *ConfigurationScenario) Setup(ctx context.Context) error   { return s.base.Setup(ctx) }
func (s *ConfigurationScenario) Cleanup(ctx context.Context) error { return s.base.Cleanup(ctx) }

func (s *ConfigurationScenario) Run(ctx context.Context) error {
	steps := []struct {
		name    string
		command string
		args    []string
	}{
		{
			name:    "View Current Config",
			command: s.wfcentralPath,
			args:    []string{"device", "config", "get", "demo-device-1"},
		},
		{
			name:    "Update Config",
			command: s.wfcentralPath,
			args:    []string{"device", "config", "set", "demo-device-1", "--file", "demo-config.json"},
		},
		{
			name:    "Verify Config Update",
			command: s.wfcentralPath,
			args:    []string{"device", "config", "get", "demo-device-1"},
		},
		{
			name:    "View Config History",
			command: s.wfcentralPath,
			args:    []string{"device", "config", "history", "demo-device-1"},
		},
	}

	for _, step := range steps {
		s.base.logger.Info("executing step", zap.String("step", step.name))

		cmd := exec.CommandContext(ctx, step.command, step.args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("step %s failed: %w", step.name, err)
		}

		time.Sleep(time.Second * 2)
	}

	return nil
}

// Stage1Scenarios returns all Stage 1 scenarios for the SysAdmin persona
func Stage1Scenarios(logger *zap.Logger, wfcentralPath string) []Scenario {
	return []Scenario{
		NewDeviceRegistrationScenario(logger, wfcentralPath),
		NewStatusMonitoringScenario(logger, wfcentralPath),
		NewConfigurationScenario(logger, wfcentralPath),
	}
}
