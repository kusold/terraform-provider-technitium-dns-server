package testhelpers

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// AcceptanceTestConfig holds configuration for acceptance tests
type AcceptanceTestConfig struct {
	Container *TechnitiumContainer
	Host      string
	Username  string
	Password  string
}

// SetupAcceptanceTest sets up an acceptance test with a fresh Technitium container
func SetupAcceptanceTest(t *testing.T) *AcceptanceTestConfig {
	t.Helper()

	// Skip acceptance tests unless explicitly requested
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	ctx := context.Background()
	container, err := StartTechnitiumContainer(ctx, t)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}

	t.Cleanup(func() {
		if err := container.Cleanup(ctx); err != nil {
			t.Logf("Warning: failed to cleanup container: %v", err)
		}
	})

	return &AcceptanceTestConfig{
		Container: container,
		Host:      container.GetAPIURL(),
		Username:  container.Username,
		Password:  container.Password,
	}
}

// GetProviderConfig returns the provider configuration for acceptance tests
func (c *AcceptanceTestConfig) GetProviderConfig() string {
	return fmt.Sprintf(`
provider "technitium" {
  host     = "%s"
  username = "%s"
  password = "%s"
}
`, c.Host, c.Username, c.Password)
}

// GetProviderFactories returns the provider factories for acceptance tests
func GetProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"technitium": func() (tfprotov6.ProviderServer, error) {
			// TODO: Import will be resolved when we remove the import cycle
			return nil, fmt.Errorf("provider factory not yet implemented - import cycle needs resolution")
		},
	}
}

// AcceptanceTestCase returns a standard acceptance test case
func (c *AcceptanceTestConfig) AcceptanceTestCase(config string, checks ...resource.TestCheckFunc) resource.TestCase {
	return resource.TestCase{
		ProtoV6ProviderFactories: GetProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: c.GetProviderConfig() + config,
				Check:  resource.ComposeAggregateTestCheckFunc(checks...),
			},
		},
	}
}

// CheckResourceExists verifies that a resource exists in the Technitium server
func CheckResourceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID not set")
		}

		// TODO: Add actual API calls to verify resource exists in Technitium
		// This will be implemented when we have the API client ready

		return nil
	}
}

// CheckResourceDestroyed verifies that a resource no longer exists in the Technitium server
func CheckResourceDestroyed(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}

			// TODO: Add actual API calls to verify resource is destroyed in Technitium
			// This will be implemented when we have the API client ready
		}
		return nil
	}
}

// WaitForTechnitiumReady waits for the Technitium server to be ready
func (c *AcceptanceTestConfig) WaitForTechnitiumReady(t *testing.T) {
	t.Helper()
	
	// TODO: Add actual health check against Technitium API
	// For now, we rely on the container's wait strategy
	t.Logf("Technitium server ready at %s", c.Host)
}