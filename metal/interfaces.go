// Package metal provides hardware abstraction and management interfaces for Raspberry Pi devices
package metal

// Option defines a functional option for configuring hardware components
type Option func(interface{}) error

// CompareState compares hardware states ignoring UpdatedAt
func CompareState(a, b interface{}) bool {
	// TODO: Implement state comparison logic
	return true
}