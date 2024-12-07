module github.com/wrale/wrale-fleet/user/api

go 1.21

require (
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.1
)

require github.com/wrale/wrale-fleet/fleet v0.0.0

require golang.org/x/net v0.17.0 // indirect

replace (
	github.com/wrale/wrale-fleet => ../..
	github.com/wrale/wrale-fleet/fleet => ../../fleet
	github.com/wrale/wrale-fleet/metal => ../../metal
)
