# Go Project Layout - IoT Fleet Management Platform

## Root Structure

The project follows standard Go project layout patterns while adapting them to our needs. At the root level, we organize code into distinct areas of responsibility:

```
wrale-fleet/
├── cmd/                   # Application entry points
├── internal/              # Private application code
├── pkg/                   # Public library code
└── api/                   # API definitions and clients
```

## Application Entry Points

The cmd directory contains our main applications, each with their own main package:

```go
cmd/
├── wfcentral/           # Enterprise control plane
│   └── main.go          # Minimal main function using functional options
└── wfdevice/            # Device agent
    └── main.go          # Minimal main function using functional options

// Example main.go showing staged capability loading
func main() {
    // Parse command-line flags
    port := flag.String("port", "8080", "Server port")
    dataDir := flag.String("data-dir", "/var/lib/wfcentral", "Data directory")
    flag.Parse()

    // Initialize server with staged capabilities
    srv, err := server.New(
        server.WithPort(*port),
        server.WithDataDir(*dataDir),
        server.WithStore(store.NewPostgres()),
    )

    if err := srv.Run(context.Background()); err != nil {
        log.Fatal(err)
    }
}
```

## Internal Application Code 

The internal code maintains strict boundaries between domain logic and server infrastructure:

```go
internal/
├── fleet/               # Core domain packages
│   ├── device/          # Device management domain
│   │   ├── device.go    # Core device type
│   │   ├── service.go   # Device operations
│   │   └── store.go     # Storage interface
│   ├── group/           # Group management domain  
│   │   ├── group.go
│   │   └── service.go
│   └── config/          # Configuration domain
│       ├── config.go    
│       └── service.go
│
├── server/              # Server infrastructure
│   ├── central/         # Control plane servers
│   │   ├── server.go    # Base server implementation
│   │   ├── sync.go      # Multi-region synchronization
│   │   ├── cluster.go   # Cluster coordination
│   │   └── stage2.go    # Stage 2 capability servers
│   └── device/          # Device agent servers
│       ├── server.go    # Base agent server
│       ├── monitor.go   # Health monitoring
│       └── proxy.go     # Device proxy handling
│
└── store/               # Storage implementations
    ├── postgres/        # Postgres implementation
    └── redis/           # Redis implementation

// Example device.go showing domain type
package device

type Device struct {
    ID          string
    Name        string
    Config      json.RawMessage
    Status      DeviceStatus
    LastSeen    time.Time
    
    // Stage 2: Multi-site capabilities
    Region     string         `json:"region,omitempty"`
    ClusterID  string         `json:"cluster_id,omitempty"`
}

// Example server.go showing infrastructure
package central

type Server struct {
    cfg        *Config
    device     *device.Service    // Domain service
    group      *group.Service     // Domain service
    store      store.Store
    httpServer *http.Server
}

func (s *Server) Run(ctx context.Context) error {
    // Wire up domain services
    // Set up infrastructure
    // Start server
}
```

## Package Boundaries and Responsibilities

1. Domain Packages (internal/fleet/*)
   - Contain core business logic and types
   - Define interfaces for persistence
   - Remain infrastructure-agnostic
   - Focus on single domain concept
   - No cross-domain dependencies

2. Server Packages (internal/server/*)
   - Implement network protocols
   - Handle distributed systems concerns
   - Coordinate between domains
   - Manage infrastructure lifecycle
   - Implement staged capabilities

3. Storage Packages (internal/store/*)
   - Implement store interfaces
   - Handle persistence details
   - Manage transactions
   - Implement optimizations

## Key Design Patterns

### Option Pattern for Server Configuration
```go
type Option func(*Server) error

func WithStore(s store.Store) Option {
    return func(srv *Server) error {
        srv.store = s
        return nil
    }
}
```

### Context Usage with Capability Awareness
```go
type ctxKey struct{} 

func FromContext(ctx context.Context) (Tenant, bool) {
    t, ok := ctx.Value(ctxKey{}).(Tenant)
    return t, ok
}
```

### Error Handling with Stage Information
```go
type Error struct {
    Code    string
    Message string
    Stage   int    // Minimum stage requirement
}

func (e *Error) Error() string {
    return fmt.Sprintf("%s: %s (stage %d+)", e.Code, e.Message, e.Stage)
}
```

## Package Organization Rules

1. Package Layout
   - One package per directory
   - Package name matches directory name
   - Files named for primary type
   - Test files next to code
   - Stage-specific code marked clearly

2. Dependencies
   - Domain packages must not depend on each other
   - Server packages can depend on multiple domains
   - Dependencies flow inward
   - Infrastructure depends on domain, not vice versa

3. Staged Evolution
   - Base capabilities in root package
   - Advanced features in stage-specific files
   - Clear stage requirements in errors
   - Graceful capability degradation

4. Testing
   - Tests next to the code they verify
   - Shared test utilities in testing packages
   - Infrastructure tests use interfaces
   - Stage-aware test helpers

## Success Criteria

A package organization is considered successful when:
- Domain boundaries are clear and enforced
- Infrastructure concerns are separated
- Stage evolution is manageable
- Testing is comprehensive
- Dependencies flow correctly
- New features fit naturally
