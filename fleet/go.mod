module github.com/wrale/wrale-fleet/fleet

go 1.21

require github.com/wrale/wrale-fleet/metal/hw v0.0.0

require (
	periph.io/x/conn/v3 v3.7.0 // indirect
	periph.io/x/host/v3 v3.8.2 // indirect
)

replace (
	github.com/wrale/wrale-fleet => ../
	github.com/wrale/wrale-fleet/metal => ../metal
	github.com/wrale/wrale-fleet/metal/hw => ../metal/hw
)
