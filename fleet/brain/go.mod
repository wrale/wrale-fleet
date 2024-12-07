module github.com/wrale/wrale-fleet/fleet/brain

go 1.21

require (
	github.com/wrale/wrale-fleet/fleet/brain/coordinator v0.0.0-00010101000000-000000000000
	github.com/wrale/wrale-fleet/fleet/brain/types v0.0.0-00010101000000-000000000000
)

replace (
	github.com/wrale/wrale-fleet => ../../
	github.com/wrale/wrale-fleet/fleet => ../
	github.com/wrale/wrale-fleet/fleet/brain/coordinator => ./coordinator
	github.com/wrale/wrale-fleet/fleet/brain/types => ./types
	github.com/wrale/wrale-fleet/metal => ../../metal
	github.com/wrale/wrale-fleet/metal/hw => ../../metal/hw
)
