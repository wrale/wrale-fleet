module github.com/wrale/wrale-fleet/sync

go 1.22
toolchain go1.23.3

require (
    github.com/stretchr/testify v1.8.4
    github.com/wrale/wrale-fleet/shared v0.0.0
)

require (
    github.com/davecgh/go-spew v1.1.1 // indirect
    github.com/pmezard/go-difflib v1.0.0 // indirect
    gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/wrale/wrale-fleet/shared => ../shared
