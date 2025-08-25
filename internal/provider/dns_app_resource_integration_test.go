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

	// Create mock ZIP file content for testing
	zipContent, err := testhelpers.CreateMockDNSAppZipBase64("test-app", "1.0.0")
	if err != nil {
		t.Fatalf("Failed to create mock ZIP content: %v", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		CheckDestroy: testAccCheckDNSAppDestroy(config),
		Steps: []resource.TestStep{
			// Create and Read testing - include the config to match what Technitium returns
			{
				Config: testAccDNSAppResourceConfig_fileWithDefaultConfig(config, "test-app", zipContent),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSAppExists(config, "technitium_dns_app.test"),
					resource.TestCheckResourceAttr("technitium_dns_app.test", "name", "test-app"),
					resource.TestCheckResourceAttr("technitium_dns_app.test", "install_method", "file"),
					resource.TestCheckResourceAttrSet("technitium_dns_app.test", "version"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "technitium_dns_app.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           "test-app",
				ImportStateVerifyIgnore: []string{"install_method", "file_content", "url"},
			},
			// Update and Read testing
			{
				Config: testAccDNSAppResourceConfig_fileWithConfig(config, "test-app", zipContent),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSAppExists(config, "technitium_dns_app.test"),
					resource.TestCheckResourceAttr("technitium_dns_app.test", "name", "test-app"),
					resource.TestCheckResourceAttr("technitium_dns_app.test", "install_method", "file"),
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

	// Create mock ZIP file content for testing
	zipContent1, err := testhelpers.CreateMockDNSAppZipBase64("update-test-app", "1.0.0")
	if err != nil {
		t.Fatalf("Failed to create mock ZIP content 1: %v", err)
	}

	zipContent2, err := testhelpers.CreateMockDNSAppZipBase64("update-test-app", "1.1.0")
	if err != nil {
		t.Fatalf("Failed to create mock ZIP content 2: %v", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		CheckDestroy: testAccCheckDNSAppDestroy(config),
		Steps: []resource.TestStep{
			// Create app
			{
				Config: testAccDNSAppResourceConfig_fileWithDefaultConfigUpdate(config, "update-test-app", zipContent1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSAppExists(config, "technitium_dns_app.test"),
					resource.TestCheckResourceAttr("technitium_dns_app.test", "name", "update-test-app"),
				),
			},
			// Update app with new file (simulating app update)
			{
				Config: testAccDNSAppResourceConfig_fileWithDefaultConfigUpdate2(config, "update-test-app", zipContent2),
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

func testAccDNSAppResourceConfig_fileWithDefaultConfig(config *testAccConfig, appName, fileContent string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_dns_app" "test" {
  name           = "%s"
  install_method = "file"
  file_content   = "%s"
}
`, appName, fileContent)
}

func testAccDNSAppResourceConfig_fileWithDefaultConfigUpdate(config *testAccConfig, appName, fileContent string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_dns_app" "test" {
  name           = "%s"
  install_method = "file"
  file_content   = "%s"
}
`, appName, fileContent)
}

func testAccDNSAppResourceConfig_fileWithDefaultConfigUpdate2(config *testAccConfig, appName, fileContent string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_dns_app" "test" {
  name           = "%s"
  install_method = "file"
  file_content   = "%s"
}
`, appName, fileContent)
}

func testAccDNSAppResourceConfig_fileWithConfig(config *testAccConfig, appName, fileContent string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_dns_app" "test" {
  name           = "%s"
  install_method = "file"
  file_content   = "%s"
}
`, appName, fileContent)
}
