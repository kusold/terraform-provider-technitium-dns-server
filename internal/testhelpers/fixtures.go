package testhelpers

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"time"
)

// TechnitiumTestZone represents a test DNS zone configuration
type TechnitiumTestZone struct {
	Name        string
	Type        string
	Description string
}

// TechnitiumTestRecord represents a test DNS record configuration
type TechnitiumTestRecord struct {
	Zone     string
	Name     string
	Type     string
	TTL      int
	Data     string
	Priority int
}

// Common test zones
var (
	TestZonePrimary = TechnitiumTestZone{
		Name:        "test.local",
		Type:        "Primary",
		Description: "Test primary zone for integration testing",
	}

	TestZoneSecondary = TechnitiumTestZone{
		Name:        "secondary.local",
		Type:        "Secondary",
		Description: "Test secondary zone for integration testing",
	}
)

// Common test DNS records
var (
	TestRecordA = TechnitiumTestRecord{
		Zone: "test.local",
		Name: "www",
		Type: "A",
		TTL:  300,
		Data: "192.168.1.100",
	}

	TestRecordAAAA = TechnitiumTestRecord{
		Zone: "test.local",
		Name: "ipv6",
		Type: "AAAA",
		TTL:  300,
		Data: "2001:db8::1",
	}

	TestRecordCNAME = TechnitiumTestRecord{
		Zone: "test.local",
		Name: "alias",
		Type: "CNAME",
		TTL:  300,
		Data: "www.test.local",
	}

	TestRecordMX = TechnitiumTestRecord{
		Zone:     "test.local",
		Name:     "mail",
		Type:     "MX",
		TTL:      300,
		Data:     "mail.test.local",
		Priority: 10,
	}

	TestRecordTXT = TechnitiumTestRecord{
		Zone: "test.local",
		Name: "txt",
		Type: "TXT",
		TTL:  300,
		Data: "v=spf1 include:_spf.google.com ~all",
	}
)

// GenerateTestZoneName generates a unique test zone name
func GenerateTestZoneName(prefix string) string {
	return fmt.Sprintf("%s-%d.test", prefix, time.Now().Unix())
}

// GenerateTestRecordName generates a unique test record name
func GenerateTestRecordName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().Unix())
}

// GetTestProviderConfig returns a test provider configuration
func GetTestProviderConfig(host, username, password string) string {
	return fmt.Sprintf(`
provider "technitium" {
  host     = "%s"
  username = "%s"
  password = "%s"
}
`, host, username, password)
}

// GetTestZoneResource returns a test zone resource configuration
func GetTestZoneResource(name, zoneName, zoneType string) string {
	return fmt.Sprintf(`
resource "technitium_zone" "%s" {
  name = "%s"
  type = "%s"
}
`, name, zoneName, zoneType)
}

// GetTestRecordResource returns a test DNS record resource configuration
func GetTestRecordResource(name string, record TechnitiumTestRecord) string {
	config := fmt.Sprintf(`
resource "technitium_dns_record" "%s" {
  zone = "%s"
  name = "%s"
  type = "%s"
  ttl  = %d
  data = "%s"
`, name, record.Zone, record.Name, record.Type, record.TTL, record.Data)

	if record.Priority > 0 {
		config += fmt.Sprintf("  priority = %d\n", record.Priority)
	}

	config += "}\n"
	return config
}

// GetTestDataSource returns a test data source configuration
func GetTestDataSource(name, zoneName string) string {
	return fmt.Sprintf(`
data "technitium_zone" "%s" {
  name = "%s"
}
`, name, zoneName)
}

// GetCompleteTestConfig returns a complete test configuration with provider, zone, and records
func GetCompleteTestConfig(host, username, password, zoneName string) string {
	return fmt.Sprintf(`
%s

%s

%s

%s

%s
`,
		GetTestProviderConfig(host, username, password),
		GetTestZoneResource("test", zoneName, "Primary"),
		GetTestRecordResource("a_record", TechnitiumTestRecord{
			Zone: zoneName,
			Name: "www",
			Type: "A",
			TTL:  300,
			Data: "192.168.1.100",
		}),
		GetTestRecordResource("cname_record", TechnitiumTestRecord{
			Zone: zoneName,
			Name: "alias",
			Type: "CNAME",
			TTL:  300,
			Data: fmt.Sprintf("www.%s", zoneName),
		}),
		GetTestDataSource("test_zone", zoneName),
	)
}

