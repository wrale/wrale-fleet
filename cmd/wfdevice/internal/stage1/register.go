package stage1

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wrale/wrale-fleet/cmd/wfdevice/options"
)

func newRegisterCmd(cfg *options.Config) (*cobra.Command, error) {
	var (
		name         string
		controlPlane string
		tags         []string
	)

	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register device with control plane",
		Long: `Register this device with the specified control plane.

Registration includes:
- Establishing device identity
- Configuring control plane connection
- Setting device metadata
- Initializing security credentials
- Verifying connectivity

Tags can be specified as key=value pairs and are used for:
- Device grouping
- Policy application
- Resource organization
- Operation targeting`,
		Example: `  # Register with required fields
  wfdevice register --name device-1 --control-plane central.example.com

  # Register with tags
  wfdevice register --name device-1 --control-plane central.example.com \
    --tags environment=production,location=datacenter-1`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRegister(cmd.Context(), cfg, name, controlPlane, tags)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Device name")
	cmd.Flags().StringVar(&controlPlane, "control-plane", "", "Control plane address")
	cmd.Flags().StringSliceVar(&tags, "tags", []string{}, "Device tags (key=value)")

	// Mark required flags and handle potential errors
	if err := cmd.MarkFlagRequired("name"); err != nil {
		return nil, fmt.Errorf("marking name flag as required: %w", err)
	}
	if err := cmd.MarkFlagRequired("control-plane"); err != nil {
		return nil, fmt.Errorf("marking control-plane flag as required: %w", err)
	}

	return cmd, nil
}

func runRegister(ctx context.Context, cfg *options.Config, name, controlPlane string, tags []string) error {
	// Validate registration parameters
	if err := validateRegisterParams(name, controlPlane); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	// Parse tags into map
	tagMap, err := parseTags(tags)
	if err != nil {
		return fmt.Errorf("parsing tags: %w", err)
	}

	// Initialize registration client
	client, err := options.NewRegistrationClient(controlPlane)
	if err != nil {
		return fmt.Errorf("initializing registration client: %w", err)
	}

	// Perform registration
	if err := client.Register(ctx, name, tagMap); err != nil {
		return fmt.Errorf("registering device: %w", err)
	}

	fmt.Printf("Successfully registered device '%s' with control plane\n", name)
	return nil
}

func validateRegisterParams(name, controlPlane string) error {
	if name == "" {
		return fmt.Errorf("device name is required")
	}
	if controlPlane == "" {
		return fmt.Errorf("control plane address is required")
	}
	return nil
}

func parseTags(tags []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, tag := range tags {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid tag format '%s': must be key=value", tag)
		}
		key, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("tag key cannot be empty")
		}
		if value == "" {
			return nil, fmt.Errorf("tag value cannot be empty")
		}
		result[key] = value
	}
	return result, nil
}
