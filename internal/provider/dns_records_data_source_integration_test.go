package provider

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// randomInt generates a random integer between min and max
func randomInt(min, max int) int {
	return rand.New(rand.NewSource(time.Now().UnixNano())).Intn(max-min) + min
}

// TestAccDNSRecordsDataSource_Basic tests the technitium_dns_records data source with a real Technitium DNS Server
func TestAccDNSRecordsDataSource_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)

	// Generate a random zone name for testing
	testZoneName := fmt.Sprintf("test-records-%d.example.com", randomInt(1000, 9999))

	// Create a zone and some records first
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				// Create zone and multiple records
				Config: config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_zone" "test" {
  name = "%s"
  type = "Primary"
}

resource "technitium_dns_record" "a" {
  zone = technitium_zone.test.name
  name = technitium_zone.test.name
  type = "A"
  ttl  = 3600
  data = "192.168.1.10"
}

resource "technitium_dns_record" "www" {
  zone = technitium_zone.test.name
  name = "www"
  type = "A"
  ttl  = 3600
  data = "192.168.1.11"
}

resource "technitium_dns_record" "mail" {
  zone = technitium_zone.test.name
  name = "mail"
  type = "A"
  ttl  = 3600
  data = "192.168.1.20"
}

resource "technitium_dns_record" "mx" {
  zone = technitium_zone.test.name
  name = technitium_zone.test.name
  type = "MX"
  ttl  = 3600
  data = "mail.%s"
  priority = 10
}

resource "technitium_dns_record" "cname" {
  zone = technitium_zone.test.name
  name = "api"
  type = "CNAME"
  ttl  = 3600
  data = "www.%s"
}

data "technitium_dns_records" "all" {
  zone = technitium_zone.test.name
  depends_on = [
    technitium_dns_record.a,
    technitium_dns_record.www,
    technitium_dns_record.mail,
    technitium_dns_record.mx,
    technitium_dns_record.cname,
  ]
}

data "technitium_dns_records" "a_records" {
  zone = technitium_zone.test.name
  record_types = ["A"]
  depends_on = [
    technitium_dns_record.a,
    technitium_dns_record.www,
    technitium_dns_record.mail,
  ]
}

data "technitium_dns_records" "specific_domain" {
  zone = technitium_zone.test.name
  domain = "www.${technitium_zone.test.name}"
  depends_on = [
    technitium_dns_record.www,
  ]
}

output "all_records_count" {
  value = length(data.technitium_dns_records.all.records)
}

output "a_records_count" {
  value = length(data.technitium_dns_records.a_records.records)
}

output "specific_domain_records" {
  value = data.technitium_dns_records.specific_domain.records
}
`, testZoneName, testZoneName, testZoneName),
				Check: resource.ComposeTestCheckFunc(
					// Check that data source has all records (SOA + NS + 5 created = 7 minimum)
					resource.TestCheckOutput("all_records_count", "7"),

					// Check A records filtering (3 A records created)
					resource.TestCheckOutput("a_records_count", "3"),

					// Check specific domain filtering returns records
					resource.TestCheckResourceAttrSet("data.technitium_dns_records.specific_domain", "records.#"),
					resource.TestCheckResourceAttr("data.technitium_dns_records.specific_domain", "domain", fmt.Sprintf("www.%s", testZoneName)),
				),
			},
		},
	})
}
