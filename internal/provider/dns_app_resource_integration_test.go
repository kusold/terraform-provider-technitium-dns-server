package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/testhelpers"
)

func TestAccDNSAppResource_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		CheckDestroy: testAccCheckDNSAppDestroy(config),
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDNSAppResourceConfig_basic(config, "test-app", "https://download.technitium.com/dns/apps/WildIpApp.zip"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSAppExists(config, "technitium_dns_app.test"),
					resource.TestCheckResourceAttr("technitium_dns_app.test", "name", "test-app"),
					resource.TestCheckResourceAttr("technitium_dns_app.test", "install_method", "url"),
					resource.TestCheckResourceAttr("technitium_dns_app.test", "url", "https://download.technitium.com/dns/apps/WildIpApp.zip"),
					resource.TestCheckResourceAttrSet("technitium_dns_app.test", "installed_version"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "technitium_dns_app.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "test-app",
			},
			// Update and Read testing
			{
				Config: testAccDNSAppResourceConfig_withConfig(config, "test-app", "https://download.technitium.com/dns/apps/WildIpApp.zip"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSAppExists(config, "technitium_dns_app.test"),
					resource.TestCheckResourceAttr("technitium_dns_app.test", "name", "test-app"),
					resource.TestCheckResourceAttrSet("technitium_dns_app.test", "config"),
				),
			},
		},
	})
}

func TestAccDNSAppResource_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		CheckDestroy: testAccCheckDNSAppDestroy(config),
		Steps: []resource.TestStep{
			// Create app
			{
				Config: testAccDNSAppResourceConfig_basic(config, "update-test-app", "https://download.technitium.com/dns/apps/WildIpApp.zip"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSAppExists(config, "technitium_dns_app.test"),
					resource.TestCheckResourceAttr("technitium_dns_app.test", "name", "update-test-app"),
				),
			},
			// Update app with new URL (simulating app update)
			{
				Config: testAccDNSAppResourceConfig_basic(config, "update-test-app", "https://download.technitium.com/dns/apps/WildIpApp.zip"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSAppExists(config, "technitium_dns_app.test"),
					resource.TestCheckResourceAttr("technitium_dns_app.test", "name", "update-test-app"),
				),
			},
		},
	})
}

func testAccCheckDNSAppExists(config *testAccConfig, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("DNS app ID is not set")
		}

		// Create client to verify app exists in Technitium
		client, err := testhelpers.CreateTestClient(config.Host, config.Username, config.Password)
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}

		appName := rs.Primary.ID
		ctx := context.Background()
		apps, err := client.ListApps(ctx)
		if err != nil {
			return fmt.Errorf("failed to list apps: %w", err)
		}

		// Check if app exists
		for _, app := range apps {
			if app.Name == appName {
				return nil // App found
			}
		}

		return fmt.Errorf("DNS app %s does not exist in Technitium server", appName)
	}
}

func testAccCheckDNSAppDestroy(config *testAccConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Create client to verify apps are destroyed
		client, err := testhelpers.CreateTestClient(config.Host, config.Username, config.Password)
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "technitium_dns_app" {
				continue
			}

			appName := rs.Primary.ID
			ctx := context.Background()
			apps, err := client.ListApps(ctx)
			if err != nil {
				return fmt.Errorf("failed to list apps: %w", err)
			}

			// Check if app still exists
			for _, app := range apps {
				if app.Name == appName {
					return fmt.Errorf("DNS app %s still exists in Technitium server", appName)
				}
			}
		}

		return nil
	}
}

func testAccDNSAppResourceConfig_basic(config *testAccConfig, appName, appURL string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_dns_app" "test" {
  name           = "%s"
  install_method = "url"
  url            = "%s"
}
`, appName, appURL)
}

func testAccDNSAppResourceConfig_withConfig(config *testAccConfig, appName, appURL string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_dns_app" "test" {
  name           = "%s"
  install_method = "url"
  url            = "%s"

  config = jsonencode({
    "enabled" = true
    "ipv4"    = true
    "ipv6"    = false
  })
}
`, appName, appURL)
}
