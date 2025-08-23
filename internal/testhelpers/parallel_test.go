package testhelpers

import (
	"context"
	"sync"
	"testing"
)

// ParallelTestRunner manages parallel test execution with TestContainers
type ParallelTestRunner struct {
	containers map[string]*TechnitiumContainer
	mutex      sync.RWMutex
}

// NewParallelTestRunner creates a new parallel test runner
func NewParallelTestRunner() *ParallelTestRunner {
	return &ParallelTestRunner{
		containers: make(map[string]*TechnitiumContainer),
	}
}

// GetContainer returns an existing container or creates a new one for the test
func (r *ParallelTestRunner) GetContainer(ctx context.Context, t *testing.T, testName string) (*TechnitiumContainer, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if container, exists := r.containers[testName]; exists {
		return container, nil
	}

	container, err := StartTechnitiumContainer(ctx, t)
	if err != nil {
		return nil, err
	}

	r.containers[testName] = container
	return container, nil
}

// CleanupAll cleans up all containers
func (r *ParallelTestRunner) CleanupAll(ctx context.Context) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for name, container := range r.containers {
		if err := container.Cleanup(ctx); err != nil {
			// Log error but don't fail cleanup
			continue
		}
		delete(r.containers, name)
	}
}

// Global test runner instance
var globalTestRunner = NewParallelTestRunner()

// SetupParallelTest sets up a test to run in parallel with proper container management
func SetupParallelTest(t *testing.T) (*TechnitiumContainer, func()) {
	t.Helper()
	t.Parallel()

	ctx := context.Background()
	container, err := globalTestRunner.GetContainer(ctx, t, t.Name())
	if err != nil {
		t.Fatalf("Failed to setup test container: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		if err := container.Cleanup(ctx); err != nil {
			t.Logf("Warning: failed to cleanup container: %v", err)
		}
	}

	return container, cleanup
}

// CleanupAllTestContainers should be called in TestMain to cleanup all containers
func CleanupAllTestContainers() {
	ctx := context.Background()
	globalTestRunner.CleanupAll(ctx)
}