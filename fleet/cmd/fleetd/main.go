package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wrale/wrale-fleet/fleet/brain/service"
	"github.com/wrale/wrale-fleet/fleet/brain/coordinator"
)

var (
	Version   string
	BuildTime string
	GitCommit string
)

func main() {
	// Parse command line flags
	metalAddr := flag.String("metal-addr", "localhost:50051", "Metal service address")
	flag.Parse()

	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Printf("Starting wrale-fleet service (Version=%s, BuildTime=%s, GitCommit=%s)", 
		Version, BuildTime, GitCommit)

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, initiating shutdown", sig)
		cancel()
	}()

	// Initialize metal client
	metalClient := coordinator.NewMetalClient(*metalAddr)

	// Create and start brain service
	brainSvc := service.NewService(metalClient)

	// TODO: Add service endpoints (gRPC/HTTP) initialization here

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down wrale-fleet service...")
}
