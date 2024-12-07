module github.com/wrale/wrale-fleet/metal/hw

go 1.21

require (
	periph.io/x/conn/v3 v3.7.0
	periph.io/x/host/v3 v3.8.2
)

replace (
	github.com/wrale/wrale-fleet => ../../
	github.com/wrale/wrale-fleet/metal => ../
)
