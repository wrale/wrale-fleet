package service

import (
    "context"
    "sync"
    "time"

    "github.com/wrale/wrale-fleet/fleet/brain/service"
    "github.com/wrale/wrale-fleet/fleet/brain/types"
    apitypes "github.com/wrale/wrale-fleet/user/api/types"
)

// WebSocketService implements real-time updates
type WebSocketService struct {
    brainSvc *service.Service
    
    // Track subscriptions
    subscribers map[chan<- *apitypes.WSMessage]map[types.DeviceID]bool
    mu          sync.RWMutex

    // Control channels
    updates     chan *apitypes.WSMessage
    done        chan struct{}
}

// NewWebSocketService creates a new websocket service
func NewWebSocketService(brainSvc *service.Service) *WebSocketService {
    ws := &WebSocketService{
        brainSvc:    brainSvc,
        subscribers: make(map[chan<- *apitypes.WSMessage]map[types.DeviceID]bool),
        updates:     make(chan *apitypes.WSMessage, 1000),
        done:        make(chan struct{}),
    }

    // Start update handling
    go ws.handleUpdates()

    // Start state polling (temporary for v1.0)
    go ws.pollDeviceStates()

    return ws
}

// Subscribe adds a new subscriber
func (ws *WebSocketService) Subscribe(deviceIDs []types.DeviceID, updates chan<- *apitypes.WSMessage) error {
    ws.mu.Lock()
    defer ws.mu.Unlock()

    // Initialize device map
    deviceMap := make(map[types.DeviceID]bool)
    for _, id := range deviceIDs {
        deviceMap[id] = true
    }

    ws.subscribers[updates] = deviceMap
    return nil
}

// Unsubscribe removes a subscriber
func (ws *WebSocketService) Unsubscribe(updates chan<- *apitypes.WSMessage) error {
    ws.mu.Lock()
    defer ws.mu.Unlock()

    delete(ws.subscribers, updates)
    return nil
}

// Broadcast sends a message to all subscribers
func (ws *WebSocketService) Broadcast(msg *apitypes.WSMessage) error {
    ws.updates <- msg
    return nil
}

// Stop stops the service
func (ws *WebSocketService) Stop() {
    close(ws.done)
}

// handleUpdates processes and distributes updates
func (ws *WebSocketService) handleUpdates() {
    for {
        select {
        case <-ws.done:
            return
        case msg := <-ws.updates:
            ws.distributeMessage(msg)
        }
    }
}

// distributeMessage sends message to interested subscribers
func (ws *WebSocketService) distributeMessage(msg *apitypes.WSMessage) {
    ws.mu.RLock()
    defer ws.mu.RUnlock()

    var deviceID types.DeviceID

    // Extract device ID based on message type
    switch v := msg.Payload.(type) {
    case *apitypes.WSStateUpdate:
        deviceID = v.DeviceID
    case *apitypes.WSMetricsUpdate:
        deviceID = v.DeviceID
    case *apitypes.WSAlertMessage:
        deviceID = v.DeviceID
    }

    // Distribute to interested subscribers
    for sub, devices := range ws.subscribers {
        // Send if subscriber wants all devices or this specific device
        if len(devices) == 0 || devices[deviceID] {
            select {
            case sub <- msg:
            default:
                // Skip if subscriber's channel is full
            }
        }
    }
}

// pollDeviceStates polls for device state changes (temporary for v1.0)
func (ws *WebSocketService) pollDeviceStates() {
    ticker := time.NewTicker(time.Second * 5)
    defer ticker.Stop()

    lastStates := make(map[types.DeviceID]types.DeviceState)

    for {
        select {
        case <-ws.done:
            return
        case <-ticker.C:
            ctx := context.Background()
            
            // Get current states
            devices, err := ws.brainSvc.ListDevices(ctx)
            if err != nil {
                continue
            }

            // Check for changes
            for _, device := range devices {
                lastState, exists := lastStates[device.ID]
                
                // Send update if state changed or new device
                if !exists || stateChanged(lastState, *device) {
                    ws.updates <- &apitypes.WSMessage{
                        Type: "state_update",
                        Payload: &apitypes.WSStateUpdate{
                            DeviceID: device.ID,
                            State:    *device,
                            Time:     time.Now(),
                        },
                    }

                    // Send metrics update if changed
                    if !exists || metricsChanged(lastState.Metrics, device.Metrics) {
                        ws.updates <- &apitypes.WSMessage{
                            Type: "metrics_update",
                            Payload: &apitypes.WSMetricsUpdate{
                                DeviceID: device.ID,
                                Metrics:  device.Metrics,
                                Time:     time.Now(),
                            },
                        }
                    }

                    lastStates[device.ID] = *device
                }
            }

            // Check for removed devices
            for id := range lastStates {
                found := false
                for _, device := range devices {
                    if device.ID == id {
                        found = true
                        break
                    }
                }
                if !found {
                    delete(lastStates, id)
                    ws.updates <- &apitypes.WSMessage{
                        Type: "device_removed",
                        Payload: map[string]interface{}{
                            "device_id": id,
                            "time":      time.Now(),
                        },
                    }
                }
            }
        }
    }
}

// stateChanged checks if device state has changed significantly
func stateChanged(old, new types.DeviceState) bool {
    if old.Status != new.Status {
        return true
    }
    if old.Location != new.Location {
        return true
    }
    return metricsChanged(old.Metrics, new.Metrics)
}

// metricsChanged checks if device metrics have changed significantly
func metricsChanged(old, new types.DeviceMetrics) bool {
    const threshold = 1.0 // Minimum change to trigger update

    if abs(old.Temperature - new.Temperature) > threshold {
        return true
    }
    if abs(old.CPULoad - new.CPULoad) > threshold {
        return true
    }
    if abs(old.MemoryUsage - new.MemoryUsage) > threshold {
        return true
    }
    if abs(old.PowerUsage - new.PowerUsage) > threshold {
        return true
    }
    return false
}

// abs returns absolute value of float64
func abs(x float64) float64 {
    if x < 0 {
        return -x
    }
    return x
}
