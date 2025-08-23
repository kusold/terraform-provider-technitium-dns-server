package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestZoneResource(t *testing.T) {
	t.Parallel()

	// Unit test - verify resource creation
	t.Run("NewZoneResource", func(t *testing.T) {
		r := NewZoneResource()
		if r == nil {
			t.Fatal("NewZoneResource should return a non-nil resource")
		}

		// Test metadata
		var resp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{
			ProviderTypeName: "technitium",
		}, &resp)

		if resp.TypeName != "technitium_zone" {
			t.Errorf("Expected TypeName to be technitium_zone, got %s", resp.TypeName)
		}
	})

	// Unit test - verify schema
	t.Run("Schema", func(t *testing.T) {
		r := NewZoneResource()
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
		if _, ok := schema.Attributes["type"]; !ok {
			t.Error("Schema should have 'type' attribute")
		}

		// Verify computed attributes
		if attr, ok := schema.Attributes["internal"]; ok {
			if !attr.IsComputed() {
				t.Error("'internal' attribute should be computed")
			}
		} else {
			t.Error("Schema should have 'internal' attribute")
		}

		if attr, ok := schema.Attributes["dnssec_status"]; ok {
			if !attr.IsComputed() {
				t.Error("'dnssec_status' attribute should be computed")
			}
		} else {
			t.Error("Schema should have 'dnssec_status' attribute")
		}
	})
}