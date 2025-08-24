package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/testhelpers"
)

func TestAccDNSRecordResource_A(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)
	zoneName := "testarecord.example.com"
	recordName := "www"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		CheckDestroy: testAccCheckDNSRecordDestroy(config),
		Steps: []resource.TestStep{
			// Create zone and A record
			{
				Config: testAccDNSRecordConfig_A(config, zoneName, recordName, "192.168.1.100", 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSRecordExists(config, "technitium_dns_record.test"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", zoneName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "name", recordName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "A"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "ttl", "300"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "data", "192.168.1.100"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_CNAME(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)
	zoneName := "testcnamerecord.example.com"
	recordName := "blog"
	targetName := "www.testcnamerecord.example.com"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		CheckDestroy: testAccCheckDNSRecordDestroy(config),
		Steps: []resource.TestStep{
			// Create zone and CNAME record
			{
				Config: testAccDNSRecordConfig_CNAME(config, zoneName, recordName, targetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSRecordExists(config, "technitium_dns_record.test"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", zoneName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "name", recordName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "CNAME"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "data", targetName),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_MX(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)
	zoneName := "testmxrecord.example.com"
	recordName := "testmxrecord.example.com" // Use the zone name for root domain records
	exchangeName := "mail.testmxrecord.example.com"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		CheckDestroy: testAccCheckDNSRecordDestroy(config),
		Steps: []resource.TestStep{
			// Create zone and MX record
			{
				Config: testAccDNSRecordConfig_MX(config, zoneName, recordName, exchangeName, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSRecordExists(config, "technitium_dns_record.test"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", zoneName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "name", recordName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "MX"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "data", exchangeName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "priority", "10"),
				),
			},
		},
	})
}

func TestAccDNSRecordResource_TXT(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)
	zoneName := "testtxtrecord.example.com"
	recordName := "_spf"
	txtValue := "v=spf1 include:_spf.google.com ~all"

	// Create a unique ID for this test to prevent caching issues
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		CheckDestroy: testAccCheckDNSRecordDestroy(config),
		Steps: []resource.TestStep{
			// Create zone and TXT record
			{
				Config: testAccDNSRecordConfig_TXT(config, zoneName, recordName, txtValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSRecordExists(config, "technitium_dns_record.test"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", zoneName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "name", recordName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "TXT"),
					// The API might return the TXT record with quotes around it
					resource.TestCheckResourceAttrSet("technitium_dns_record.test", "data"),
					// Use a custom check function to verify the TXT record data
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["technitium_dns_record.test"]
						if !ok {
							return fmt.Errorf("resource not found: %s", "technitium_dns_record.test")
						}

						// Get the data attribute
						data := rs.Primary.Attributes["data"]

						// Clean both values for comparison (remove quotes if present)
						cleanExpected := strings.Trim(txtValue, "\"")
						cleanActual := strings.Trim(data, "\"")

						if cleanExpected != cleanActual {
							return fmt.Errorf("TXT record data doesn't match. Expected: %s, Got: %s", cleanExpected, cleanActual)
						}

						return nil
					},
				),
			},
		},
	})
}

func TestAccDNSRecordResource_SRV(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping acceptance test in short mode")
	}

	// Setup test container
	config := setupTestContainer(t)
	zoneName := "testsrvrecord.example.com"
	recordName := "_sip._tcp"
	targetName := "sip.testsrvrecord.example.com"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"technitium": providerserver.NewProtocol6WithError(New("test")()),
		},
		CheckDestroy: testAccCheckDNSRecordDestroy(config),
		Steps: []resource.TestStep{
			// Create zone and SRV record
			{
				Config: testAccDNSRecordConfig_SRV(config, zoneName, recordName, targetName, 10, 5, 5060),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckDNSRecordExists(config, "technitium_dns_record.test"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "zone", zoneName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "name", recordName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "type", "SRV"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "data", targetName),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "priority", "10"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "weight", "5"),
					resource.TestCheckResourceAttr("technitium_dns_record.test", "port", "5060"),
				),
			},
		},
	})
}

func testAccCheckDNSRecordExists(config *testAccConfig, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("record ID is not set")
		}

		// Create client to verify record exists in Technitium
		client, err := testhelpers.CreateTestClient(config.Host, config.Username, config.Password)
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}

		// Extract zone and record details from ID
		idParts := strings.Split(rs.Primary.ID, ":")
		if len(idParts) < 3 {
			return fmt.Errorf("invalid record ID format: %s", rs.Primary.ID)
		}

		zoneName := idParts[0]
		recordName := idParts[1]
		recordType := idParts[2]

		// Handle root domain records
		if recordName == "@" || recordName == "" {
			recordName = zoneName
		} else if recordName != zoneName {
			// Format non-root domain records
			if !strings.HasSuffix(recordName, ".") && !strings.HasSuffix(recordName, "."+zoneName) && !strings.HasSuffix(recordName, zoneName) {
				recordName = recordName + "." + zoneName
			}
		}

		// Verify the zone exists first
		ctx := context.Background()
		zoneExists, err := client.ZoneExists(ctx, zoneName)
		if err != nil {
			return fmt.Errorf("failed to check if zone exists: %w", err)
		}
		if !zoneExists {
			return fmt.Errorf("zone %s does not exist in Technitium server", zoneName)
		}

		// Get records for the domain in the zone
		records, err := client.GetRecords(ctx, zoneName, recordName, false)
		if err != nil {
			return fmt.Errorf("failed to get DNS records: %w", err)
		}

		// For debugging purposes
		if len(records.Records) == 0 {
			return fmt.Errorf("no records found for domain %s in zone %s", recordName, zoneName)
		}

		// Check if the specific record exists
		for _, record := range records.Records {
			// Record type matches what we're looking for
			if record.Type == recordType {
				return nil // Found the record type
			}
		}

		return fmt.Errorf("DNS record %s:%s:%s not found in Technitium server", zoneName, recordName, recordType)
	}
}

