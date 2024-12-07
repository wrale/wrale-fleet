package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wrale/wrale-fleet/metal/hardware"
	"github.com/wrale/wrale-fleet/metal/internal/server"
	"github.com/wrale/wrale-fleet/metal/secure"
	"github.com/wrale/wrale-fleet/metal/thermal"
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
	thermalMonitor, err := hardware.NewThermalMonitor(*deviceID)
	if err != nil {
		log.Fatalf("Failed to create thermal monitor: %v", err)
	}

	securityMonitor, err := hardware.NewSecureMonitor(*deviceID)
	if err != nil {
		log.Fatalf("Failed to create security monitor: %v", err)
	}

	// Create policy managers
	thermalMgr := thermal.NewPolicyManager(*deviceID, thermalMonitor, thermal.DefaultPolicy())
	securityMgr := secure.NewPolicyManager(*deviceID, securityMonitor, secure.DefaultPolicy())

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