package main

import (
	"os"
	"testing"
)

// TestMain sets up and tears down test infrastructure
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// TODO: Cleanup all test containers when API client is ready
	// testhelpers.CleanupAllTestContainers()

	// Exit with the test result code
	os.Exit(code)
}