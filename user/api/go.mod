module github.com/wrale/wrale-fleet/user/api

go 1.22

toolchain go1.23.3

require (
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.1
)

require (
	github.com/wrale/wrale-fleet/fleet v0.0.0
	google.golang.org/grpc v1.68.1
)

require (
	golang.org/x/net v0.29.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
	golang.org/x/text v0.18.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240903143218-8af14fe29dc1 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace (
	github.com/wrale/wrale-fleet => ../..
	github.com/wrale/wrale-fleet/fleet => ../../fleet
	github.com/wrale/wrale-fleet/metal => ../../metal
)
