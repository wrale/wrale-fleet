package sysadmin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"go.uber.org/zap"
)

// DeviceRegistrationScenario demonstrates basic device registration workflow
type DeviceRegistrationScenario struct {
	*BaseScenario
	wfcentralPath string
}

// NewDeviceRegistrationScenario creates a new device registration demo
func NewDeviceRegistrationScenario(logger *zap.Logger, wfcentralPath string) *DeviceRegistrationScenario {
	return &DeviceRegistrationScenario{
		BaseScenario: NewBaseScenario(
			"device-registration",
			"Demonstrates registering a new device with wfcentral",
			logger,
		),
		wfcentralPath: wfcentralPath,
	}
}

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
		s.logger.Info("executing step", zap.String("step", step.name))

		cmd := exec.CommandContext(ctx, step.command, step.args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("step %s failed: %w", step.name, err)
		}

		// Add a small delay for readability
		time.Sleep(time.Second)
	}

	return nil
}

// StatusMonitoringScenario demonstrates device status monitoring
type StatusMonitoringScenario struct {
	*BaseScenario
	wfcentralPath string
}

// NewStatusMonitoringScenario creates a new status monitoring demo
func NewStatusMonitoringScenario(logger *zap.Logger, wfcentralPath string) *StatusMonitoringScenario {
	return &StatusMonitoringScenario{
		BaseScenario: NewBaseScenario(
			"status-monitoring",
			"Demonstrates monitoring device status and health",
			logger,
		),
		wfcentralPath: wfcentralPath,
	}
}

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
		s.logger.Info("executing step", zap.String("step", step.name))

		cmd := exec.CommandContext(ctx, step.command, step.args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("step %s failed: %w", step.name, err)
		}

		// Add a small delay for readability
		time.Sleep(time.Second * 2)
	}

	return nil
}

// ConfigurationScenario demonstrates device configuration management
type ConfigurationScenario struct {
	*BaseScenario
	wfcentralPath string
}

// NewConfigurationScenario creates a new configuration management demo
func NewConfigurationScenario(logger *zap.Logger, wfcentralPath string) *ConfigurationScenario {
	return &ConfigurationScenario{
		BaseScenario: NewBaseScenario(
			"configuration-management",
			"Demonstrates device configuration workflows",
			logger,
		),
		wfcentralPath: wfcentralPath,
	}
}

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
		s.logger.Info("executing step", zap.String("step", step.name))

		cmd := exec.CommandContext(ctx, step.command, step.args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("step %s failed: %w", step.name, err)
		}

		// Add a small delay for readability
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
