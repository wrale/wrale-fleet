package server

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
}

// Stage1Config holds configuration specific to Stage 1 capabilities.
type Stage1Config struct {
	// DeviceStorageType specifies the device storage backend (e.g., "memory", "postgres")
	DeviceStorageType string

	// Additional Stage 1 specific settings can be added here
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if c.Port == "" {
		c.Port = defaultPort
	}

	if c.DataDir == "" {
		c.DataDir = defaultDataDir
	}

	if c.LogLevel == "" {
		c.LogLevel = defaultLogLevel
	}

	return nil
}
