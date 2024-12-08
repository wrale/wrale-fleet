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
	executor *commandExecutor
}

// NewDeviceRegistrationScenario creates a new device registration demo
func NewDeviceRegistrationScenario(logger *zap.Logger, wfcentralPath string) (*DeviceRegistrationScenario, error) {
	executor, err := newCommandExecutor(wfcentralPath)
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
	steps := []struct {
		name string
		args []string
	}{
		{
			name: "Register Device",
			args: []string{"device", "register", "--name", "demo-device-1"},
		},
		{
			name: "Verify Registration",
			args: []string{"device", "get", "demo-device-1"},
		},
	}

	for _, step := range steps {
		s.base.Logger().Info("executing step", zap.String("step", step.name))

		if err := s.executor.executeCommand(ctx, step.args); err != nil {
			return fmt.Errorf("step %s failed: %w", step.name, err)
		}

		time.Sleep(time.Second)
	}

	return nil
}

// StatusMonitoringScenario demonstrates device status monitoring
type StatusMonitoringScenario struct {
	base     *scenario.BaseScenario
	executor *commandExecutor
}

// NewStatusMonitoringScenario creates a new status monitoring demo
func NewStatusMonitoringScenario(logger *zap.Logger, wfcentralPath string) (*StatusMonitoringScenario, error) {
	executor, err := newCommandExecutor(wfcentralPath)
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
	steps := []struct {
		name string
		args []string
	}{
		{
			name: "View Device Status",
			args: []string{"device", "status", "demo-device-1"},
		},
		{
			name: "Monitor Health Metrics",
			args: []string{"device", "health", "demo-device-1"},
		},
		{
			name: "Check Alert History",
			args: []string{"device", "alerts", "demo-device-1"},
		},
	}

	for _, step := range steps {
		s.base.Logger().Info("executing step", zap.String("step", step.name))

		if err := s.executor.executeCommand(ctx, step.args); err != nil {
			return fmt.Errorf("step %s failed: %w", step.name, err)
		}

		time.Sleep(time.Second * 2)
	}

	return nil
}

// ConfigurationScenario demonstrates device configuration management
type ConfigurationScenario struct {
	base     *scenario.BaseScenario
	executor *commandExecutor
}

// NewConfigurationScenario creates a new configuration management demo
func NewConfigurationScenario(logger *zap.Logger, wfcentralPath string) (*ConfigurationScenario, error) {
	executor, err := newCommandExecutor(wfcentralPath)
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
	steps := []struct {
		name string
		args []string
	}{
		{
			name: "View Current Config",
			args: []string{"device", "config", "get", "demo-device-1"},
		},
		{
			name: "Update Config",
			args: []string{"device", "config", "set", "demo-device-1", "--file", "demo-config.json"},
		},
		{
			name: "Verify Config Update",
			args: []string{"device", "config", "get", "demo-device-1"},
		},
		{
			name: "View Config History",
			args: []string{"device", "config", "history", "demo-device-1"},
		},
	}

	for _, step := range steps {
		s.base.Logger().Info("executing step", zap.String("step", step.name))

		if err := s.executor.executeCommand(ctx, step.args); err != nil {
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
