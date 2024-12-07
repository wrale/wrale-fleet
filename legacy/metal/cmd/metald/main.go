package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wrale/wrale-fleet/metal/internal/hw"
	"github.com/wrale/wrale-fleet/metal/internal/server"
)

func main() {
	// Parse command line flags
	deviceID := flag.String("device-id", "", "Unique device identifier")
	httpAddr := flag.String("http-addr", ":8080", "HTTP API address")
	flag.Parse()

	if *deviceID == "" {
		log.Fatal("device-id is required")
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize hardware monitors
	monitorCfg := hw.MonitorConfig{
		DeviceID: *deviceID,
	}

	thermalMonitor, err := hw.NewThermalMonitor(monitorCfg)
	if err != nil {
		log.Fatalf("Failed to create thermal monitor: %v", err)
	}

	securityMonitor, err := hw.NewSecurityMonitor(monitorCfg)
	if err != nil {
		log.Fatalf("Failed to create security monitor: %v", err)
	}

	// Create policy managers
	thermalMgr, err := server.NewThermalManager(*deviceID, thermalMonitor)
	if err != nil {
		log.Fatalf("Failed to create thermal manager: %v", err)
	}

	securityMgr, err := server.NewSecurityManager(*deviceID, securityMonitor)
	if err != nil {
		log.Fatalf("Failed to create security manager: %v", err)
	}

	// Create server with managers
	srv, err := server.New(server.Config{
		DeviceID:    *deviceID,
		HTTPAddr:    *httpAddr,
		ThermalMgr:  thermalMgr,
		SecurityMgr: securityMgr,
	})
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Handle shutdown gracefully
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Run server
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}