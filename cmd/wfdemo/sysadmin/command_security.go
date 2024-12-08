package sysadmin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// commandExecutor provides secure command execution for demo scenarios
type commandExecutor struct {
	wfcentralPath string
	allowedCmds   map[string]*regexp.Regexp
}

// newCommandExecutor creates a new command executor with security validation
func newCommandExecutor(wfcentralPath string) (*commandExecutor, error) {
	// Validate wfcentral path
	absPath, err := filepath.Abs(wfcentralPath)
	if err != nil {
		return nil, fmt.Errorf("invalid wfcentral path: %w", err)
	}

	// Verify file exists and is executable
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("wfcentral not found: %w", err)
	}

	if info.Mode()&0111 == 0 {
		return nil, fmt.Errorf("wfcentral is not executable")
	}

	// Define allowed command patterns
	allowedCmds := map[string]*regexp.Regexp{
		"device register": regexp.MustCompile(`^device register --name [a-zA-Z0-9\-]+$`),
		"device get":      regexp.MustCompile(`^device get [a-zA-Z0-9\-]+$`),
		"device status":   regexp.MustCompile(`^device status [a-zA-Z0-9\-]+$`),
		"device health":   regexp.MustCompile(`^device health [a-zA-Z0-9\-]+$`),
		"device alerts":   regexp.MustCompile(`^device alerts [a-zA-Z0-9\-]+$`),
		"device config":   regexp.MustCompile(`^device config (get|set|history) [a-zA-Z0-9\-]+(?: --file [a-zA-Z0-9\-\.]+)?$`),
	}

	return &commandExecutor{
		wfcentralPath: absPath,
		allowedCmds:   allowedCmds,
	}, nil
}

// executeCommand safely executes a whitelisted command
func (c *commandExecutor) executeCommand(ctx context.Context, args []string) error {
	// Validate command against whitelist
	cmdStr := strings.Join(args, " ")
	var allowed bool
	for _, pattern := range c.allowedCmds {
		if pattern.MatchString(cmdStr) {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("command not allowed: %s", cmdStr)
	}

	// Execute command with proper context and output handling
	cmd := exec.CommandContext(ctx, c.wfcentralPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
