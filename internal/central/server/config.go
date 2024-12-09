package server

import (
	"fmt"
	"strconv"
)

// Config holds the server configuration.
type Config struct {
	// Port is the server listening port for primary API endpoints
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
)

// ManagementConfig holds configuration for the management API endpoints
type ManagementConfig struct {
	// Port for management API endpoints (health, readiness)
	// This must be explicitly configured for proper security setup
	Port string

	// ExposureLevel controls how much information is exposed in health endpoints
	ExposureLevel ExposureLevel
}

// Validate checks the configuration for errors and ensures all required values
// have been properly configured. No default values are provided for security-
// sensitive settings like ports to ensure explicit configuration in production.
func (c *Config) Validate() error {
	// Require and validate main API port
	if c.Port == "" {
		return fmt.Errorf("port must be specified")
	}
	if _, err := strconv.Atoi(c.Port); err != nil {
		return fmt.Errorf("invalid port number: %s", c.Port)
	}

	// Set basic defaults that don't impact security
	if c.DataDir == "" {
		c.DataDir = defaultDataDir
	}
	if c.LogLevel == "" {
		c.LogLevel = defaultLogLevel
	}

	// Require management configuration
	if c.ManagementConfig == nil {
		return fmt.Errorf("management configuration must be provided")
	}

	// Require and validate management port
	if c.ManagementConfig.Port == "" {
		return fmt.Errorf("management port must be specified")
	}
	if _, err := strconv.Atoi(c.ManagementConfig.Port); err != nil {
		return fmt.Errorf("invalid management port number: %s", c.ManagementConfig.Port)
	}

	// Set default exposure level if not specified
	if c.ManagementConfig.ExposureLevel == "" {
		c.ManagementConfig.ExposureLevel = ExposureStandard
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
