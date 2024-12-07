package power

// State represents the power state of a device
type State string

const (
	// PowerOn represents the powered on state
	PowerOn State = "on"
	// PowerOff represents the powered off state
	PowerOff State = "off"
	// PowerStandby represents the standby state
	PowerStandby State = "standby"
)
