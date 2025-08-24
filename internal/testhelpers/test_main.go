package testhelpers

import (
	"flag"
	"os"
	"testing"
)

// TestConfiguration holds test configuration flags
type TestConfiguration struct {
	AcceptanceTests bool
	ParallelTests   bool
	ContainerReuse  bool
	Verbose         bool
}

var testConfig TestConfiguration

func init() {
	flag.BoolVar(&testConfig.AcceptanceTests, "acceptance", false, "Run acceptance tests")
	flag.BoolVar(&testConfig.ParallelTests, "parallel", true, "Run tests in parallel")
	flag.BoolVar(&testConfig.ContainerReuse, "container-reuse", false, "Reuse containers between tests")
	flag.BoolVar(&testConfig.Verbose, "verbose", false, "Verbose test output")
}

// GetTestConfig returns the current test configuration
func GetTestConfig() TestConfiguration {
	return testConfig
}

// ShouldRunAcceptanceTests returns true if acceptance tests should run
func ShouldRunAcceptanceTests() bool {
	return os.Getenv("TF_ACC") != "" || testConfig.AcceptanceTests
}

// ShouldRunInParallel returns true if tests should run in parallel
func ShouldRunInParallel() bool {
	return testConfig.ParallelTests && os.Getenv("NO_PARALLEL") == ""
}

// SetupTestEnvironment prepares the test environment
func SetupTestEnvironment(t *testing.T) {
	t.Helper()

	if testConfig.Verbose {
		t.Logf("Test environment configuration:")
		t.Logf("  Acceptance tests: %v", ShouldRunAcceptanceTests())
		t.Logf("  Parallel tests: %v", ShouldRunInParallel())
		t.Logf("  Container reuse: %v", testConfig.ContainerReuse)
	}

	if ShouldRunInParallel() {
		t.Parallel()
	}
}

// SkipIfNotAcceptance skips the test if acceptance tests are not enabled
func SkipIfNotAcceptance(t *testing.T) {
	t.Helper()
	if !ShouldRunAcceptanceTests() {
		t.Skip("Acceptance tests skipped (set TF_ACC=1 to enable)")
	}
}
