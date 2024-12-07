package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wrale/wrale-fleet/user/api/server"
	"github.com/wrale/wrale-fleet/user/api/service"
)

func main() {
	// Parse command line flags
	httpAddr := flag.String("http-addr", ":8080", "HTTP API address")
	fleetEndpoint := flag.String("fleet-endpoint", "localhost:9090", "Fleet server endpoint")
	flag.Parse()

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize services
	deviceSvc := service.NewDeviceService(*fleetEndpoint)
	fleetSvc := service.NewFleetService(*fleetEndpoint)
	wsSvc := service.NewWebSocketService(*fleetEndpoint)
	authSvc := service.NewAuthService()

	// Create server with services
	srv := server.NewServer(
		deviceSvc,
		fleetSvc,
		wsSvc,
		authSvc,
	)

	// Handle shutdown gracefully
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Received signal %v, shutting down...", sig)
		cancel()
	}()

	// Start server
	log.Printf("Starting API server on %s", *httpAddr)
	if err := srv.Start(*httpAddr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}