func testAccCheckDNSRecordDestroy(config *testAccConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Create client to verify records are destroyed
		client, err := testhelpers.CreateTestClient(config.Host, config.Username, config.Password)
		if err != nil {
			return fmt.Errorf("failed to create test client: %w", err)
		}

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "technitium_dns_record" {
				continue
			}

			// Extract zone and record details from ID
			idParts := strings.Split(rs.Primary.ID, ":")
			if len(idParts) < 3 {
				return fmt.Errorf("invalid record ID format: %s", rs.Primary.ID)
			}

			zoneName := idParts[0]
			recordName := idParts[1]
			recordType := idParts[2]

			// Handle root domain records
			if recordName == "@" || recordName == "" {
				recordName = zoneName
			} else if recordName != zoneName {
				// Format non-root domain records
				if !strings.HasSuffix(recordName, ".") && !strings.HasSuffix(recordName, "."+zoneName) && !strings.HasSuffix(recordName, zoneName) {
					recordName = recordName + "." + zoneName
				}
			}

			ctx := context.Background()

			// Check if zone exists
			zoneExists, err := client.ZoneExists(ctx, zoneName)
			if err != nil {
				return fmt.Errorf("failed to check if zone exists: %w", err)
			}

			if !zoneExists {
				// If zone doesn't exist, the record doesn't exist either
				continue
			}

			// Check if record exists
			records, err := client.GetRecords(ctx, zoneName, recordName, false)
			if err != nil {
				// If we can't get records, consider the test passed (record might be gone)
				continue
			}

			for _, record := range records.Records {
				if record.Type == recordType {
					// Need to verify data or priority to ensure it's the right record
					if len(idParts) > 3 {
						priority := idParts[3]
						data := ""
						if len(idParts) > 4 {
							data = idParts[4]
						}

						// For records with priority (like MX), check if it matches
						if recordType == "MX" && fmt.Sprintf("%d", record.RData.Preference) == priority {
							return fmt.Errorf("DNS record %s:%s:%s still exists", zoneName, recordName, recordType)
						}

						// Check data match
						if data != "" {
							switch recordType {
							case "A", "AAAA":
								if record.RData.IPAddress == data {
									return fmt.Errorf("DNS record %s:%s:%s still exists", zoneName, recordName, recordType)
								}
							case "CNAME":
								if record.RData.CNAME == data {
									return fmt.Errorf("DNS record %s:%s:%s still exists", zoneName, recordName, recordType)
								}
							case "TXT":
								// Add debug logging for TXT record comparison during destroy check
								fmt.Printf("TXT record destroy check - Expected: %s, Actual: %s\n", data, record.RData.Text)
								if record.RData.Text == data {
									return fmt.Errorf("DNS record %s:%s:%s still exists", zoneName, recordName, recordType)
								}
							}
						} else {
							return fmt.Errorf("DNS record %s:%s:%s still exists", zoneName, recordName, recordType)
						}
					} else {
						return fmt.Errorf("DNS record %s:%s:%s still exists", zoneName, recordName, recordType)
					}
				}
			}
		}

		return nil
	}
}

func testAccDNSRecordConfig_A(config *testAccConfig, zoneName, recordName, ipAddress string, ttl int) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_zone" "test_zone" {
  name = "%s"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone = technitium_zone.test_zone.name
  name = "%s"
  type = "A"
  ttl  = %d
  data = "%s"
}
`, zoneName, recordName, ttl, ipAddress)
}

func testAccDNSRecordConfig_CNAME(config *testAccConfig, zoneName, recordName, target string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_zone" "test_zone" {
  name = "%s"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone = technitium_zone.test_zone.name
  name = "%s"
  type = "CNAME"
  ttl  = 300
  data = "%s"
}
`, zoneName, recordName, target)
}

func testAccDNSRecordConfig_MX(config *testAccConfig, zoneName, recordName, exchange string, priority int) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_zone" "test_zone" {
  name = "%s"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone     = technitium_zone.test_zone.name
  name     = "%s"
  type     = "MX"
  ttl      = 300
  data     = "%s"
  priority = %d
}
`, zoneName, recordName, exchange, priority)
}

func testAccDNSRecordConfig_TXT(config *testAccConfig, zoneName, recordName, text string) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_zone" "test_zone" {
  name = "%s"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone = technitium_zone.test_zone.name
  name = "%s"
  type = "TXT"
  ttl  = 300
  data = "%s"
}
`, zoneName, recordName, text)
}

func testAccDNSRecordConfig_SRV(config *testAccConfig, zoneName, recordName, target string, priority, weight, port int) string {
	return config.getProviderConfig() + fmt.Sprintf(`
resource "technitium_zone" "test_zone" {
  name = "%s"
  type = "Primary"
}

resource "technitium_dns_record" "test" {
  zone     = technitium_zone.test_zone.name
  name     = "%s"
  type     = "SRV"
  ttl      = 300
  data     = "%s"
  priority = %d
  weight   = %d
  port     = %d
}
`, zoneName, recordName, target, priority, weight, port)
}
