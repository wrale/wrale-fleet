package agent

import (
    "context"
    "fmt"
    "sync"
    "time"

    "github.com/wrale/wrale-fleet/fleet/brain/types"
)

// Agent implements the edge agent functionality
type Agent struct {
    config      AgentConfig
    state       AgentState
    metalClient MetalClient
    brainClient BrainClient
    stateStore  StateStore
    
    commandChan chan Command
    resultChan  chan CommandResult
    stopChan    chan struct{}
    mu          sync.RWMutex
}

// NewAgent creates a new edge agent instance
func NewAgent(
    config AgentConfig,
    metalClient MetalClient,
    brainClient BrainClient,
    stateStore StateStore,
) *Agent {
    return &Agent{
        config:      config,
        metalClient: metalClient,
        brainClient: brainClient,
        stateStore:  stateStore,
        commandChan: make(chan Command, 100),
        resultChan:  make(chan CommandResult, 100),
        stopChan:    make(chan struct{}),
    }
}

// Start begins the agent's operation
func (a *Agent) Start(ctx context.Context) error {
    // Load initial state
    storedState, err := a.stateStore.GetState()
    if err != nil {
        return fmt.Errorf("failed to load initial state: %w", err)
    }
    a.state = storedState

    // Start operation loops
    go a.stateLoop(ctx)
    go a.commandLoop(ctx)
    go a.healthLoop(ctx)

    return nil
}

// Stop gracefully stops the agent
func (a *Agent) Stop() {
    close(a.stopChan)
}

// stateLoop periodically updates and syncs device state
func (a *Agent) stateLoop(ctx context.Context) {
    ticker := time.NewTicker(a.config.UpdateInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-a.stopChan:
            return
        case <-ticker.C:
            if err := a.updateAndSyncState(ctx); err != nil {
                // Log error but continue operation
                a.handleError("state_sync", err)
            }
        }
    }
}

// commandLoop processes incoming commands
func (a *Agent) commandLoop(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case <-a.stopChan:
            return
        default:
            // Get commands from brain if in normal mode
            if a.getMode() == ModeNormal {
                commands, err := a.brainClient.GetCommands()
                if err != nil {
                    a.handleError("get_commands", err)
                    time.Sleep(time.Second) // Backoff on error
                    continue
                }

                // Process commands
                for _, cmd := range commands {
                    select {
                    case a.commandChan <- cmd:
                    default:
                        // Command buffer full, log warning
                    }
                }
            }

            // Process command from channel
            select {
            case cmd := <-a.commandChan:
                result := a.executeCommand(ctx, cmd)
                
                // Report result if in normal mode
                if a.getMode() == ModeNormal {
                    if err := a.brainClient.ReportCommandResult(result); err != nil {
                        a.handleError("report_result", err)
                    }
                }

                // Store result locally
                if err := a.stateStore.AddCommandResult(result); err != nil {
                    a.handleError("store_result", err)
                }

            case <-time.After(time.Second):
                // No commands to process
                continue
            }
        }
    }
}

// healthLoop monitors device health
func (a *Agent) healthLoop(ctx context.Context) {
    ticker := time.NewTicker(time.Second * 30)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-a.stopChan:
            return
        case <-ticker.C:
            healthy, err := a.metalClient.GetHealthStatus()
            if err != nil {
                a.handleError("health_check", err)
                continue
            }

            diagnostics, err := a.metalClient.RunDiagnostics()
            if err != nil {
                a.handleError("diagnostics", err)
                continue
            }

            a.updateHealth(healthy, diagnostics)
        }
    }
}

// updateAndSyncState updates local state and syncs with brain
func (a *Agent) updateAndSyncState(ctx context.Context) error {
    // Get latest metrics
    metrics, err := a.metalClient.GetMetrics()
    if err != nil {
        return fmt.Errorf("failed to get metrics: %w", err)
    }

    // Update local state
    a.mu.Lock()
    a.state.DeviceState.Metrics = metrics
    a.state.LastSync = time.Now()
    a.mu.Unlock()

    // Sync with brain if in normal mode
    if a.getMode() == ModeNormal {
        if err := a.brainClient.SyncState(a.state.DeviceState); err != nil {
            return fmt.Errorf("failed to sync state: %w", err)
        }
    }

    // Store updated state
    return a.stateStore.UpdateState(a.state)
}

// executeCommand executes a command and returns the result
func (a *Agent) executeCommand(ctx context.Context, cmd Command) CommandResult {
    result := CommandResult{
        CommandID: cmd.ID,
        CompletedAt: time.Now(),
    }

    switch cmd.Type {
    case CmdUpdateState:
        // Handle state update
        if err := a.updateAndSyncState(ctx); err != nil {
            result.Error = err
        } else {
            result.Success = true
        }

    case CmdExecuteTask:
        // Execute operation on metal layer
        if task, ok := cmd.Payload.(string); ok {
            if err := a.metalClient.ExecuteOperation(task); err != nil {
                result.Error = err
            } else {
                result.Success = true
            }
        } else {
            result.Error = fmt.Errorf("invalid task payload")
        }

    case CmdUpdateConfig:
        // Update configuration
        if config, ok := cmd.Payload.(map[string]interface{}); ok {
            if err := a.stateStore.UpdateConfig(config); err != nil {
                result.Error = err
            } else {
                result.Success = true
            }
        } else {
            result.Error = fmt.Errorf("invalid config payload")
        }

    case CmdEnterSafeMode:
        a.setMode(ModeSafe)
        result.Success = true

    case CmdExitSafeMode:
        a.setMode(ModeNormal)
        result.Success = true

    default:
        result.Error = fmt.Errorf("unknown command type: %s", cmd.Type)
    }

    return result
}

// updateHealth updates the agent's health status
func (a *Agent) updateHealth(healthy bool, diagnostics map[string]interface{}) {
    a.mu.Lock()
    defer a.mu.Unlock()

    a.state.IsHealthy = healthy
    
    // Update operation mode based on health
    if !healthy && a.state.Mode == ModeNormal {
        a.state.Mode = ModeSafe
    }

    // Report health to brain if in normal mode
    if a.state.Mode == ModeNormal {
        if err := a.brainClient.ReportHealth(healthy, diagnostics); err != nil {
            a.handleError("health_report", err)
        }
    }
}

// getMode returns the current operation mode
func (a *Agent) getMode() OperationMode {
    a.mu.RLock()
    defer a.mu.RUnlock()
    return a.state.Mode
}

// setMode updates the operation mode
func (a *Agent) setMode(mode OperationMode) {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.state.Mode = mode
}

// handleError processes operational errors
func (a *Agent) handleError(context string, err error) {
    a.mu.Lock()
    defer a.mu.Unlock()

    a.state.LastError = fmt.Errorf("%s: %w", context, err)
    
    // Update mode if communication with brain is lost
    if context == "state_sync" && a.state.Mode == ModeNormal {
        a.state.Mode = ModeAutonomous
    }
}