// Package agent implements the edge agent functionality
package agent

import (
    "time"

    "github.com/wrale/wrale-fleet/fleet/brain/types"
)

// AgentConfig defines the edge agent configuration
type AgentConfig struct {
    DeviceID      types.DeviceID
    BrainEndpoint string
    UpdateInterval time.Duration
    MetalEndpoint string
}

// AgentState represents the current state of the edge agent
type AgentState struct {
    DeviceState types.DeviceState
    LastSync    time.Time
    LastError   error
    IsHealthy   bool
    Mode        OperationMode
}

// OperationMode represents the agent's current operating mode
type OperationMode string

const (
    // ModeNormal indicates normal operation with brain connectivity
    ModeNormal OperationMode = "normal"
    
    // ModeAutonomous indicates autonomous operation without brain connectivity
    ModeAutonomous OperationMode = "autonomous"
    
    // ModeSafe indicates restricted operation due to issues
    ModeSafe OperationMode = "safe"
)

// CommandType identifies different types of commands
type CommandType string

const (
    CmdUpdateState    CommandType = "update_state"
    CmdExecuteTask    CommandType = "execute_task"
    CmdUpdateConfig   CommandType = "update_config"
    CmdEnterSafeMode  CommandType = "enter_safe_mode"
    CmdExitSafeMode   CommandType = "exit_safe_mode"
)

// Command represents a command to be executed by the agent
type Command struct {
    Type      CommandType
    Payload   interface{}
    Priority  int
    Deadline  time.Time
    ID        string
    CreatedAt time.Time
}

// CommandResult represents the result of a command execution
type CommandResult struct {
    CommandID string
    Success   bool
    Error     error
    Data      interface{}
    CompletedAt time.Time
}

// MetalClient defines the interface for interacting with the metal layer
type MetalClient interface {
    // Hardware operations
    GetMetrics() (types.DeviceMetrics, error)
    UpdatePowerState(state string) error
    UpdateThermalConfig(config map[string]interface{}) error
    
    // Command execution
    ExecuteOperation(operation string) error
    GetOperationStatus(operationID string) (string, error)
    
    // Health and diagnostics
    GetHealthStatus() (bool, error)
    RunDiagnostics() (map[string]interface{}, error)
}

// BrainClient defines the interface for communicating with the fleet brain
type BrainClient interface {
    // State management
    SyncState(state types.DeviceState) error
    GetCommands() ([]Command, error)
    ReportCommandResult(result CommandResult) error
    
    // Health reporting
    ReportHealth(healthy bool, details map[string]interface{}) error
    
    // Configuration
    GetConfig() (map[string]interface{}, error)
}

// StateStore defines the interface for persistent state storage
type StateStore interface {
    // State operations
    GetState() (AgentState, error)
    UpdateState(state AgentState) error
    
    // Command history
    GetCommandHistory() ([]CommandResult, error)
    AddCommandResult(result CommandResult) error
    
    // Configuration
    GetConfig() (map[string]interface{}, error)
    UpdateConfig(config map[string]interface{}) error
}