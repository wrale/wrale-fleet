package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:   "wfdemo",
	Short: "Wrale Fleet Management Platform Demonstration Tool",
	Long: `wfdemo provides guided demonstrations of the Wrale Fleet Management Platform
capabilities through the lens of different personas and use cases. Each demo
showcases specific platform features and workflows.`,
}

func init() {
	// Add demo command groups
	rootCmd.AddCommand(sysadminCmd())
}

// safeSync attempts to sync the logger, handling common sync issues gracefully.
// This is a simplified version of the sync handler used in fleetd, appropriate
// for the demo tool's needs.
func safeSync(logger *zap.Logger) error {
	err := logger.Sync()
	if err == nil {
		return nil
	}

	// Handle common stdout/stderr sync issues that can be safely ignored
	errStr := err.Error()
	if strings.Contains(errStr, "invalid argument") ||
		strings.Contains(errStr, "inappropriate ioctl for device") {
		return nil
	}

	// Return unexpected sync errors for handling
	return err
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := safeSync(logger); err != nil {
			fmt.Fprintf(os.Stderr, "failed to sync logger: %v\n", err)
		}
	}()

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		logger.Info("received signal", zap.String("signal", sig.String()))
		cancel()
	}()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		logger.Error("command execution failed", zap.Error(err))
		os.Exit(1)
	}
}
