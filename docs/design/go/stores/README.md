# Store Pattern Design

## Introduction

The Wrale Fleet Platform employs a carefully designed store pattern that balances domain isolation with implementation flexibility. This pattern enables us to maintain clear separation between business logic and data persistence while supporting multiple storage backends and deployment scenarios. Understanding this pattern is crucial for maintaining architectural integrity as the system grows.

## Core Principles

Our store pattern builds upon several fundamental principles that guide its implementation across all domains:

### Domain Independence

Each domain defines its own store interface that perfectly matches its specific needs. This interface resides within the domain package itself, ensuring that the domain remains the sole authority over its data requirements. For instance, the device domain defines exactly what device storage operations it needs, without any concern for how those operations might be implemented.

### Implementation Isolation

Store implementations live separately from domain logic, typically organized as:

```
internal/fleet/domain/
    ├── domain.go          # Core domain types
    ├── service.go         # Domain logic
    ├── store.go           # Store interface
    └── store/
        ├── factory/       # Store creation
        └── memory/        # In-memory implementation
```

This structure ensures that implementation details cannot leak into domain logic, while maintaining clear organization of different storage backends.

### Factory-Based Creation

We use a factory pattern to provide clean instantiation of store implementations. The factory package serves as a bridge between domains and their store implementations, preventing direct dependencies while maintaining proper encapsulation. This approach makes it easy to add new storage backends without modifying domain code.

## Implementation Guide

### Store Interface Definition

When designing a store interface for a domain, focus on the domain's actual data needs:

```go
// store.go in domain package
package device

type Store interface {
    // Create stores a new device
    Create(ctx context.Context, device *Device) error
    
    // Get retrieves a device by ID
    Get(ctx context.Context, tenantID, deviceID string) (*Device, error)
    
    // List retrieves devices matching options
    List(ctx context.Context, opts ListOptions) ([]*Device, error)
    
    // Update modifies an existing device
    Update(ctx context.Context, device *Device) error
    
    // Delete removes a device
    Delete(ctx context.Context, tenantID, deviceID string) error
}
```

### Factory Implementation

Create a factory package that provides constructors for all available implementations:

```go
// factory/factory.go
package factory

import "github.com/wrale/fleet/internal/fleet/domain"

// NewMemoryStore creates an in-memory implementation
func NewMemoryStore() domain.Store {
    return memory.New()
}

// NewPostgresStore creates a Postgres-backed implementation
func NewPostgresStore(cfg PostgresConfig) (domain.Store, error) {
    return postgres.New(cfg)
}
```

### Memory Implementation

Provide a memory-based implementation for development and testing:

```go
// memory/store.go
package memory

type Store struct {
    mu    sync.RWMutex
    items map[string]map[string]*domain.Item
}

func New() *Store {
    return &Store{
        items: make(map[string]map[string]*domain.Item),
    }
}

// Implementation methods...
```

## Best Practices

### Tenant Isolation

All store implementations must maintain strict tenant isolation. The first level of map keys in memory stores should always be tenant IDs. For database implementations, tenant ID should be part of primary keys and included in all queries.

### Error Handling

Store implementations should translate storage-specific errors into domain-appropriate errors:

```go
func (s *Store) Get(ctx context.Context, tenantID, itemID string) (*domain.Item, error) {
    if err := db.QueryRow(...).Scan(...); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, domain.ErrItemNotFound
        }
        return nil, fmt.Errorf("querying item: %w", err)
    }
    // ...
}
```

### Context Usage

All store operations should accept and respect context.Context for cancellation and timeout support:

```go
func (s *Store) Create(ctx context.Context, item *domain.Item) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Proceed with creation
    }
}
```

### Validation

Store implementations should validate inputs but should not enforce domain rules:

```go
func (s *Store) Create(ctx context.Context, item *domain.Item) error {
    // Do validate required fields
    if item.ID == "" || item.TenantID == "" {
        return domain.ErrInvalidInput
    }
    
    // Don't validate domain rules - that's the service's job
    // The store just ensures data consistency
}
```

## Testing Approach

Testing presents a fundamental tension in store implementation: balancing clean architecture with practical testing needs. Over time, we have evaluated three potential approaches to making stores easily testable:

1. Convenience Constructor Approach:
Adding constructor functions directly to domain packages might seem simple:

```go
// In domain package, seems convenient but problematic
package domain

func NewMemoryStore() Store {     // Creates import cycle!
    return factory.NewMemoryStore()  
}
```

However, this approach inevitably creates import cycles because:
- Domain package would import factory package
- Factory package already imports domain for interface
- Testing becomes easy but architecture suffers

2. Separate Types Package Approach:
Moving interfaces to a shared package can break cycles:

```go
// Separate from domain, but loses context
package types

type Store interface {
    Create(ctx context.Context, item any) error
    // etc
}
```

