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

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if c.Port == "" {
		c.Port = defaultPort
	}

	// Validate port is numeric
	if _, err := strconv.Atoi(c.Port); err != nil {
		return fmt.Errorf("invalid port number: %s", c.Port)
	}

	if c.DataDir == "" {
		c.DataDir = defaultDataDir
	}

	if c.LogLevel == "" {
		c.LogLevel = defaultLogLevel
	}

	// Initialize management config if not set
	if c.ManagementConfig == nil {
		c.ManagementConfig = &ManagementConfig{
			ExposureLevel: ExposureStandard,
		}
	}

	// Set default management port if not specified
	if c.ManagementConfig.Port == "" {
		basePort, _ := strconv.Atoi(c.Port)
		c.ManagementConfig.Port = strconv.Itoa(basePort + defaultManagementPortOffset)
	}

	// Validate management port
	if _, err := strconv.Atoi(c.ManagementConfig.Port); err != nil {
		return fmt.Errorf("invalid management port number: %s", c.ManagementConfig.Port)
	}

	// Validate exposure level
	switch c.ManagementConfig.ExposureLevel {
	case ExposureMinimal, ExposureStandard, ExposureFull:
		// Valid exposure levels
	case "":
		c.ManagementConfig.ExposureLevel = ExposureStandard
	default:
		return fmt.Errorf("invalid exposure level: %s", c.ManagementConfig.ExposureLevel)
	}

	return nil
}
