// Package main implements the wfdevice command, which provides local device management
// capabilities for the Wrale Fleet Management Platform.
package main

import (
	"github.com/wrale/wrale-fleet/cmd/wfdevice/internal/root"
)

func main() {
	root.Execute()
}
