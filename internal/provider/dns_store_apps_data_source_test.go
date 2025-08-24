package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestDNSStoreAppsDataSource(t *testing.T) {
	t.Parallel()

	// Unit test - verify data source creation
	t.Run("NewDNSStoreAppsDataSource", func(t *testing.T) {
		ds := NewDNSStoreAppsDataSource()
		if ds == nil {
			t.Fatal("NewDNSStoreAppsDataSource should return a non-nil data source")
		}

		// Test metadata
		var resp datasource.MetadataResponse
		ds.Metadata(context.Background(), datasource.MetadataRequest{
			ProviderTypeName: "technitium",
		}, &resp)

		if resp.TypeName != "technitium_dns_store_apps" {
			t.Errorf("Expected TypeName to be technitium_dns_store_apps, got %s", resp.TypeName)
		}
	})

	// Unit test - verify schema
	t.Run("Schema", func(t *testing.T) {
		ds := NewDNSStoreAppsDataSource()
		var resp datasource.SchemaResponse
		ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Schema validation failed: %v", resp.Diagnostics.Errors())
		}

		// Verify required attributes
		schema := resp.Schema
		if _, ok := schema.Attributes["id"]; !ok {
			t.Error("Schema should have 'id' attribute")
		}
		if _, ok := schema.Attributes["store_apps"]; !ok {
			t.Error("Schema should have 'store_apps' attribute")
		}

		idAttr := schema.Attributes["id"]
		if !idAttr.IsComputed() {
			t.Error("ID attribute should be computed")
		}

		storeAppsAttr := schema.Attributes["store_apps"]
		if !storeAppsAttr.IsComputed() {
			t.Error("Store apps attribute should be computed")
		}
	})

	// Unit test - verify configure method
	t.Run("Configure", func(t *testing.T) {
		ds := NewDNSStoreAppsDataSource().(*DNSStoreAppsDataSource)

		// Test with nil provider data
		var resp datasource.ConfigureResponse
		ds.Configure(context.Background(), datasource.ConfigureRequest{
			ProviderData: nil,
		}, &resp)

		if resp.Diagnostics.HasError() {
			t.Errorf("Configure should not fail with nil provider data: %v", resp.Diagnostics.Errors())
		}

		// Test with wrong provider data type
		resp = datasource.ConfigureResponse{}
		ds.Configure(context.Background(), datasource.ConfigureRequest{
			ProviderData: "wrong-type",
		}, &resp)

		if !resp.Diagnostics.HasError() {
			t.Error("Configure should fail with wrong provider data type")
		}
	})
}

func TestDNSStoreAppsDataSource_SchemaValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		attributeName string
		shouldExist   bool
		isComputed    bool
	}{
		{
			name:          "id attribute",
			attributeName: "id",
			shouldExist:   true,
			isComputed:    true,
		},
		{
			name:          "store_apps attribute",
			attributeName: "store_apps",
			shouldExist:   true,
			isComputed:    true,
		},
	}

	ds := NewDNSStoreAppsDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema validation failed: %v", resp.Diagnostics.Errors())
	}

	schema := resp.Schema

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			attr, exists := schema.Attributes[test.attributeName]

			if test.shouldExist && !exists {
				t.Errorf("Attribute %s should exist", test.attributeName)
				return
			}

			if !test.shouldExist && exists {
				t.Errorf("Attribute %s should not exist", test.attributeName)
				return
			}

			if !exists {
				return // Skip further checks if attribute doesn't exist
			}

			if test.isComputed && !attr.IsComputed() {
				t.Errorf("Attribute %s should be computed", test.attributeName)
			}

			// Data source attributes should never be required
			if attr.IsRequired() {
				t.Errorf("Data source attribute %s should not be required", test.attributeName)
			}
		})
	}
}
