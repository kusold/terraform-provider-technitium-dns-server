package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/testhelpers"
)

func TestAccDNSAppsDataSource_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Read testing - initially should be empty
			{
				Config: testAccDNSAppsDataSourceConfig_basic(config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.technitium_dns_apps.test", "apps.#"),
				),
			},
		},
	})
}

func TestAccDNSAppsDataSource_WithApp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)

	// Create mock ZIP file content for testing
	zipContent, err := testhelpers.CreateMockDNSAppZipBase64("test-data-app", "1.0.0")
	if err != nil {
		t.Fatalf("Failed to create mock ZIP content: %v", err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create an app first, then read the data source
			{
				Config: testAccDNSAppsDataSourceConfig_withApp(config, "test-data-app", zipContent),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.technitium_dns_apps.test", "apps.#", "1"),
					resource.TestCheckResourceAttr("data.technitium_dns_apps.test", "apps.0.name", "test-data-app"),
					resource.TestCheckResourceAttrSet("data.technitium_dns_apps.test", "apps.0.version"),
					resource.TestCheckResourceAttrSet("data.technitium_dns_apps.test", "apps.0.dns_apps.#"),
				),
			},
		},
	})
}

func TestAccDNSStoreAppsDataSource_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccDNSStoreAppsDataSourceConfig_basic(config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.technitium_dns_store_apps.test", "store_apps.#"),
				),
			},
		},
	})
}

func testAccDNSAppsDataSourceConfig_basic(config *testAccConfig) string {
	return config.getProviderConfig() + `
data "technitium_dns_apps" "test" {}
`
}

func testAccDNSAppsDataSourceConfig_withApp(config *testAccConfig, appName, fileContent string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_dns_app" "test_app" {
  name           = "%s"
  install_method = "file"
  file_content   = "%s"

  config = jsonencode({
    "displayName" = "%s Test App"
    "version" = "1.0.0"
    "description" = "Test DNS application for integration testing"
    "applicationRecordDataTemplate" = "127.0.0.1"
    "author" = "Test"
  })
}

data "technitium_dns_apps" "test" {
  depends_on = [technitium_dns_app.test_app]
}
`, appName, fileContent, appName)
}

func testAccDNSStoreAppsDataSourceConfig_basic(config *testAccConfig) string {
	return config.getProviderConfig() + `
data "technitium_dns_store_apps" "test" {}
`
}
