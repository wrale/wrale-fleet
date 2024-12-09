package memory

import (
	"encoding/json"
	"fmt"

	"github.com/wrale/wrale-fleet/internal/fleet/config"
)

// createTestTemplate creates a new template for testing purposes
// with deterministic IDs for easier testing
func createTestTemplate(id, tenantID string) *config.Template {
	schema := json.RawMessage(`{"type": "object"}`)
	template := config.NewTemplate(tenantID, fmt.Sprintf("template-%s", id), schema)
	template.ID = id // Override UUID for deterministic testing
	return template
}

// createTestVersion creates a new version for testing purposes
// with a specified version number for deterministic testing
func createTestVersion(templateID string, number int) *config.Version {
	configData := json.RawMessage(`{"key": "value"}`)
	version := config.NewVersion(configData, templateID, "test-user")
	version.Number = number // Set for testing
	return version
}

// createTestDeployment creates a new deployment for testing purposes
// with deterministic IDs for easier testing
func createTestDeployment(id, tenantID, deviceID string) *config.Deployment {
	version := createTestVersion("template-1", 1)
	deployment := config.NewDeployment(tenantID, deviceID, version)
	deployment.ID = id // Override UUID for deterministic testing
	return deployment
}
