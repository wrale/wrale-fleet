module github.com/wrale/wrale-fleet/user

go 1.22

toolchain go1.23.3

require (
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.3
	github.com/wrale/wrale-fleet/fleet v0.0.0
)

replace (
	github.com/wrale/wrale-fleet/fleet => ../fleet
	github.com/wrale/wrale-fleet/metal => ../metal
	github.com/wrale/wrale-fleet/shared => ../shared
)
