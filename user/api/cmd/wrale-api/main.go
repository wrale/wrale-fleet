package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wrale/wrale-fleet/user/api/server"
)

var (
	// Build information - set via ldflags
	Version    string
	BuildTime  string
	GitCommit  string

	// Command line flags
	httpAddr = flag.String("http", ":8080", "HTTP API listen address")
)

func main() {
	flag.Parse()

	// Log build info
	log.Printf("Starting wrale-api version %s (%s) built at %s", Version, GitCommit, BuildTime)

	// Initialize server
	cfg := server.Config{
		HTTPAddr: *httpAddr,
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		log.Printf("Received signal %v, initiating shutdown", sig)
		cancel()
	}()

	// Run server
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}