// CreateMockDNSAppZip creates a mock DNS app ZIP file for testing
func CreateMockDNSAppZip(appName, version string) (string, error) {
	// Create temporary file
	tempFile, err := os.CreateTemp("", fmt.Sprintf("%s-*.zip", appName))
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Create ZIP writer
	zipWriter := zip.NewWriter(tempFile)
	defer zipWriter.Close()

	// Create dnsApp.config file (required for DNS apps)
	configContent := fmt.Sprintf(`{
  "displayName": "%s Test App",
  "version": "%s",
  "description": "Test DNS application for integration testing",
  "applicationRecordDataTemplate": "127.0.0.1",
  "author": "Test"
}`, appName, version)

	configFile, err := zipWriter.Create("dnsApp.config")
	if err != nil {
		return "", fmt.Errorf("failed to create config file: %w", err)
	}
	_, err = configFile.Write([]byte(configContent))
	if err != nil {
		return "", fmt.Errorf("failed to write config content: %w", err)
	}

	// Create a basic DLL file (empty but valid structure)
	dllContent := []byte("MZ") // Basic PE header signature
	dllFile, err := zipWriter.Create(fmt.Sprintf("%s.dll", appName))
	if err != nil {
		return "", fmt.Errorf("failed to create DLL file: %w", err)
	}
	_, err = dllFile.Write(dllContent)
	if err != nil {
		return "", fmt.Errorf("failed to write DLL content: %w", err)
	}

	return tempFile.Name(), nil
}

// CreateMockDNSAppZipBase64 creates a mock DNS app ZIP file and returns it as base64 encoded string
func CreateMockDNSAppZipBase64(appName, version string) (string, error) {
	// Create temporary file
	tempFile, err := os.CreateTemp("", fmt.Sprintf("%s-*.zip", appName))
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	// Create ZIP writer
	zipWriter := zip.NewWriter(tempFile)

	// Create dnsApp.config file (required for DNS apps)
	configContent := fmt.Sprintf(`{
  "displayName": "%s Test App",
  "version": "%s",
  "description": "Test DNS application for integration testing",
  "applicationRecordDataTemplate": "127.0.0.1",
  "author": "Test"
}`, appName, version)

	configFile, err := zipWriter.Create("dnsApp.config")
	if err != nil {
		return "", fmt.Errorf("failed to create config file: %w", err)
	}
	_, err = configFile.Write([]byte(configContent))
	if err != nil {
		return "", fmt.Errorf("failed to write config content: %w", err)
	}

	// Create a basic DLL file (empty but valid structure)
	dllContent := []byte("MZ") // Basic PE header signature
	dllFile, err := zipWriter.Create(fmt.Sprintf("%s.dll", appName))
	if err != nil {
		return "", fmt.Errorf("failed to create DLL file: %w", err)
	}
	_, err = dllFile.Write(dllContent)
	if err != nil {
		return "", fmt.Errorf("failed to write DLL content: %w", err)
	}

	// Close the zip writer to finalize the file
	err = zipWriter.Close()
	if err != nil {
		return "", fmt.Errorf("failed to close zip writer: %w", err)
	}

	// Read the file and encode as base64
	_, err = tempFile.Seek(0, 0)
	if err != nil {
		return "", fmt.Errorf("failed to seek to beginning of file: %w", err)
	}

	content, err := io.ReadAll(tempFile)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	return base64.StdEncoding.EncodeToString(content), nil
}

// CleanupMockZipFile removes the temporary ZIP file
func CleanupMockZipFile(filePath string) error {
	return os.Remove(filePath)
}
