package service

import (
    "context"
    "fmt"
    "log"
    "sync"
    "time"
    "google.golang.org/grpc"

    "github.com/gorilla/websocket"
    "github.com/wrale/wrale-fleet/fleet/brain/service"
    "github.com/wrale/wrale-fleet/fleet/brain/types"
)

// WebSocketService handles real-time device updates
type WebSocketService struct {
    brainSvc  *service.Service
    conn      *grpc.ClientConn
    clients   map[*websocket.Conn]bool
    broadcast chan interface{}
    mu        sync.RWMutex
}

// NewWebSocketService creates a new websocket service
func NewWebSocketService(fleetEndpoint string) *WebSocketService {
    // Connect to fleet brain service
    conn, err := grpc.Dial(fleetEndpoint, grpc.WithInsecure())
    if err != nil {
        log.Fatalf("Failed to connect to fleet brain: %v", err)
    }

    svc := &WebSocketService{
        brainSvc:  service.NewClient(conn),
        conn:      conn,
        clients:   make(map[*websocket.Conn]bool),
        broadcast: make(chan interface{}, 256),
    }

    // Start event processing
    go svc.processEvents()
    go svc.broadcastLoop()

    return svc
}

// Close releases resources
func (s *WebSocketService) Close() error {
    if s.conn != nil {
        return s.conn.Close()
    }
    return nil
}

// AddClient registers a new websocket client
func (s *WebSocketService) AddClient(client *websocket.Conn) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.clients[client] = true
}

// RemoveClient unregisters a websocket client
func (s *WebSocketService) RemoveClient(client *websocket.Conn) {
    s.mu.Lock()
    defer s.mu.Unlock()
    delete(s.clients, client)
    client.Close()
}

// processEvents subscribes to brain events and broadcasts them
func (s *WebSocketService) processEvents() {
    for {
        ctx := context.Background()
        stream, err := s.brainSvc.SubscribeEvents(ctx)
        if err != nil {
            log.Printf("Error subscribing to events: %v", err)
            time.Sleep(5 * time.Second)
            continue
        }

        for {
            event, err := stream.Recv()
            if err != nil {
                log.Printf("Error receiving event: %v", err)
                break
            }

            s.broadcast <- event
        }
    }
}

// broadcastLoop sends events to all websocket clients
func (s *WebSocketService) broadcastLoop() {
    for {
        msg := <-s.broadcast

        s.mu.RLock()
        for client := range s.clients {
            if err := client.WriteJSON(msg); err != nil {
                log.Printf("Error broadcasting to client: %v", err)
                client.Close()
                delete(s.clients, client)
            }
        }
        s.mu.RUnlock()
    }
}

// GetDeviceUpdates subscribes to updates for a specific device
func (s *WebSocketService) GetDeviceUpdates(deviceID types.DeviceID) (<-chan interface{}, error) {
    updates := make(chan interface{}, 32)

    // Subscribe to device specific events
    ctx := context.Background()
    stream, err := s.brainSvc.SubscribeDeviceEvents(ctx, deviceID)
    if err != nil {
        return nil, fmt.Errorf("failed to subscribe to device events: %w", err)
    }

    // Forward events to channel
    go func() {
        defer close(updates)
        for {
            event, err := stream.Recv()
            if err != nil {
                log.Printf("Error receiving device event: %v", err)
                return
            }
            updates <- event
        }
    }()

    return updates, nil
}