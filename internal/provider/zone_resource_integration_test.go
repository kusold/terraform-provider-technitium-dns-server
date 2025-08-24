package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/testhelpers"
)

// testAccConfig represents test configuration
type testAccConfig struct {
	Host     string
	Username string
	Password string
}

// setupTestContainer sets up a test container for acceptance tests
func setupTestContainer(t *testing.T) *testAccConfig {
	t.Helper()

	// Skip acceptance tests unless explicitly requested
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	ctx := context.Background()
	container, err := testhelpers.StartTechnitiumContainer(ctx, t)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}

	t.Cleanup(func() {
		if err := container.Cleanup(ctx); err != nil {
			t.Logf("Warning: failed to cleanup container: %v", err)
		}
	})

	return &testAccConfig{
		Host:     container.GetAPIURL(),
		Username: container.Username,
		Password: container.Password,
	}
}

// getProviderConfig returns the provider configuration for acceptance tests
func (c *testAccConfig) getProviderConfig() string {
	return fmt.Sprintf(`
provider "technitium" {
  host     = "%s"
  username = "%s"
  password = "%s"
}
`, c.Host, c.Username, c.Password)
}

func TestAccZoneResource_Primary(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		CheckDestroy: testAccCheckZoneDestroy(config),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccZoneResourceConfig_primary(config, "test-primary.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZoneExists(config, "technitium_zone.test"),
					resource.TestCheckResourceAttr("technitium_zone.test", "name", "test-primary.example.com"),
					resource.TestCheckResourceAttr("technitium_zone.test", "type", "Primary"),
					resource.TestCheckResourceAttrSet("technitium_zone.test", "dnssec_status"),
					resource.TestCheckResourceAttrSet("technitium_zone.test", "internal"),
					resource.TestCheckResourceAttrSet("technitium_zone.test", "disabled"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "technitium_zone.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "test-primary.example.com",
			},
			// Update and Read testing
			{
				Config: testAccZoneResourceConfig_primaryWithOptions(config, "test-primary.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZoneExists(config, "technitium_zone.test"),
					resource.TestCheckResourceAttr("technitium_zone.test", "name", "test-primary.example.com"),
					resource.TestCheckResourceAttr("technitium_zone.test", "type", "Primary"),
					resource.TestCheckResourceAttr("technitium_zone.test", "use_soa_serial_date_scheme", "true"),
				),
			},
		},
	})
}

func TestAccZoneResource_Secondary(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// This test would require setting up a proper zone transfer
	// which is not feasible in the current test environment
	t.Skip("Skipping secondary zone test as it requires actual DNS zone transfers")
}

func TestAccZoneResource_Forwarder(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		CheckDestroy: testAccCheckZoneDestroy(config),
		Steps: []resource.TestStep{
			// Create forwarder zone
			{
				Config: testAccZoneResourceConfig_forwarder(config, "test-forwarder.example.com"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZoneExists(config, "technitium_zone.test"),
					resource.TestCheckResourceAttr("technitium_zone.test", "name", "test-forwarder.example.com"),
					resource.TestCheckResourceAttr("technitium_zone.test", "type", "Forwarder"),
					resource.TestCheckResourceAttr("technitium_zone.test", "forwarder", "8.8.8.8"),
					resource.TestCheckResourceAttr("technitium_zone.test", "protocol", "Udp"),
					resource.TestCheckResourceAttr("technitium_zone.test", "initialize_forwarder", "true"),
				),
			},
		},
	})
}

func testAccCheckZoneExists(config *testAccConfig, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("zone ID is not set")
		}

		// Create client to verify zone exists in Technitium
		client, err := testhelpers.CreateTestClient(config.Host, config.Username, config.Password)
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}

		zoneName := rs.Primary.ID
		ctx := context.Background()
		exists, err := client.ZoneExists(ctx, zoneName)
		if err != nil {
			return fmt.Errorf("failed to check if zone exists: %w", err)
		}

		if !exists {
			return fmt.Errorf("zone %s does not exist in Technitium server", zoneName)
		}

		return nil
	}
}

func testAccCheckZoneDestroy(config *testAccConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Create client to verify zones are destroyed
		client, err := testhelpers.CreateTestClient(config.Host, config.Username, config.Password)
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "technitium_zone" {
				continue
			}

			zoneName := rs.Primary.ID
			ctx := context.Background()
			exists, err := client.ZoneExists(ctx, zoneName)
			if err != nil {
				return fmt.Errorf("failed to check if zone exists: %w", err)
			}

			if exists {
				return fmt.Errorf("zone %s still exists in Technitium server", zoneName)
			}
		}

		return nil
	}
}

func testAccZoneResourceConfig_primary(config *testAccConfig, zoneName string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_zone" "test" {
  name = "%s"
  type = "Primary"
}
`, zoneName)
}

func testAccZoneResourceConfig_primaryWithOptions(config *testAccConfig, zoneName string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_zone" "test" {
  name                       = "%s"
  type                       = "Primary"
  use_soa_serial_date_scheme = true
}
`, zoneName)
}

func testAccZoneResourceConfig_forwarder(config *testAccConfig, zoneName string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_zone" "test" {
  name                 = "%s"
  type                 = "Forwarder"
  initialize_forwarder = true
  forwarder           = "8.8.8.8"
  protocol            = "Udp"
}
`, zoneName)
}
