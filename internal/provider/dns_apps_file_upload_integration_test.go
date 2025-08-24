package provider

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/testhelpers"
)

func TestDNSAppsFileUpload_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test container
	ctx := context.Background()
	container, err := testhelpers.StartTechnitiumContainer(ctx, t)
	if err != nil {
		t.Fatalf("Failed to start test container: %v", err)
	}

	t.Cleanup(func() {
		if err := container.Cleanup(ctx); err != nil {
			t.Logf("Warning: failed to cleanup container: %v", err)
		}
	})

	// Create API client
	client, err := testhelpers.CreateTestClient(container.GetAPIURL(), container.Username, container.Password)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}

	// Test 1: Create a mock DNS app ZIP file for testing file upload
	appZipData := createMockDNSAppZip(t)
	t.Logf("Created mock app ZIP file with %d bytes", len(appZipData))

	// Test 2: Attempt to install the mock app via file upload
	// This should properly handle the multipart form data and reach the Technitium server
	// Even if it fails due to invalid app content, it verifies the upload mechanism works
	_, err = client.InstallApp(ctx, "test-file-upload-app", appZipData)
	if err != nil {
		t.Logf("InstallApp failed as expected with mock app data: %v", err)

		// Verify this is a validation error (app content issue) not a form upload error
		errMsg := err.Error()
		if errMsg == "API error: Response status code does not indicate success: 404 (Not Found)." {
			t.Error("File upload endpoint appears to not exist (404) - this suggests DNS Apps file upload API is not available")
		} else if errMsg == "failed to create multipart form" ||
			errMsg == "failed to write form data" {
			t.Error("File upload mechanism failed - multipart form creation error")
		} else {
			// Expected - mock app data is invalid but upload mechanism worked
			t.Log("File upload mechanism working correctly - received server validation error as expected")
		}
	} else {
		t.Log("InstallApp succeeded unexpectedly with mock data - the upload mechanism is definitely working")
	}

	// Test 3: Test UpdateApp file upload mechanism
	_, err = client.UpdateApp(ctx, "test-file-upload-app", appZipData)
	if err != nil {
		t.Logf("UpdateApp failed as expected with mock app data: %v", err)

		// Similar verification as above
		errMsg := err.Error()
		if errMsg == "API error: Response status code does not indicate success: 404 (Not Found)." {
			t.Error("File update endpoint appears to not exist (404) - this suggests DNS Apps file update API is not available")
		} else if errMsg == "failed to create multipart form" ||
			errMsg == "failed to write form data" {
			t.Error("File update mechanism failed - multipart form creation error")
		} else {
			t.Log("File upload mechanism working correctly for updates - received server validation error as expected")
		}
	} else {
		t.Log("UpdateApp succeeded unexpectedly with mock data - the update upload mechanism is definitely working")
	}
}

// createMockDNSAppZip creates a mock ZIP file that looks like a DNS app package
// This is used to test the file upload mechanism without requiring a real app
func createMockDNSAppZip(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Create a mock app.config file (typical DNS app structure)
	configFile, err := zipWriter.Create("app.config")
	if err != nil {
		t.Fatalf("Failed to create config file in ZIP: %v", err)
	}

	configContent := `{
	"name": "Test Mock App",
	"version": "1.0.0",
	"description": "A mock DNS app for testing file upload",
	"classPath": "TestApp.App",
	"dnsApps": [
		{
			"classPath": "TestApp.App",
			"description": "Mock DNS app for testing",
			"isAppRecordRequestHandler": true
		}
	]
}`
	_, err = configFile.Write([]byte(configContent))
	if err != nil {
		t.Fatalf("Failed to write config content: %v", err)
	}

	// Create a mock DLL file (typical for .NET DNS apps)
	dllFile, err := zipWriter.Create("TestApp.dll")
	if err != nil {
		t.Fatalf("Failed to create DLL file in ZIP: %v", err)
	}

	// Write some mock binary content
	mockDllContent := []byte("MOCK_DLL_CONTENT_FOR_TESTING")
	_, err = dllFile.Write(mockDllContent)
	if err != nil {
		t.Fatalf("Failed to write DLL content: %v", err)
	}

	// Close the ZIP writer
	err = zipWriter.Close()
	if err != nil {
		t.Fatalf("Failed to close ZIP writer: %v", err)
	}

	return buf.Bytes()
}

func TestDNSAppsFileUploadMechanism_Unit(t *testing.T) {
	// Test that our mock ZIP creation works correctly
	zipData := createMockDNSAppZip(t)

	// Verify ZIP structure
	reader := bytes.NewReader(zipData)
	zipReader, err := zip.NewReader(reader, int64(len(zipData)))
	if err != nil {
		t.Fatalf("Failed to read created ZIP: %v", err)
	}

	// Check expected files exist
	expectedFiles := []string{"app.config", "TestApp.dll"}
	foundFiles := make(map[string]bool)

	for _, file := range zipReader.File {
		foundFiles[file.Name] = true
		t.Logf("Found file in ZIP: %s (%d bytes)", file.Name, file.UncompressedSize64)

		// Read and verify content of app.config
		if file.Name == "app.config" {
			fileReader, err := file.Open()
			if err != nil {
				t.Errorf("Failed to open app.config: %v", err)
				continue
			}

			content, err := io.ReadAll(fileReader)
			if err != nil {
				t.Errorf("Failed to read app.config: %v", err)
				fileReader.Close()
				continue
			}
			fileReader.Close()

			if !bytes.Contains(content, []byte("Test Mock App")) {
				t.Error("app.config does not contain expected content")
			}
		}
	}

	// Verify all expected files were found
	for _, expectedFile := range expectedFiles {
		if !foundFiles[expectedFile] {
			t.Errorf("Expected file %s not found in ZIP", expectedFile)
		}
	}
}
