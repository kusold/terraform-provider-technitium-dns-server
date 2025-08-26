package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDNSRecordResource(t *testing.T) {
	t.Parallel()

	// Unit test - verify resource creation
	t.Run("NewDNSRecordResource", func(t *testing.T) {
		r := NewDNSRecordResource()
		if r == nil {
			t.Fatal("NewDNSRecordResource should return a non-nil resource")
		}

		// Test metadata
		var resp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{
			ProviderTypeName: "technitium",
		}, &resp)

		if resp.TypeName != "technitium_dns_record" {
			t.Errorf("Expected TypeName to be technitium_dns_record, got %s", resp.TypeName)
		}
	})

	// Unit test - verify schema
	t.Run("Schema", func(t *testing.T) {
		r := NewDNSRecordResource()
		var resp resource.SchemaResponse
		r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Schema validation failed: %v", resp.Diagnostics.Errors())
		}

		// Verify required attributes
		schema := resp.Schema
		if _, ok := schema.Attributes["zone"]; !ok {
			t.Error("Schema should have 'zone' attribute")
		}
		if _, ok := schema.Attributes["name"]; !ok {
			t.Error("Schema should have 'name' attribute")
		}
		if _, ok := schema.Attributes["type"]; !ok {
			t.Error("Schema should have 'type' attribute")
		}
		if _, ok := schema.Attributes["ttl"]; !ok {
			t.Error("Schema should have 'ttl' attribute")
		}
		if _, ok := schema.Attributes["data"]; !ok {
			t.Error("Schema should have 'data' attribute")
		}

		// Verify optional attributes
		if attr, ok := schema.Attributes["priority"]; ok {
			if !attr.IsOptional() {
				t.Error("'priority' attribute should be optional")
			}
		} else {
			t.Error("Schema should have 'priority' attribute")
		}

		// Verify computed attributes
		if attr, ok := schema.Attributes["disabled"]; ok {
			if !attr.IsComputed() {
				t.Error("'disabled' attribute should be computed")
			}
		} else {
			t.Error("Schema should have 'disabled' attribute")
		}

		if attr, ok := schema.Attributes["dnssec_status"]; ok {
			if !attr.IsComputed() {
				t.Error("'dnssec_status' attribute should be computed")
			}
		} else {
			t.Error("Schema should have 'dnssec_status' attribute")
		}

		// Verify FWD record specific attributes
		if _, ok := schema.Attributes["protocol"]; !ok {
			t.Error("Schema should have 'protocol' attribute for FWD records")
		}
		if _, ok := schema.Attributes["forwarder"]; !ok {
			t.Error("Schema should have 'forwarder' attribute for FWD records")
		}
		if _, ok := schema.Attributes["forwarder_priority"]; !ok {
			t.Error("Schema should have 'forwarder_priority' attribute for FWD records")
		}
		if _, ok := schema.Attributes["dnssec_validation"]; !ok {
			t.Error("Schema should have 'dnssec_validation' attribute for FWD records")
		}
		if _, ok := schema.Attributes["proxy_type"]; !ok {
			t.Error("Schema should have 'proxy_type' attribute for FWD records")
		}
		if _, ok := schema.Attributes["proxy_address"]; !ok {
			t.Error("Schema should have 'proxy_address' attribute for FWD records")
		}
		if _, ok := schema.Attributes["proxy_port"]; !ok {
			t.Error("Schema should have 'proxy_port' attribute for FWD records")
		}
		if _, ok := schema.Attributes["proxy_username"]; !ok {
			t.Error("Schema should have 'proxy_username' attribute for FWD records")
		}
		if attr, ok := schema.Attributes["proxy_password"]; ok {
			if !attr.IsSensitive() {
				t.Error("'proxy_password' attribute should be sensitive")
			}
		} else {
			t.Error("Schema should have 'proxy_password' attribute for FWD records")
		}
	})

	// Unit test - validate record function
	t.Run("ValidateRecord", func(t *testing.T) {
		r := &DNSRecordResource{}

		// Test A record validation
		t.Run("A Record Valid", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type: types.StringValue("A"),
				Data: types.StringValue("192.168.1.1"),
			}
			options := map[string]string{"ipAddress": "192.168.1.1"}

			err := r.validateRecord(data, options)
			if err != nil {
				t.Errorf("Expected no error for valid A record, got: %v", err)
			}
		})

		t.Run("A Record Invalid", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type: types.StringValue("A"),
				Data: types.StringValue("invalid-ip"),
			}
			options := map[string]string{"ipAddress": "invalid-ip"}

			err := r.validateRecord(data, options)
			if err == nil {
				t.Error("Expected error for invalid A record, got nil")
			}
		})

		// Test MX record validation
		t.Run("MX Record Missing Priority", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type: types.StringValue("MX"),
				Data: types.StringValue("mail.example.com"),
				// Priority is missing
			}
			options := map[string]string{"exchange": "mail.example.com"}

			err := r.validateRecord(data, options)
			if err == nil {
				t.Error("Expected error for MX record without priority, got nil")
			}
		})

		t.Run("MX Record Valid", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type:     types.StringValue("MX"),
				Data:     types.StringValue("mail.example.com"),
				Priority: types.Int64Value(10),
			}
			options := map[string]string{
				"exchange":   "mail.example.com",
				"preference": "10",
			}

			err := r.validateRecord(data, options)
			if err != nil {
				t.Errorf("Expected no error for valid MX record, got: %v", err)
			}
		})

		// Test SRV record validation
		t.Run("SRV Record Missing Fields", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type:     types.StringValue("SRV"),
				Data:     types.StringValue("sip.example.com"),
				Priority: types.Int64Value(10),
				// Weight and Port are missing
			}
			options := map[string]string{
				"target":   "sip.example.com",
				"priority": "10",
			}

			err := r.validateRecord(data, options)
			if err == nil {
				t.Error("Expected error for SRV record with missing fields, got nil")
			}
		})

		t.Run("SRV Record Valid", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type:     types.StringValue("SRV"),
				Data:     types.StringValue("sip.example.com"),
				Priority: types.Int64Value(10),
				Weight:   types.Int64Value(5),
				Port:     types.Int64Value(5060),
			}
			options := map[string]string{
				"target":   "sip.example.com",
				"priority": "10",
				"weight":   "5",
				"port":     "5060",
			}

			err := r.validateRecord(data, options)
			if err != nil {
				t.Errorf("Expected no error for valid SRV record, got: %v", err)
			}
		})

		// Test FWD record validation
		t.Run("FWD Record Missing Forwarder", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type: types.StringValue("FWD"),
				// Both forwarder and data are missing
			}
			options := map[string]string{}

			err := r.validateRecord(data, options)
			if err == nil {
				t.Error("Expected error for FWD record without forwarder, got nil")
			}
		})

		t.Run("FWD Record Valid with Data Field", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type: types.StringValue("FWD"),
				Data: types.StringValue("8.8.8.8"),
			}
			options := map[string]string{"forwarder": "8.8.8.8"}

			err := r.validateRecord(data, options)
			if err != nil {
				t.Errorf("Expected no error for valid FWD record with data field, got: %v", err)
			}
		})

		t.Run("FWD Record Valid with Forwarder Field", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type:      types.StringValue("FWD"),
				Forwarder: types.StringValue("8.8.8.8"),
			}
			options := map[string]string{"forwarder": "8.8.8.8"}

			err := r.validateRecord(data, options)
			if err != nil {
				t.Errorf("Expected no error for valid FWD record with forwarder field, got: %v", err)
			}
		})

		t.Run("FWD Record Invalid Protocol", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type:     types.StringValue("FWD"),
				Data:     types.StringValue("8.8.8.8"),
				Protocol: types.StringValue("Invalid"),
			}
			options := map[string]string{"forwarder": "8.8.8.8"}

			err := r.validateRecord(data, options)
			if err == nil {
				t.Error("Expected error for FWD record with invalid protocol, got nil")
			}
		})

		t.Run("FWD Record Valid Protocol", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type:     types.StringValue("FWD"),
				Data:     types.StringValue("8.8.8.8"),
				Protocol: types.StringValue("Https"),
			}
			options := map[string]string{"forwarder": "8.8.8.8"}

			err := r.validateRecord(data, options)
			if err != nil {
				t.Errorf("Expected no error for valid FWD record with valid protocol, got: %v", err)
			}
		})

		t.Run("FWD Record Invalid Proxy Type", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type:      types.StringValue("FWD"),
				Data:      types.StringValue("8.8.8.8"),
				ProxyType: types.StringValue("InvalidProxy"),
			}
			options := map[string]string{"forwarder": "8.8.8.8"}

			err := r.validateRecord(data, options)
			if err == nil {
				t.Error("Expected error for FWD record with invalid proxy type, got nil")
			}
		})

		t.Run("FWD Record Proxy Missing Address", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type:      types.StringValue("FWD"),
				Data:      types.StringValue("8.8.8.8"),
				ProxyType: types.StringValue("Http"),
				// ProxyAddress is missing
			}
			options := map[string]string{"forwarder": "8.8.8.8"}

			err := r.validateRecord(data, options)
			if err == nil {
				t.Error("Expected error for FWD record with Http proxy type but missing address, got nil")
			}
		})

		t.Run("FWD Record Valid with Proxy", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type:         types.StringValue("FWD"),
				Data:         types.StringValue("8.8.8.8"),
				ProxyType:    types.StringValue("Http"),
				ProxyAddress: types.StringValue("proxy.example.com"),
				ProxyPort:    types.Int64Value(8080),
			}
			options := map[string]string{"forwarder": "8.8.8.8"}

			err := r.validateRecord(data, options)
			if err != nil {
				t.Errorf("Expected no error for valid FWD record with proxy, got: %v", err)
			}
		})
	})

	// Test the buildRecordOptions function
	t.Run("BuildRecordOptions", func(t *testing.T) {
		r := &DNSRecordResource{}
		ctx := context.Background()

		// Test A record
		t.Run("A Record Options", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type: types.StringValue("A"),
				Data: types.StringValue("192.168.1.1"),
			}

			options := r.buildRecordOptions(ctx, data, "create")
			if ip, ok := options["ipAddress"]; !ok || ip != "192.168.1.1" {
				t.Errorf("Expected ipAddress=192.168.1.1, got %v", options)
			}
		})

		// Test MX record
		t.Run("MX Record Options", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type:     types.StringValue("MX"),
				Data:     types.StringValue("mail.example.com"),
				Priority: types.Int64Value(10),
				Comments: types.StringValue("Mail server"),
			}

			options := r.buildRecordOptions(ctx, data, "create")
			if exchange, ok := options["exchange"]; !ok || exchange != "mail.example.com" {
				t.Errorf("Expected exchange=mail.example.com, got %v", options)
			}
			if preference, ok := options["preference"]; !ok || preference != "10" {
				t.Errorf("Expected preference=10, got %v", options)
			}
			if comments, ok := options["comments"]; !ok || comments != "Mail server" {
				t.Errorf("Expected comments='Mail server', got %v", options)
			}
		})

		// Test update operation (new values)
		t.Run("Update Options", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type: types.StringValue("A"),
				Data: types.StringValue("192.168.1.2"),
			}

			options := r.buildRecordOptions(ctx, data, "new")
			if ip, ok := options["newIpAddress"]; !ok || ip != "192.168.1.2" {
				t.Errorf("Expected newIpAddress=192.168.1.2, got %v", options)
			}
		})

		// Test FWD record options
		t.Run("FWD Record Options Basic", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type:     types.StringValue("FWD"),
				Data:     types.StringValue("8.8.8.8"),
				Protocol: types.StringValue("Https"),
			}

			options := r.buildRecordOptions(ctx, data, "create")
			if forwarder, ok := options["forwarder"]; !ok || forwarder != "8.8.8.8" {
				t.Errorf("Expected forwarder=8.8.8.8, got %v", options)
			}
			if protocol, ok := options["protocol"]; !ok || protocol != "Https" {
				t.Errorf("Expected protocol=Https, got %v", options)
			}
		})

		t.Run("FWD Record Options with Forwarder Field", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type:      types.StringValue("FWD"),
				Forwarder: types.StringValue("1.1.1.1"),
				Protocol:  types.StringValue("Tls"),
			}

			options := r.buildRecordOptions(ctx, data, "create")
			if forwarder, ok := options["forwarder"]; !ok || forwarder != "1.1.1.1" {
				t.Errorf("Expected forwarder=1.1.1.1, got %v", options)
			}
			if protocol, ok := options["protocol"]; !ok || protocol != "Tls" {
				t.Errorf("Expected protocol=Tls, got %v", options)
			}
		})

		t.Run("FWD Record Options Default Protocol", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type: types.StringValue("FWD"),
				Data: types.StringValue("8.8.8.8"),
				// No protocol specified
			}

			options := r.buildRecordOptions(ctx, data, "create")
			if protocol, ok := options["protocol"]; !ok || protocol != "Udp" {
				t.Errorf("Expected default protocol=Udp, got %v", options)
			}
		})

		t.Run("FWD Record Options with All Fields", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type:              types.StringValue("FWD"),
				Forwarder:         types.StringValue("8.8.8.8"),
				Protocol:          types.StringValue("Https"),
				ForwarderPriority: types.Int64Value(10),
				DnssecValidation:  types.BoolValue(true),
				ProxyType:         types.StringValue("Http"),
				ProxyAddress:      types.StringValue("proxy.example.com"),
				ProxyPort:         types.Int64Value(8080),
				ProxyUsername:     types.StringValue("user"),
				ProxyPassword:     types.StringValue("pass"),
			}

			options := r.buildRecordOptions(ctx, data, "create")

			expectedOptions := map[string]string{
				"forwarder":         "8.8.8.8",
				"protocol":          "Https",
				"forwarderPriority": "10",
				"dnssecValidation":  "true",
				"proxyType":         "Http",
				"proxyAddress":      "proxy.example.com",
				"proxyPort":         "8080",
				"proxyUsername":     "user",
				"proxyPassword":     "pass",
			}

			for key, expected := range expectedOptions {
				if actual, ok := options[key]; !ok || actual != expected {
					t.Errorf("Expected %s=%s, got %s=%s", key, expected, key, actual)
				}
			}
		})

		t.Run("FWD Record Update Options", func(t *testing.T) {
			data := &DNSRecordResourceModel{
				Type:      types.StringValue("FWD"),
				Forwarder: types.StringValue("9.9.9.9"),
				Protocol:  types.StringValue("Quic"),
			}

			options := r.buildRecordOptions(ctx, data, "new")
			if forwarder, ok := options["newForwarder"]; !ok || forwarder != "9.9.9.9" {
				t.Errorf("Expected newForwarder=9.9.9.9, got %v", options)
			}
			if protocol, ok := options["newProtocol"]; !ok || protocol != "Quic" {
				t.Errorf("Expected newProtocol=Quic, got %v", options)
			}
		})
	})
}
