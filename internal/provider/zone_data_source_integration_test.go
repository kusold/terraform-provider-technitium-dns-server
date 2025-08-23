package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/kusold/terraform-provider-technitium-dns-server/internal/testhelpers"
)

func TestAccZoneDataSource(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)
	zoneName := "testdatasource.example.com"

	// Create a zone first
	ctx := context.Background()
	client, err := testhelpers.CreateTestClient(config.Host, config.Username, config.Password)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.CreateZone(ctx, zoneName, "Primary"); err != nil {
		t.Fatal(err)
	}

	// Run the test
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccZoneDataSourceConfig(config, zoneName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.technitium_zone.test", "name", zoneName),
					resource.TestCheckResourceAttr("data.technitium_zone.test", "id", zoneName),
					resource.TestCheckResourceAttr("data.technitium_zone.test", "type", "Primary"),
					resource.TestCheckResourceAttr("data.technitium_zone.test", "internal", "false"),
				),
			},
		},
	})
}

func testAccZoneDataSourceConfig(config *testAccConfig, zoneName string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
data "technitium_zone" "test" {
  name = "%s"
}
`, zoneName)
}
