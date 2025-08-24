package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

func TestProvider(t *testing.T) {
	t.Parallel()

	// Test that the provider factory works
	provider := New("test")()
	if provider == nil {
		t.Fatal("Provider should not be nil")
	}

	// Test that we can create a provider server
	server := providerserver.NewProtocol6(provider)
	if server == nil {
		t.Fatal("Provider server should not be nil")
	}
}

// ProviderServerFactory is used for acceptance testing
func ProviderServerFactory() func() tfprotov6.ProviderServer {
	return providerserver.NewProtocol6(New("test")())
}
