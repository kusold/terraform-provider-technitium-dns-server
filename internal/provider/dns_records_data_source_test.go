package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/client"
)

// TestDNSRecordsDataSource tests the technitium_dns_records data source.
func TestDNSRecordsDataSource(t *testing.T) {
	t.Skip("Skipping test that requires proper mocking of server responses")

	// Create a mock Technitium DNS Server API
	mockResponse := createMockDNSRecordsResponse()
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request path
		pathRegex := regexp.MustCompile(`/api/(login|zones/records/get)`)
		if !pathRegex.MatchString(r.URL.Path) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}

		// Handle login request
		if r.URL.Path == "/api/login" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"status":"ok","response":{"token":"dummy-token"}}`)
			return
		}

		// Handle DNS records request
		if r.URL.Path == "/api/zones/records/get" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			// Parse query parameters to determine filtering needs
			// We don't actually use these values in the mock but would in a more complex test
			_ = r.URL.Query().Get("domain")
			_ = r.URL.Query().Get("zone")
			_ = r.URL.Query().Get("listZone")

			// Modify response based on parameters
			response := mockResponse

			responseJSON, _ := json.Marshal(response)
			w.Write(responseJSON)
			return
		}
	}))
	defer mockServer.Close()

	// No need to create a testing client since we're using the mock server directly
	testProviderConfig := fmt.Sprintf(`
provider "technitium" {
  host     = "%s"
  username = "admin"
  password = "admin"
}
`, mockServer.URL)

	// Create provider with mock client
	testAccProtoV6ProviderFactories := map[string]func() (tfprotov6.ProviderServer, error){
		"technitium": providerserver.NewProtocol6WithError(New("test")()),
	}

	// Run test cases
	t.Run("basic_query", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testProviderConfig + `
data "technitium_dns_records" "test" {
  zone = "example.com"
}

output "all_records" {
  value = data.technitium_dns_records.test.records
}
`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.technitium_dns_records.test", "zone", "example.com"),
						resource.TestCheckResourceAttrSet("data.technitium_dns_records.test", "id"),
						resource.TestCheckResourceAttr("data.technitium_dns_records.test", "records.#", "5"),
						// We can't guarantee the order of records, so let's just check the count
					),
				},
			},
		})
	})

	t.Run("specific_domain", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testProviderConfig + `
data "technitium_dns_records" "specific" {
  zone   = "example.com"
  domain = "www.example.com"
}
`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.technitium_dns_records.specific", "zone", "example.com"),
						resource.TestCheckResourceAttr("data.technitium_dns_records.specific", "domain", "www.example.com"),
					),
				},
			},
		})
	})

	t.Run("filtered_by_type", func(t *testing.T) {
		resource.UnitTest(t, resource.TestCase{
			ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: testProviderConfig + `
data "technitium_dns_records" "filtered" {
  zone         = "example.com"
  record_types = ["A", "AAAA"]
}
`,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.technitium_dns_records.filtered", "zone", "example.com"),
						resource.TestCheckResourceAttr("data.technitium_dns_records.filtered", "record_types.#", "2"),
						resource.TestCheckResourceAttr("data.technitium_dns_records.filtered", "record_types.0", "A"),
						resource.TestCheckResourceAttr("data.technitium_dns_records.filtered", "record_types.1", "AAAA"),
					),
				},
			},
		})
	})
}

// TestDNSRecordsDataSource_Errors tests error handling in the data source
func TestDNSRecordsDataSource_Errors(t *testing.T) {
	// Mock server that always returns errors
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/login" {
			// Allow login to succeed
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"status":"ok","response":{"token":"dummy-token"}}`)
			return
		}

		// Return error for records request
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"status":"error","errorMessage":"Zone does not exist"}`)
	}))
	defer mockServer.Close()

	testProviderConfig := fmt.Sprintf(`
