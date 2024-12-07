module github.com/wrale/wrale-fleet/metal/hw

go 1.21

require (
	github.com/wrale/wrale-fleet-metal-hw v0.0.0-20241206205933-2023e062363a
	periph.io/x/conn/v3 v3.7.0
	periph.io/x/host/v3 v3.8.2
)

replace (
	github.com/wrale/wrale-fleet => ../../
	github.com/wrale/wrale-fleet/metal => ../
)
