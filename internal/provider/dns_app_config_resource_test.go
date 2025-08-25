package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestDNSAppConfigResource(t *testing.T) {
	t.Parallel()

	// Unit test - verify resource creation
	t.Run("NewDNSAppConfigResource", func(t *testing.T) {
		r := NewDNSAppConfigResource()
		if r == nil {
			t.Fatal("NewDNSAppConfigResource should return a non-nil resource")
		}

		// Test metadata
		var resp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{
			ProviderTypeName: "technitium",
		}, &resp)

		if resp.TypeName != "technitium_dns_app_config" {
			t.Errorf("Expected TypeName to be technitium_dns_app_config, got %s", resp.TypeName)
		}
	})

	// Unit test - verify schema
	t.Run("Schema", func(t *testing.T) {
		r := NewDNSAppConfigResource()
		var resp resource.SchemaResponse
		r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Schema validation failed: %v", resp.Diagnostics.Errors())
		}

		// Verify required attributes
		schema := resp.Schema
		if _, ok := schema.Attributes["name"]; !ok {
			t.Error("Schema should have 'name' attribute")
		}
		if _, ok := schema.Attributes["config"]; !ok {
			t.Error("Schema should have 'config' attribute")
		}

		// Verify computed attributes
		if _, ok := schema.Attributes["id"]; !ok {
			t.Error("Schema should have 'id' attribute")
		}
	})

	// Unit test - verify configure method
	t.Run("Configure", func(t *testing.T) {
		r := NewDNSAppConfigResource().(*DNSAppConfigResource)
		var resp resource.ConfigureResponse

		// Test with nil provider data
		r.Configure(context.Background(), resource.ConfigureRequest{
			ProviderData: nil,
		}, &resp)

		if resp.Diagnostics.HasError() {
			t.Error("Configure should not error with nil provider data")
		}

		// Test with wrong type
		r.Configure(context.Background(), resource.ConfigureRequest{
			ProviderData: "wrong type",
		}, &resp)

		if !resp.Diagnostics.HasError() {
			t.Error("Configure should error with wrong provider data type")
		}
	})
}
