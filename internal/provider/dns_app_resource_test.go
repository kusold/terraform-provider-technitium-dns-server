package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestDNSAppResource(t *testing.T) {
	t.Parallel()

	// Unit test - verify resource creation
	t.Run("NewDNSAppResource", func(t *testing.T) {
		r := NewDNSAppResource()
		if r == nil {
			t.Fatal("NewDNSAppResource should return a non-nil resource")
		}

		// Test metadata
		var resp resource.MetadataResponse
		r.Metadata(context.Background(), resource.MetadataRequest{
			ProviderTypeName: "technitium",
		}, &resp)

		if resp.TypeName != "technitium_dns_app" {
			t.Errorf("Expected TypeName to be technitium_dns_app, got %s", resp.TypeName)
		}
	})

	// Unit test - verify schema
	t.Run("Schema", func(t *testing.T) {
		r := NewDNSAppResource()
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
		if _, ok := schema.Attributes["install_method"]; !ok {
			t.Error("Schema should have 'install_method' attribute")
		}

		// Verify computed attributes
		if _, ok := schema.Attributes["version"]; !ok {
			t.Error("Schema should have 'version' computed attribute")
		}
		if _, ok := schema.Attributes["dns_apps"]; !ok {
			t.Error("Schema should have 'dns_apps' computed attribute")
		}

		// Verify optional attributes
		if _, ok := schema.Attributes["url"]; !ok {
			t.Error("Schema should have 'url' attribute")
		}
		if _, ok := schema.Attributes["file_content"]; !ok {
			t.Error("Schema should have 'file_content' attribute")
		}

		nameAttr := schema.Attributes["name"]
		if !nameAttr.IsRequired() {
			t.Error("Name attribute should be required")
		}

		installMethodAttr := schema.Attributes["install_method"]
		if !installMethodAttr.IsRequired() {
			t.Error("Install method attribute should be required")
		}

		urlAttr := schema.Attributes["url"]
		if !urlAttr.IsOptional() {
			t.Error("URL attribute should be optional")
		}

		versionAttr := schema.Attributes["version"]
		if !versionAttr.IsComputed() {
			t.Error("Version attribute should be computed")
		}
	})

	// Unit test - verify configure method
	t.Run("Configure", func(t *testing.T) {
		r := NewDNSAppResource().(*DNSAppResource)

		// Test with nil provider data
		var resp resource.ConfigureResponse
		r.Configure(context.Background(), resource.ConfigureRequest{
			ProviderData: nil,
		}, &resp)

		if resp.Diagnostics.HasError() {
			t.Errorf("Configure should not fail with nil provider data: %v", resp.Diagnostics.Errors())
		}

		// Test with wrong provider data type
		resp = resource.ConfigureResponse{}
		r.Configure(context.Background(), resource.ConfigureRequest{
			ProviderData: "wrong-type",
		}, &resp)

		if !resp.Diagnostics.HasError() {
			t.Error("Configure should fail with wrong provider data type")
		}
	})
}

func TestDNSAppResource_SchemaValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		attributeName string
		shouldExist   bool
		isRequired    bool
		isOptional    bool
		isComputed    bool
	}{
		{
			name:          "name attribute",
			attributeName: "name",
			shouldExist:   true,
			isRequired:    true,
			isOptional:    false,
			isComputed:    false,
		},
		{
			name:          "install_method attribute",
			attributeName: "install_method",
			shouldExist:   true,
			isRequired:    true,
			isOptional:    false,
			isComputed:    false,
		},
		{
			name:          "url attribute",
			attributeName: "url",
			shouldExist:   true,
			isRequired:    false,
			isOptional:    true,
			isComputed:    false,
		},
		{
			name:          "file_content attribute",
			attributeName: "file_content",
			shouldExist:   true,
			isRequired:    false,
			isOptional:    true,
			isComputed:    false,
		},
		{
			name:          "version attribute",
			attributeName: "version",
			shouldExist:   true,
			isRequired:    false,
			isOptional:    false,
			isComputed:    true,
		},
		{
			name:          "dns_apps attribute",
			attributeName: "dns_apps",
			shouldExist:   true,
			isRequired:    false,
			isOptional:    false,
			isComputed:    true,
		},
	}

	r := NewDNSAppResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

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

			if test.isRequired && !attr.IsRequired() {
				t.Errorf("Attribute %s should be required", test.attributeName)
			}

			if test.isOptional && !attr.IsOptional() {
				t.Errorf("Attribute %s should be optional", test.attributeName)
			}

			if test.isComputed && !attr.IsComputed() {
				t.Errorf("Attribute %s should be computed", test.attributeName)
			}

			// Check that required attributes are not optional or computed
			if test.isRequired && (attr.IsOptional() || attr.IsComputed()) {
				t.Errorf("Attribute %s should be required only (not optional or computed)", test.attributeName)
			}

			// Check that computed attributes are not required
			if test.isComputed && attr.IsRequired() {
				t.Errorf("Attribute %s should not be both computed and required", test.attributeName)
			}
		})
	}
}
