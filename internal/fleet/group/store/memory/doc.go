// Package memory provides an in-memory implementation of the group.Store interface.
//
// This package is primarily intended for testing and demonstration purposes.
// It implements all required store operations with proper multi-tenant isolation
// and supports the full range of group management features, including hierarchy
// operations and device management.
//
// The implementation is not intended for production use, as it does not provide
// persistence or distributed operation capabilities required for enterprise
// deployments.
package memory
