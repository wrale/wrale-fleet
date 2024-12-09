package server

import (
	"fmt"
	"strconv"
)

// Config holds the server configuration.
type Config struct {
	// Port is the server listening port
	Port string

	// DataDir is the path to the data directory
	DataDir string

	// LogLevel sets the logging verbosity
	LogLevel string

	// Stage1Config holds Stage 1 specific configuration
	Stage1Config *Stage1Config

	// ManagementConfig holds configuration for the management API
	ManagementConfig *ManagementConfig
}

// Stage1Config holds configuration specific to Stage 1 capabilities.
type Stage1Config struct {
	// DeviceStorageType specifies the device storage backend (e.g., "memory", "postgres")
	DeviceStorageType string

	// Additional Stage 1 specific settings can be added here
}

// ExposureLevel defines how much information is exposed in health endpoints
type ExposureLevel string

const (
	// ExposureMinimal provides only basic health status
	ExposureMinimal ExposureLevel = "minimal"
	// ExposureStandard includes version and uptime information
	ExposureStandard ExposureLevel = "standard"
	// ExposureFull provides all available health information
	ExposureFull ExposureLevel = "full"

	// Default management port offset from main port
	defaultManagementPortOffset = 1
)

// ManagementConfig holds configuration for the management API endpoints
type ManagementConfig struct {
	// Port for management API endpoints (health, readiness)
	// If empty, defaults to main port + 1
	Port string

	// ExposureLevel controls how much information is exposed in health endpoints
	ExposureLevel ExposureLevel
}

// Validate checks the configuration for errors and ensures all required values
// are properly set with appropriate defaults.
func (c *Config) Validate() error {
	// Validate and default main port
	if c.Port == "" {
		c.Port = defaultPort
	}

	// Validate port is numeric
	mainPort, err := strconv.Atoi(c.Port)
	if err != nil {
		return fmt.Errorf("invalid port number: %s", c.Port)
	}

	// Set other basic defaults
	if c.DataDir == "" {
		c.DataDir = defaultDataDir
	}

	if c.LogLevel == "" {
		c.LogLevel = defaultLogLevel
	}

	// Initialize and validate management config
	if c.ManagementConfig == nil {
		c.ManagementConfig = &ManagementConfig{
			ExposureLevel: ExposureStandard,
		}
	}

	// Always set default values for management config to ensure it's complete
	if c.ManagementConfig.ExposureLevel == "" {
		c.ManagementConfig.ExposureLevel = ExposureStandard
	}

	// Set default management port based on main port if not specified
	if c.ManagementConfig.Port == "" {
		c.ManagementConfig.Port = strconv.Itoa(mainPort + defaultManagementPortOffset)
	}

	// Validate management port
	if _, err := strconv.Atoi(c.ManagementConfig.Port); err != nil {
		return fmt.Errorf("invalid management port number: %s", c.ManagementConfig.Port)
	}

	// Validate exposure level
	switch c.ManagementConfig.ExposureLevel {
	case ExposureMinimal, ExposureStandard, ExposureFull:
		// Valid exposure levels
	default:
		return fmt.Errorf("invalid exposure level: %s", c.ManagementConfig.ExposureLevel)
	}

	return nil
}
