// Package memory provides an in-memory implementation of the device.Store interface.
//
// This package is primarily used for testing, development, and demonstration
// purposes. It maintains device data in memory using a thread-safe map structure.
// The implementation provides full CRUD operations and filtering capabilities
// while ensuring data consistency through proper locking mechanisms.
//
// Note that this implementation is not suitable for production use as it does
// not persist data across process restarts and cannot scale beyond a single
// process.
//
// Example usage:
//
//	store := memory.New()
//	service := device.NewService(store)
package memory
