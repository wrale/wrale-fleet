module github.com/wrale/wrale-fleet/user/api

go 1.23.3

require (
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.3
	github.com/wrale/wrale-fleet/fleet/brain v0.0.0-00010101000000-000000000000
)

require golang.org/x/net v0.17.0 // indirect

replace github.com/wrale/wrale-fleet/fleet/brain => ../../fleet/brain