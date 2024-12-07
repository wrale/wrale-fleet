module github.com/wrale/wrale-fleet/metal

go 1.21

require (
	github.com/wrale/wrale-fleet/shared v0.0.0
	periph.io/x/conn/v3 v3.7.0
	periph.io/x/host/v3 v3.8.2
)

replace github.com/wrale/wrale-fleet => ../

replace github.com/wrale/wrale-fleet/metal => ../metal

replace github.com/wrale/wrale-fleet/shared => ../shared

replace github.com/wrale/wrale-fleet/sync => ../sync
