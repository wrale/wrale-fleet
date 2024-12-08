// Package sysadmin provides secure command execution for fleet management demos
package sysadmin

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/wrale/fleet/internal/demo/scenario"
)

// DeviceRegistrationScenario demonstrates basic device registration workflow
type DeviceRegistrationScenario struct {
	base     *scenario.BaseScenario
	executor *CommandExecutor
}

// NewDeviceRegistrationScenario creates a new device registration demo
func NewDeviceRegistrationScenario(logger *zap.Logger, wfcentralPath string) (*DeviceRegistrationScenario, error) {
	executor, err := NewCommandExecutor(wfcentralPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create command executor: %w", err)
	}

	return &DeviceRegistrationScenario{
		base:     scenario.NewBaseScenario("device-registration", "Demonstrates registering a new device with wfcentral", logger),
		executor: executor,
	}, nil
}

func (s *DeviceRegistrationScenario) Name() string                      { return s.base.Name() }
func (s *DeviceRegistrationScenario) Description() string               { return s.base.Description() }
func (s *DeviceRegistrationScenario) Setup(ctx context.Context) error   { return s.base.Setup(ctx) }
func (s *DeviceRegistrationScenario) Cleanup(ctx context.Context) error { return s.base.Cleanup(ctx) }

func (s *DeviceRegistrationScenario) Run(ctx context.Context) error {
	const deviceName = "demo-device-1"

	steps := []struct {
		name string
		cmd  func() (Command, error)
	}{
		{
			name: "Register Device",
			cmd:  func() (Command, error) { return s.executor.NewRegisterDeviceCommand(deviceName) },
		},
		{
			name: "Verify Registration",
			cmd:  func() (Command, error) { return s.executor.NewGetDeviceCommand(deviceName) },
		},
	}

	for _, step := range steps {
		s.base.Logger().Info("executing step", zap.String("step", step.name))

		cmd, err := step.cmd()
		if err != nil {
			return fmt.Errorf("failed to create command for step %s: %w", step.name, err)
		}

		if err := cmd.Execute(ctx); err != nil {
			return fmt.Errorf("step %s failed: %w", step.name, err)
		}

		time.Sleep(time.Second)
	}

	return nil
}

// StatusMonitoringScenario demonstrates device status monitoring
type StatusMonitoringScenario struct {
	base     *scenario.BaseScenario
	executor *CommandExecutor
}

// NewStatusMonitoringScenario creates a new status monitoring demo
func NewStatusMonitoringScenario(logger *zap.Logger, wfcentralPath string) (*StatusMonitoringScenario, error) {
	executor, err := NewCommandExecutor(wfcentralPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create command executor: %w", err)
	}

	return &StatusMonitoringScenario{
		base:     scenario.NewBaseScenario("status-monitoring", "Demonstrates monitoring device status and health", logger),
		executor: executor,
	}, nil
}

func (s *StatusMonitoringScenario) Name() string                      { return s.base.Name() }
func (s *StatusMonitoringScenario) Description() string               { return s.base.Description() }
func (s *StatusMonitoringScenario) Setup(ctx context.Context) error   { return s.base.Setup(ctx) }
func (s *StatusMonitoringScenario) Cleanup(ctx context.Context) error { return s.base.Cleanup(ctx) }

func (s *StatusMonitoringScenario) Run(ctx context.Context) error {
	const deviceName = "demo-device-1"

	steps := []struct {
		name string
		cmd  func() (Command, error)
	}{
		{
			name: "View Device Status",
			cmd:  func() (Command, error) { return s.executor.NewDeviceStatusCommand(deviceName) },
		},
		{
			name: "Monitor Health Metrics",
			cmd:  func() (Command, error) { return s.executor.NewDeviceHealthCommand(deviceName) },
		},
		{
			name: "Check Alert History",
			cmd:  func() (Command, error) { return s.executor.NewDeviceAlertsCommand(deviceName) },
		},
	}

	for _, step := range steps {
		s.base.Logger().Info("executing step", zap.String("step", step.name))

		cmd, err := step.cmd()
		if err != nil {
			return fmt.Errorf("failed to create command for step %s: %w", step.name, err)
		}

		if err := cmd.Execute(ctx); err != nil {
			return fmt.Errorf("step %s failed: %w", step.name, err)
		}

		time.Sleep(time.Second * 2)
	}

	return nil
}

// ConfigurationScenario demonstrates device configuration management
type ConfigurationScenario struct {
	base     *scenario.BaseScenario
	executor *CommandExecutor
}

// NewConfigurationScenario creates a new configuration management demo
func NewConfigurationScenario(logger *zap.Logger, wfcentralPath string) (*ConfigurationScenario, error) {
	executor, err := NewCommandExecutor(wfcentralPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create command executor: %w", err)
	}

	return &ConfigurationScenario{
		base:     scenario.NewBaseScenario("configuration-management", "Demonstrates device configuration workflows", logger),
		executor: executor,
	}, nil
}

func (s *ConfigurationScenario) Name() string                      { return s.base.Name() }
func (s *ConfigurationScenario) Description() string               { return s.base.Description() }
func (s *ConfigurationScenario) Setup(ctx context.Context) error   { return s.base.Setup(ctx) }
func (s *ConfigurationScenario) Cleanup(ctx context.Context) error { return s.base.Cleanup(ctx) }

func (s *ConfigurationScenario) Run(ctx context.Context) error {
	const (
		deviceName = "demo-device-1"
		configFile = "demo-config.json"
	)

	steps := []struct {
		name string
		cmd  func() (Command, error)
	}{
		{
			name: "View Current Config",
			cmd:  func() (Command, error) { return s.executor.NewDeviceConfigGetCommand(deviceName) },
		},
		{
			name: "Update Config",
			cmd:  func() (Command, error) { return s.executor.NewDeviceConfigSetCommand(deviceName, configFile) },
		},
		{
			name: "Verify Config Update",
			cmd:  func() (Command, error) { return s.executor.NewDeviceConfigGetCommand(deviceName) },
		},
		{
			name: "View Config History",
			cmd:  func() (Command, error) { return s.executor.NewDeviceConfigHistoryCommand(deviceName) },
		},
	}

	for _, step := range steps {
		s.base.Logger().Info("executing step", zap.String("step", step.name))

		cmd, err := step.cmd()
		if err != nil {
			return fmt.Errorf("failed to create command for step %s: %w", step.name, err)
		}

		if err := cmd.Execute(ctx); err != nil {
			return fmt.Errorf("step %s failed: %w", step.name, err)
		}

		time.Sleep(time.Second * 2)
	}

	return nil
}

// Stage1Scenarios returns all Stage 1 scenarios for the SysAdmin persona
// It may return a partial list of scenarios if some fail to initialize
func Stage1Scenarios(logger *zap.Logger, wfcentralPath string) []scenario.Scenario {
	var scenarios []scenario.Scenario

	// Try to create each scenario, logging errors but continuing
	if reg, err := NewDeviceRegistrationScenario(logger, wfcentralPath); err != nil {
		logger.Error("failed to create device registration scenario", zap.Error(err))
	} else {
		scenarios = append(scenarios, reg)
	}

	if mon, err := NewStatusMonitoringScenario(logger, wfcentralPath); err != nil {
		logger.Error("failed to create status monitoring scenario", zap.Error(err))
	} else {
		scenarios = append(scenarios, mon)
	}

	if conf, err := NewConfigurationScenario(logger, wfcentralPath); err != nil {
		logger.Error("failed to create configuration scenario", zap.Error(err))
	} else {
		scenarios = append(scenarios, conf)
	}

	return scenarios
}
