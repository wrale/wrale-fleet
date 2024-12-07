module github.com/wrale/wrale-fleet/metal/core

go 1.21

require (
	github.com/wrale/wrale-fleet/metal/hw v0.0.0
	github.com/wrale/wrale-fleet/shared v0.0.0
)

require (
	github.com/wrale/wrale-fleet-metal-hw v0.0.0-20241206205933-2023e062363a // indirect
	periph.io/x/conn/v3 v3.7.0 // indirect
	periph.io/x/host/v3 v3.8.2 // indirect
)

replace (
	github.com/wrale/wrale-fleet => ../..
	github.com/wrale/wrale-fleet/metal/hw => ../hw
	github.com/wrale/wrale-fleet/shared => ../../shared
)
