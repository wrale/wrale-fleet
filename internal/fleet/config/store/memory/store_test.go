package memory

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wrale/fleet/internal/fleet/config"
)

// TestNew verifies proper store initialization
func TestNew(t *testing.T) {
	store := New()
	assert.NotNil(t, store.templates)
	assert.NotNil(t, store.versions)
	assert.NotNil(t, store.deployments)
}

// TestStore_Integration tests the interaction between different operations
func TestStore_Integration(t *testing.T) {
	store := New()
	ctx := context.Background()

	// Create initial template
	template := createTestTemplate("test-1", "tenant-1")
	require.NoError(t, store.CreateTemplate(ctx, template))

	// Add multiple versions to the template
	t.Run("Version Management", func(t *testing.T) {
		// Create initial version
		v1 := createTestVersion(template.ID, 0)
		require.NoError(t, store.CreateVersion(ctx, template.TenantID, template.ID, v1))
		assert.Equal(t, 1, v1.Number)

		// Create second version
		v2 := createTestVersion(template.ID, 0)
		require.NoError(t, store.CreateVersion(ctx, template.TenantID, template.ID, v2))
		assert.Equal(t, 2, v2.Number)

		// List and verify versions
		versions, err := store.ListVersions(ctx, template.TenantID, template.ID)
		require.NoError(t, err)
		assert.Len(t, versions, 2)
	})

	// Test deployment lifecycle
	t.Run("Deployment Lifecycle", func(t *testing.T) {
		// Create deployment
		deployment := createTestDeployment("deploy-1", template.TenantID, "device-1")
		require.NoError(t, store.CreateDeployment(ctx, deployment))

		// Update deployment status
		deployment.Status = "completed"
		require.NoError(t, store.UpdateDeployment(ctx, deployment))

		// Verify deployment
		updated, err := store.GetDeployment(ctx, deployment.TenantID, deployment.ID)
		require.NoError(t, err)
		assert.Equal(t, "completed", updated.Status)

		// List deployments
		deployments, err := store.ListDeployments(ctx, config.ListOptions{
			TenantID: template.TenantID,
		})
		require.NoError(t, err)
		assert.Len(t, deployments, 1)
	})

	// Test template updates with versions
	t.Run("Template Updates", func(t *testing.T) {
		// Update template
		template.Name = "Updated Template"
		require.NoError(t, store.UpdateTemplate(ctx, template))

		// Verify template update didn't affect versions
		versions, err := store.ListVersions(ctx, template.TenantID, template.ID)
		require.NoError(t, err)
		assert.Len(t, versions, 2)

		// Delete template and verify cascade
		require.NoError(t, store.DeleteTemplate(ctx, template.TenantID, template.ID))
		_, err = store.GetTemplate(ctx, template.TenantID, template.ID)
		require.Error(t, err)

		// Verify versions are deleted
		versions, err = store.ListVersions(ctx, template.TenantID, template.ID)
		require.Error(t, err)
		assert.Nil(t, versions)
	})
}

// TestStore_Concurrency tests thread safety of the store
func TestStore_Concurrency(t *testing.T) {
	store := New()
	ctx := context.Background()

	// Create base template
	template := createTestTemplate("concurrent", "tenant-concurrent")
	require.NoError(t, store.CreateTemplate(ctx, template))

	t.Run("Concurrent Template Operations", func(t *testing.T) {
		var wg sync.WaitGroup
		operations := 100

		// Concurrent template reads
		wg.Add(operations)
		for i := 0; i < operations; i++ {
			go func() {
				defer wg.Done()
				_, _ = store.GetTemplate(ctx, template.TenantID, template.ID)
			}()
		}

		// Concurrent template updates
		wg.Add(operations)
		for i := 0; i < operations; i++ {
			go func(i int) {
				defer wg.Done()
				tmpl := createTestTemplate(template.ID, template.TenantID)
				tmpl.Name = fmt.Sprintf("Updated Name %d", i)
				_ = store.UpdateTemplate(ctx, tmpl)
			}(i)
		}

		// Wait with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for concurrent template operations")
		}

		// Verify template still accessible
		stored, err := store.GetTemplate(ctx, template.TenantID, template.ID)
		require.NoError(t, err)
		assert.NotNil(t, stored)
	})

	t.Run("Concurrent Version Operations", func(t *testing.T) {
		var wg sync.WaitGroup
		operations := 50

		// Concurrent version creation
		wg.Add(operations)
		for i := 0; i < operations; i++ {
			go func() {
				defer wg.Done()
				version := createTestVersion(template.ID, 0)
				_ = store.CreateVersion(ctx, template.TenantID, template.ID, version)
			}()
		}

		// Wait with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for concurrent version operations")
		}

		// Verify versions
		versions, err := store.ListVersions(ctx, template.TenantID, template.ID)
		require.NoError(t, err)
		assert.Len(t, versions, operations)

		// Verify version numbers are sequential
		numbers := make(map[int]bool)
		for _, v := range versions {
			assert.False(t, numbers[v.Number], "duplicate version number found")
			numbers[v.Number] = true
		}
	})

	t.Run("Concurrent Deployment Operations", func(t *testing.T) {
		var wg sync.WaitGroup
		operations := 50

		// Create initial deployment
		deployment := createTestDeployment("deploy-concurrent", template.TenantID, "device-concurrent")
		require.NoError(t, store.CreateDeployment(ctx, deployment))

		// Concurrent deployment updates
		wg.Add(operations)
		for i := 0; i < operations; i++ {
			go func(i int) {
				defer wg.Done()
				d := createTestDeployment(deployment.ID, deployment.TenantID, deployment.DeviceID)
				d.Status = fmt.Sprintf("status-%d", i)
				_ = store.UpdateDeployment(ctx, d)
			}(i)
		}

		// Wait with timeout
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for concurrent deployment operations")
		}

		// Verify deployment is still accessible
		stored, err := store.GetDeployment(ctx, deployment.TenantID, deployment.ID)
		require.NoError(t, err)
		assert.NotNil(t, stored)
	})
}

// TestStore_ErrorHandling tests error scenarios across operations
func TestStore_ErrorHandling(t *testing.T) {
	store := New()
	ctx := context.Background()

	t.Run("Invalid Operations", func(t *testing.T) {
		// Test nil template creation
		err := store.CreateTemplate(ctx, nil)
		require.Error(t, err)

		// Test nil version creation
		err = store.CreateVersion(ctx, "tenant-1", "template-1", nil)
		require.Error(t, err)

		// Test nil deployment creation
		err = store.CreateDeployment(ctx, nil)
		require.Error(t, err)
	})

	t.Run("Resource Not Found", func(t *testing.T) {
		// Test getting non-existent resources
		_, err := store.GetTemplate(ctx, "missing", "missing")
		require.Error(t, err)

		_, err = store.GetVersion(ctx, "missing", "missing", 1)
		require.Error(t, err)

		_, err = store.GetDeployment(ctx, "missing", "missing")
		require.Error(t, err)
	})

	t.Run("Invalid Updates", func(t *testing.T) {
		// Test updating non-existent resources
		err := store.UpdateTemplate(ctx, createTestTemplate("missing", "tenant-1"))
		require.Error(t, err)

		err = store.UpdateVersion(ctx, "tenant-1", "template-1", createTestVersion("template-1", 99))
		require.Error(t, err)

		err = store.UpdateDeployment(ctx, createTestDeployment("missing", "tenant-1", "device-1"))
		require.Error(t, err)
	})
}
