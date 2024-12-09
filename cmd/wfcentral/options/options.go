// Package options provides configuration and initialization for the wfcentral command.
package options

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/wrale/wrale-fleet/internal/central/server"
	"github.com/wrale/wrale-fleet/internal/fleet/logging"
	"github.com/wrale/wrale-fleet/internal/fleet/logging/store/memory"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds the command-line options for wfcentral.
// This separates command-line concerns from the core server configuration.
type Config struct {
	// Port is the main HTTP server port for device management APIs
	Port string

	// DataDir is the path for persistent storage
	DataDir string

	// Logging configuration
	LogLevel string // Logging level (debug, info, warn, error)
	LogFile  string // Log file path (empty for stdout)
	LogJSON  bool   // Enable JSON log format
	LogStage int    // Stage-aware logging (1-6)

	// ManagementPort is the port for health and readiness endpoints
	// This must be explicitly configured for proper security setup
	ManagementPort string

	// HealthExposure controls how much information is exposed in health endpoints
	// Valid values are: "minimal", "standard", "full"
	// - minimal: Only basic health status
	// - standard: Includes version and uptime (default)
	// - full: All available health information
	HealthExposure string
}

// New creates a new Config with sensible default values that prioritize security
// while requiring explicit port configuration. The management port must be
// explicitly set at runtime, so we don't default it here.
func New() *Config {
	return &Config{
		Port:           "8600",               // Default main API port
		DataDir:        "/var/lib/wfcentral", // Default data directory
		LogLevel:       "info",               // Default log level
		LogStage:       1,                    // Default to Stage 1 capabilities
		HealthExposure: "standard",           // Default to standard health information exposure
	}
}

// NewServer creates and configures a central server instance based on
// the provided configuration options. This method handles the initialization
// of all necessary components including logging, monitoring, and the separate
// management server for health endpoints.
func NewServer(cfg *Config) (*server.Server, error) {
	// Basic validation
	basePort, err := strconv.Atoi(cfg.Port)
	if err != nil {
		return nil, fmt.Errorf("invalid port number: %s", cfg.Port)
	}

	// Management port must be explicitly configured
	if cfg.ManagementPort == "" {
		return nil, fmt.Errorf("management-port must be specified (use --management-port flag)")
	}

	// Validate management port
	mgmtPort, err := strconv.Atoi(cfg.ManagementPort)
	if err != nil {
		return nil, fmt.Errorf("invalid management port number: %s", cfg.ManagementPort)
	}

	// Ensure ports are different
	if basePort == mgmtPort {
		return nil, fmt.Errorf("management port must be different from main API port")
	}

	// Initialize the logging service
	loggingStore := memory.New()
	loggingService, err := logging.NewService(loggingStore, nil,
		logging.WithRetentionPolicy(logging.EventSystem, 30*24*time.Hour),      // 30 days for system events
		logging.WithRetentionPolicy(logging.EventSecurity, 90*24*time.Hour),    // 90 days for security events
		logging.WithRetentionPolicy(logging.EventAudit, 365*24*time.Hour),      // 1 year for audit events
		logging.WithRetentionPolicy(logging.EventCompliance, 730*24*time.Hour), // 2 years for compliance events
	)
	if err != nil {
		return nil, fmt.Errorf("initializing logging service: %w", err)
	}

	// Convert log level
	var level zapcore.Level
	switch cfg.LogLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create encoder based on format preference
	var encoder zapcore.Encoder
	if cfg.LogJSON {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Configure output
	var output zapcore.WriteSyncer
	if cfg.LogFile != "" {
		f, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, fmt.Errorf("opening log file: %w", err)
		}
		output = zapcore.AddSync(f)
	} else {
		output = zapcore.AddSync(os.Stdout)
	}

	// Create the logger
	core := zapcore.NewCore(encoder, output, level)
	logger := zap.New(core,
		zap.AddCaller(),
		zap.Fields(
			zap.Int("stage", cfg.LogStage),
			zap.String("component", "wfcentral"),
		),
	)

	// Create internal server configuration
	serverConfig := &server.Config{
		Port:     cfg.Port,
		DataDir:  cfg.DataDir,
		LogLevel: cfg.LogLevel,
		ManagementConfig: &server.ManagementConfig{
			Port:          cfg.ManagementPort,
			ExposureLevel: server.ExposureLevel(cfg.HealthExposure),
		},
		LoggingService: loggingService,
	}

	// Create and validate server instance
	srv, err := server.New(serverConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("initializing server: %w", err)
	}

	return srv, nil
}

// ValidateHealthExposure checks if the given exposure level is valid.
// This helper function can be used by CLI commands to validate user input
// before attempting server creation.
func ValidateHealthExposure(level string) bool {
	switch server.ExposureLevel(level) {
	case server.ExposureMinimal, server.ExposureStandard, server.ExposureFull:
		return true
	default:
		return false
	}
}