This avoids cycles but:
- Separates interfaces from their domain context
- Makes abstraction boundaries less clear
- Loses domain-specific type safety
- Testing becomes complicated

3. Testing Package Approach (Our Choice):
Providing test helpers in a dedicated subpackage:

```go
// Clear purpose, maintains boundaries
package testing

func NewTestStore(t *testing.T) domain.Store {
    return factory.NewMemoryStore()
}
```

This solution:
- Respects domain boundaries
- Avoids import cycles
- Makes testing straightforward
- Preserves type safety
- Allows sharing test fixtures
- Keeps testing utilities together

We have chosen the testing package approach because it provides the best balance of clean architecture and testing practicality. Implementation follows this structure:

```
internal/fleet/domain/
    ├── domain.go          # Core domain types
    ├── service.go         # Domain logic
    ├── store.go           # Store interface
    ├── store/
    │   ├── factory/       # Store creation
    │   └── memory/        # In-memory implementation
    ├── domain_test.go     # Tests using testing package
    └── testing/           # Testing utilities
        ├── helpers.go     # Test store creation
        └── fixtures.go    # Common test data
```

The testing package provides clear, focused utilities:

```go
package testing

import (
    "context"
    "testing"
    
    "github.com/wrale/fleet/internal/fleet/domain"
    "github.com/wrale/fleet/internal/fleet/domain/store/factory"
)

// NewTestStore creates a configured memory store for testing
func NewTestStore(t *testing.T) domain.Store {
    t.Helper()
    return factory.NewMemoryStore()
}

// NewTestService creates a service with test dependencies
func NewTestService(t *testing.T) *domain.Service {
    t.Helper()
    store := NewTestStore(t)
    return domain.NewService(store, testLogger)
}

// CreateTestData populates the store with standard test fixtures
func CreateTestData(ctx context.Context, t *testing.T, store domain.Store) error {
    t.Helper()
    // Create canonical test data set
    return nil
}
```

This enables straightforward test code:

```go
func TestService(t *testing.T) {
    // Clean test setup using helpers
    store := testing.NewTestStore(t)
    service := testing.NewTestService(t)
    
    // Standard test data population
    ctx := context.Background()
    err := testing.CreateTestData(ctx, t, store)
    require.NoError(t, err)
    
    // Clear test logic without setup complexity
    t.Run("List", func(t *testing.T) {
        results, err := service.List(ctx, domain.ListOptions{})
        require.NoError(t, err)
        assert.NotEmpty(t, results)
    })
}
```

### Interface Compliance

Every store implementation must verify interface compliance through compile-time checks:

```go
var (
    // Ensure Store implements domain.Store
    _ domain.Store = (*Store)(nil)
)
```

### Common Test Suites

Store implementations should verify core functionality through standard test suites:

```go
func RunStoreTests(t *testing.T, newStore func() domain.Store) {
    t.Run("Create", testStoreCreate(newStore))
    t.Run("Get", testStoreGet(newStore))
    t.Run("List", testStoreList(newStore))
    t.Run("Update", testStoreUpdate(newStore))
    t.Run("Delete", testStoreDelete(newStore))
}
```

### Isolation Testing

All implementations must verify tenant isolation explicitly:

```go
func TestStoreIsolation(t *testing.T) {
    store := memory.New()
    
    // Create item in tenant1
    err := store.Create(ctx, tenant1, item1)
    require.NoError(t, err)
    
    // Verify tenant2 cannot access it
    _, err = store.Get(ctx, tenant2, item1.ID)
    assert.Equal(t, domain.ErrItemNotFound, err)
}
```

## Stage-Aware Design

Our store interfaces should support staged capability growth:

1. Start with basic CRUD operations in Stage 1
2. Add querying and filtering in Stage 2
3. Introduce cross-region operations in Stage 3
4. Support advanced security features in Stage 4
5. Enable mesh operations in Stage 5
6. Add enterprise features in Stage 6

## Success Criteria

A store implementation is considered successful when it:

1. Maintains complete tenant isolation
2. Properly implements the domain interface
3. Passes all common test suites
4. Handles errors appropriately
5. Respects context cancellation
6. Provides consistent performance
7. Supports staged evolution
8. Maintains proper separation from domain logic

## Migration Considerations

When adding new store implementations:

1. Create the implementation package
2. Add factory method
3. Implement the interface
4. Add appropriate tests
5. Provide migration tooling if needed
6. Update documentation
7. Add monitoring support
8. Consider performance implications

## Future Considerations

As we develop additional store implementations, we should consider:

1. Advanced caching strategies
2. Improved query optimization
3. Better bulk operation support
4. Enhanced monitoring capabilities
5. Automated failover support
6. Extended backup features
7. Advanced security controls
8. Improved performance metrics

The store pattern provides a solid foundation for growth while maintaining clean architecture and clear domain boundaries. By following these guidelines, we ensure consistent and maintainable persistence implementations across the platform.
