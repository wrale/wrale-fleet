package metal

import "context"

// Monitor defines common monitoring capabilities
type Monitor interface {
    // GetState returns current state
    GetState() interface{}
    
    // Close releases resources
    Close() error
}

// StateMonitor extends Monitor with type-safe state access
type StateMonitor[T any] interface {
    Monitor
    GetTypedState() T
}

// EventMonitor provides event streaming capabilities
type EventMonitor interface {
    Monitor
    WatchEvents(ctx context.Context) (<-chan interface{}, error)
}

// TypedEventMonitor extends EventMonitor with type-safe events
type TypedEventMonitor[T any] interface {
    EventMonitor
    WatchTypedEvents(ctx context.Context) (<-chan T, error)
}
