package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDNSRecordResource_FWD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)
	zoneName := "testfwdrecord.example.com"
	recordName := "forward"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		CheckDestroy: testAccCheckDNSRecordDestroy(config),
		Steps: []resource.TestStep{
			// Create zone and FWD record
			{
				Config: testAccDNSRecordConfig_FWD(config, zoneName, recordName, "8.8.8.8", "Udp"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSRecordExists(config, "technitium_dns_record.test"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", zoneName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "name", recordName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "FWD"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "data", "8.8.8.8"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "protocol", "Udp"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "forwarder", "8.8.8.8"),
				),
			},
			// Update FWD record
			{
				Config: testAccDNSRecordConfig_FWD(config, zoneName, recordName, "1.1.1.1", "Https"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSRecordExists(config, "technitium_dns_record.test"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", zoneName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "name", recordName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "FWD"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "data", "1.1.1.1"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "protocol", "Https"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "forwarder", "1.1.1.1"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_FWD_Advanced(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)
	zoneName := "testfwdadvanced.example.com"
	recordName := "advanced"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		CheckDestroy: testAccCheckDNSRecordDestroy(config),
		Steps: []resource.TestStep{
			// Create zone and advanced FWD record
			{
				Config: testAccDNSRecordConfig_FWD_Advanced(config, zoneName, recordName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSRecordExists(config, "technitium_dns_record.test"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", zoneName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "name", recordName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "FWD"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "forwarder", "9.9.9.9"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "protocol", "Tls"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "dnssec_validation", "true"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "forwarder_priority", "10"),
				),
			},
		},
	})
}

func testAccDNSRecordConfig_FWD(config *testAccConfig, zoneName, recordName, forwarder, protocol string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_zone" "test_zone" {
  name = "%s"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone      = technitium_zone.test_zone.name
  name      = "%s"
  type      = "FWD"
  ttl       = 3600
  data      = "%s"
  protocol  = "%s"
  forwarder = "%s"
}
`, zoneName, recordName, forwarder, protocol, forwarder)
}

func testAccDNSRecordConfig_FWD_Advanced(config *testAccConfig, zoneName, recordName string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_zone" "test_zone" {
  name = "%s"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone               = technitium_zone.test_zone.name
  name               = "%s"
  type               = "FWD"
  ttl                = 1800
  forwarder          = "9.9.9.9"
  protocol           = "Tls"
  forwarder_priority = 10
  dnssec_validation  = true
  comments           = "Advanced FWD record with TLS"
}
`, zoneName, recordName)
}
