// Package main implements the wfcentral command, which provides the central control plane
// for managing global fleets of devices in the Wrale Fleet Management Platform.
package main

import (
	"github.com/wrale/wrale-fleet/cmd/wfcentral/internal/root"
)

func main() {
	root.Execute()
}