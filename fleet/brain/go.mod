module github.com/wrale/wrale-fleet/fleet/brain

go 1.21

require (
	github.com/wrale/wrale-fleet/metal/hw v0.0.0-00010101000000-000000000000
	github.com/wrale/wrale-fleet/fleet v0.0.0-00010101000000-000000000000
)

require (
    github.com/wrale/wrale-fleet/metal v0.0.0-00010101000000-000000000000 // indirect
)

replace (
    github.com/wrale/wrale-fleet => ../../
    github.com/wrale/wrale-fleet/fleet => ../
    github.com/wrale/wrale-fleet/metal => ../../metal
    github.com/wrale/wrale-fleet/metal/hw => ../../metal/hw
)