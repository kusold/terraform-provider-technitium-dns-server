package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create an app first, then read the data source
			{
				Config: testAccDNSAppsDataSourceConfig_withApp(config, "test-data-app", "https://download.technitium.com/dns/apps/WildIpApp.zip"),
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

func testAccDNSAppsDataSourceConfig_withApp(config *testAccConfig, appName, appURL string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_dns_app" "test_app" {
  name           = "%s"
  install_method = "url"
  url            = "%s"
}

data "technitium_dns_apps" "test" {
  depends_on = [technitium_dns_app.test_app]
}
`, appName, appURL)
}

func testAccDNSStoreAppsDataSourceConfig_basic(config *testAccConfig) string {
	return config.getProviderConfig() + `
data "technitium_dns_store_apps" "test" {}
`
}
