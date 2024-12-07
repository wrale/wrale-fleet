package diag

// Results represents diagnostic test results
type Results struct {
	HardwareCheck         bool    `json:"hardware_check"`
	TemperatureWithinRange bool   `json:"temperature_within_range"`
	PowerStable           bool    `json:"power_stable"`
	AdditionalInfo        map[string]interface{} `json:"additional_info,omitempty"`
}
