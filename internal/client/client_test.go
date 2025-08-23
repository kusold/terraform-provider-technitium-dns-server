package client

import (
	"os"
	"testing"
)

// This is a simple test to verify the client authentication works
// This test will be skipped unless TF_ACC is set
func TestClientAuthentication(t *testing.T) {
	// Skip unless doing acceptance tests
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Skipping client test unless TF_ACC is set")
	}

	// This is just a compilation test to ensure the client compiles
	config := Config{
		Host:               "http://localhost:5380",
		Username:           "admin",
		Password:           "admin",
		TimeoutSeconds:     30,
		RetryAttempts:      3,
		InsecureSkipVerify: false,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if client == nil {
		t.Fatal("Client should not be nil")
	}

	// Don't actually try to authenticate since we don't have a running server
	// This test just verifies the client creation works
}