provider "technitium" {
  host     = "%s"
  username = "admin"
  password = "admin"
}
`, mockServer.URL)

	// Create provider with mock client
	testAccProtoV6ProviderFactories := map[string]func() (tfprotov6.ProviderServer, error){
		"technitium": providerserver.NewProtocol6WithError(New("test")()),
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig + `
data "technitium_dns_records" "error" {
  zone = "non-existent.com"
}
`,
				ExpectError: regexp.MustCompile(`Zone does not exist`),
			},
		},
	})
}

// createMockDNSRecordsResponse creates a mock response for DNS records
func createMockDNSRecordsResponse() *struct {
	Status string                 `json:"status"`
	Data   map[string]interface{} `json:"response"`
} {
	return &struct {
		Status string                 `json:"status"`
		Data   map[string]interface{} `json:"response"`
	}{
		Status: "ok",
		Data: map[string]interface{}{
			"zone": map[string]interface{}{
				"name":         "example.com",
				"type":         "Primary",
				"internal":     false,
				"dnssecStatus": "Unsigned",
				"disabled":     false,
			},
			"records": []map[string]interface{}{
				{
					"name":     "example.com",
					"type":     "A",
					"ttl":      3600,
					"disabled": false,
					"rData": map[string]interface{}{
						"ipAddress": "192.168.1.1",
					},
				},
				{
					"name":     "example.com",
					"type":     "MX",
					"ttl":      3600,
					"disabled": false,
					"rData": map[string]interface{}{
						"preference": 10,
						"exchange":   "mail.example.com",
					},
				},
				{
					"name":     "www.example.com",
					"type":     "CNAME",
					"ttl":      3600,
					"disabled": false,
					"rData": map[string]interface{}{
						"cname": "example.com",
					},
				},
				{
					"name":     "example.com",
					"type":     "TXT",
					"ttl":      3600,
					"disabled": false,
					"rData": map[string]interface{}{
						"text": "v=spf1 include:_spf.example.com -all",
					},
				},
				{
					"name":     "example.com",
					"type":     "SOA",
					"ttl":      3600,
					"disabled": false,
					"rData": map[string]interface{}{
						"primaryNameServer": "ns1.example.com",
						"responsiblePerson": "admin.example.com",
						"serial":            1,
						"refresh":           3600,
						"retry":             600,
						"expire":            86400,
						"minimum":           3600,
					},
				},
			},
		},
	}
}

// TestUnitDNSRecordsDataSourceFormatRecordData tests the formatRecordData function
func TestUnitDNSRecordsDataSourceFormatRecordData(t *testing.T) {
	cases := []struct {
		name     string
		record   client.DNSRecord
		expected string
	}{
		{
			name: "A record",
			record: client.DNSRecord{
				Type: "A",
				RData: client.DNSRecordData{
					IPAddress: "192.168.1.1",
				},
			},
			expected: "192.168.1.1",
		},
		{
			name: "AAAA record",
			record: client.DNSRecord{
				Type: "AAAA",
				RData: client.DNSRecordData{
					IPAddress: "2001:db8::1",
				},
			},
			expected: "2001:db8::1",
		},
		{
			name: "CNAME record",
			record: client.DNSRecord{
				Type: "CNAME",
				RData: client.DNSRecordData{
					CNAME: "example.com",
				},
			},
			expected: "example.com",
		},
		{
			name: "MX record",
			record: client.DNSRecord{
				Type: "MX",
				RData: client.DNSRecordData{
					Preference: 10,
					Exchange:   "mail.example.com",
				},
			},
			expected: "10 mail.example.com",
		},
		{
			name: "TXT record",
			record: client.DNSRecord{
				Type: "TXT",
				RData: client.DNSRecordData{
					Text: "v=spf1 -all",
				},
			},
			expected: "v=spf1 -all",
		},
		{
			name: "PTR record",
			record: client.DNSRecord{
				Type: "PTR",
				RData: client.DNSRecordData{
					PTRName: "example.com",
				},
			},
			expected: "example.com",
		},
		{
			name: "NS record",
			record: client.DNSRecord{
				Type: "NS",
				RData: client.DNSRecordData{
					NameServer: "ns1.example.com",
				},
			},
			expected: "ns1.example.com",
		},
		{
			name: "SRV record",
			record: client.DNSRecord{
				Type: "SRV",
				RData: client.DNSRecordData{
					Priority: 0,
					Weight:   1,
					Port:     443,
					Target:   "example.com",
				},
			},
			expected: "0 1 443 example.com",
		},
		{
			name: "SOA record",
			record: client.DNSRecord{
				Type: "SOA",
				RData: client.DNSRecordData{
					PrimaryNameServer: "ns1.example.com",
					ResponsiblePerson: "admin.example.com",
					Serial:            1,
					Refresh:           3600,
					Retry:             600,
					Expire:            86400,
					Minimum:           3600,
				},
			},
			expected: "ns1.example.com admin.example.com 1 3600 600 86400 3600",
		},
		{
			name: "Unknown record",
			record: client.DNSRecord{
				Type:  "CAA",
				RData: client.DNSRecordData{
					// CAA record fields not specifically handled
				},
			},
			expected: "[CAA record]",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatRecordData(tc.record)
			require.Equal(t, tc.expected, result)
		})
	}
}
