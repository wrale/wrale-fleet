package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/wrale/wrale-fleet/fleet/brain/coordinator"
	"github.com/wrale/wrale-fleet/fleet/brain/service"
)

var (
	Version   string
	BuildTime string
	GitCommit string
)

type server struct {
	brainSvc *service.Service
}

func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "ok",
		"version":   Version,
		"buildTime": BuildTime,
		"gitCommit": GitCommit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

func (s *server) deviceListHandler(w http.ResponseWriter, r *http.Request) {
	devices, err := s.brainSvc.ListDevices(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
}

func main() {
	// Parse command line flags
	metalAddr := flag.String("metal-addr", "http://localhost:8081", "Metal service address")
	listenAddr := flag.String("listen", ":8080", "HTTP server listen address")
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

	// Initialize metal client and brain service
	metalClient := coordinator.NewMetalClient(*metalAddr)
	brainSvc := service.NewService(metalClient)

	// Initialize HTTP server
	srv := &server{brainSvc: brainSvc}
	
	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.healthHandler)
	mux.HandleFunc("/devices", srv.deviceListHandler)

	httpServer := &http.Server{
		Addr:    *listenAddr,
		Handler: mux,
	}

	// Start HTTP server
	go func() {
		log.Printf("Starting HTTP server on %s", *listenAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutting down wrale-fleet service...")

	// Graceful shutdown
	if err := httpServer.Shutdown(context.Background()); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}
