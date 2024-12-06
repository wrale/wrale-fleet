module github.com/wrale/wrale-fleet/metal/core

go 1.21

require (
    github.com/wrale/wrale-fleet v0.0.0
    github.com/wrale/wrale-fleet/metal/hw v0.0.0
    github.com/wrale/wrale-fleet/shared v0.0.0
)

replace (
    github.com/wrale/wrale-fleet => ../..
    github.com/wrale/wrale-fleet/metal/hw => ../hw
    github.com/wrale/wrale-fleet/shared => ../../shared
)
