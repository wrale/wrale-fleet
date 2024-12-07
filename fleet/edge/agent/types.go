package agent

import (
    "time"

    "github.com/wrale/wrale-fleet/fleet/brain/types"
)

// Command types
const (
    CmdUpdateState    = "UPDATE_STATE"
    CmdExecuteTask    = "EXECUTE_TASK"
    CmdUpdateConfig   = "UPDATE_CONFIG"
    CmdEnterSafeMode  = "ENTER_SAFE_MODE"
    CmdExitSafeMode   = "EXIT_SAFE_MODE"
    
    // Thermal commands
    CmdUpdateThermalPolicy = "UPDATE_THERMAL_POLICY"
    CmdSetFanSpeed        = "SET_FAN_SPEED"
    CmdSetThrottling      = "SET_THROTTLING"
    CmdGetThermalState    = "GET_THERMAL_STATE"
)

// Operation modes
type OperationMode string

const (
    ModeNormal     OperationMode = "NORMAL"
    ModeSafe       OperationMode = "SAFE"
    ModeAutonomous OperationMode = "AUTONOMOUS"
)

// AgentConfig holds agent configuration
type AgentConfig struct {
    UpdateInterval time.Duration
    MetalEndpoint  string
    BrainEndpoint  string
}

// AgentState holds current agent state
type AgentState struct {
    DeviceState types.DeviceState
    Mode        OperationMode
    IsHealthy   bool
    LastError   error
    LastSync    time.Time
}

// Command represents a command to be executed
type Command struct {
    ID        string
    Type      string
    Payload   interface{}
    Timestamp time.Time
}

// CommandResult represents the result of a command execution
type CommandResult struct {
    CommandID   string
    Success     bool
    Error       error
    CompletedAt time.Time
    Payload     interface{} // Stores command-specific response data
}

// MetalClient interface for metal layer communication
type MetalClient interface {
    GetMetrics() (types.DeviceMetrics, error)
    GetThermalState() (*types.ThermalMetrics, error)  // Updated to use ThermalMetrics
    UpdateThermalPolicy(policy types.ThermalPolicy) error
    SetFanSpeed(speed uint32) error
    SetThrottling(enabled bool) error
    UpdatePowerState(state string) error
    ExecuteOperation(operation string) error
    GetHealthStatus() (bool, error)
    RunDiagnostics() (map[string]interface{}, error)
}

// BrainClient interface for brain communication
type BrainClient interface {
    GetCommands() ([]Command, error)
    SyncState(state types.DeviceState) error
    SyncThermalState(state *types.ThermalMetrics) error  // Updated to use ThermalMetrics
    ReportCommandResult(result CommandResult) error
    ReportHealth(healthy bool, diagnostics map[string]interface{}) error
}

// StateStore interface for persistent state storage
type StateStore interface {
    GetState() (AgentState, error)
    UpdateState(state AgentState) error
    UpdateConfig(config map[string]interface{}) error
    AddCommandResult(result CommandResult) error
}