package offline

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// Maximum values for validation
	maxBufferSize   = 1024 * 1024 * 1024 * 10 // 10GB
	maxSyncInterval = 7 * 24 * time.Hour      // 1 week
	maxOperations   = 100                     // Maximum number of supported operations

	// Minimum values for validation
	minBufferSize   = 1024 * 1024     // 1MB
	minSyncInterval = 5 * time.Minute // 5 minutes
)

var (
	// timeRangeRegex validates time range format (HH:MM-HH:MM)
	timeRangeRegex = regexp.MustCompile(`^([0-1][0-9]|2[0-3]):[0-5][0-9]-([0-1][0-9]|2[0-3]):[0-5][0-9]$`)
)

// validateBufferSize checks if the buffer size is within acceptable limits
func validateBufferSize(size int64) error {
	if size < minBufferSize {
		return newError(ErrInvalidConfig,
			fmt.Sprintf("buffer size must be at least %d bytes", minBufferSize))
	}
	if size > maxBufferSize {
		return newError(ErrInvalidConfig,
			fmt.Sprintf("buffer size cannot exceed %d bytes", maxBufferSize))
	}
	return nil
}

// validateSyncInterval checks if the sync interval is within acceptable limits
func validateSyncInterval(interval time.Duration) error {
	if interval < minSyncInterval {
		return newError(ErrInvalidConfig,
			fmt.Sprintf("sync interval must be at least %v", minSyncInterval))
	}
	if interval > maxSyncInterval {
		return newError(ErrInvalidConfig,
			fmt.Sprintf("sync interval cannot exceed %v", maxSyncInterval))
	}
	return nil
}

// validateOperations checks if the operations are valid and within limits
func validateOperations(ops []Operation) error {
	if len(ops) == 0 {
		return newError(ErrInvalidConfig, "at least one operation must be supported")
	}
	if len(ops) > maxOperations {
		return newError(ErrInvalidConfig,
			fmt.Sprintf("cannot exceed %d operations", maxOperations))
	}

	seen := make(map[Operation]bool)
	for _, op := range ops {
		// Check for duplicates
		if seen[op] {
			return newError(ErrInvalidConfig,
				fmt.Sprintf("duplicate operation: %s", op))
		}
		seen[op] = true

		// Validate operation type
		switch op {
		case OpStatusUpdate, OpMetricCollection, OpLogCollection,
			OpConfigValidation, OpHealthCheck:
			// Valid operations
		default:
			return newError(ErrInvalidOperation,
				fmt.Sprintf("unsupported operation: %s", op))
		}
	}

	return nil
}

// validateSyncSchedule checks if the sync schedule is valid
func validateSyncSchedule(schedule map[string]string) error {
	if len(schedule) == 0 {
		return nil // Empty schedule is valid - falls back to interval-based sync
	}

	validDays := map[string]bool{
		"monday": true, "tuesday": true, "wednesday": true,
		"thursday": true, "friday": true, "saturday": true, "sunday": true,
	}

	for day, timeRange := range schedule {
		// Validate day
		day = strings.ToLower(day)
		if !validDays[day] {
			return newError(ErrInvalidConfig,
				fmt.Sprintf("invalid day of week: %s", day))
		}

		// Validate time range format
		if !timeRangeRegex.MatchString(timeRange) {
			return newError(ErrInvalidConfig,
				fmt.Sprintf("invalid time range format: %s", timeRange))
		}

		// Parse and validate time values
		times := strings.Split(timeRange, "-")
		start, err := parseTimeString(times[0])
		if err != nil {
			return newError(ErrInvalidConfig,
				fmt.Sprintf("invalid start time: %s", times[0]))
		}

		end, err := parseTimeString(times[1])
		if err != nil {
			return newError(ErrInvalidConfig,
				fmt.Sprintf("invalid end time: %s", times[1]))
		}

		if end.Before(start) {
			return newError(ErrInvalidConfig,
				fmt.Sprintf("end time cannot be before start time: %s", timeRange))
		}
	}

	return nil
}

// validateBufferStats checks if buffer stats are consistent
func validateBufferStats(stats *BufferStats) error {
	if stats == nil {
		return nil
	}

	if stats.TotalSize < 0 {
		return newError(ErrInvalidConfig, "total buffer size cannot be negative")
	}
	if stats.UsedSize < 0 {
		return newError(ErrInvalidConfig, "used buffer size cannot be negative")
	}
	if stats.AvailableSize < 0 {
		return newError(ErrInvalidConfig, "available buffer size cannot be negative")
	}
	if stats.ItemCount < 0 {
		return newError(ErrInvalidConfig, "item count cannot be negative")
	}

	if stats.UsedSize > stats.TotalSize {
		return newError(ErrInvalidConfig, "used size cannot exceed total size")
	}
	if stats.AvailableSize > stats.TotalSize {
		return newError(ErrInvalidConfig, "available size cannot exceed total size")
	}
	if stats.UsedSize+stats.AvailableSize != stats.TotalSize {
		return newError(ErrInvalidConfig, "used size + available size must equal total size")
	}

	return nil
}

// parseTimeString parses a time string in HH:MM format
func parseTimeString(timeStr string) (time.Time, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid time format")
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return time.Time{}, err
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil {
		return time.Time{}, err
	}

	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC), nil
}
