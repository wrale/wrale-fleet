module github.com/wrale/wrale-fleet/fleet/brain

go 1.21

require (
	github.com/stretchr/testify v1.10.0
	github.com/wrale/wrale-fleet v0.0.0-00010101000000-000000000000
	github.com/wrale/wrale-fleet/metal/hw v0.0.0-00010101000000-000000000000
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/wrale/wrale-fleet => ../../
	github.com/wrale/wrale-fleet/fleet => ../
	github.com/wrale/wrale-fleet/fleet/brain/engine => ./engine
	github.com/wrale/wrale-fleet/metal => ../../metal
	github.com/wrale/wrale-fleet/metal/hw => ../../metal/hw
)