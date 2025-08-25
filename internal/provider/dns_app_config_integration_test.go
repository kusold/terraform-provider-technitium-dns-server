package provider

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/client"
	"github.com/kusold/terraform-provider-technitium-dns-server/internal/testhelpers"
)

func TestDNSAppConfig_SplitHorizon_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Parallel()

	ctx := context.Background()

	// Start Technitium container
	container, err := testhelpers.StartTechnitiumContainer(ctx, t)
	if err != nil {
		t.Fatalf("Failed to setup container: %v", err)
	}
	defer func() {
		if err := container.Cleanup(ctx); err != nil {
			t.Logf("Failed to cleanup container: %v", err)
		}
	}()

	// Create API client
	apiClient, err := testhelpers.CreateTestClient(container.GetAPIURL(), container.Username, container.Password)
	if err != nil {
		t.Fatalf("Failed to create API client: %v", err)
	}

	// Authenticate
	if err := apiClient.Authenticate(ctx); err != nil {
		t.Fatalf("Failed to authenticate: %v", err)
	}

	// Test the complete Split Horizon app lifecycle:
	// 1. List store apps to find Split Horizon
	// 2. Install Split Horizon app
	// 3. Verify installation
	// 4. Set configuration
	// 5. Verify configuration was set
	// 6. Update configuration
	// 7. Verify configuration was updated
	// 8. Clear configuration
	// 9. Verify configuration was cleared

	t.Run("ListStoreApps", func(t *testing.T) {
		storeApps, err := apiClient.ListStoreApps(ctx)
		if err != nil {
			t.Fatalf("Failed to list store apps: %v", err)
		}

		t.Logf("Found %d store apps", len(storeApps))

		// Find Split Horizon app
		var splitHorizonApp *client.StoreApp
		for _, app := range storeApps {
			if app.Name == "Split Horizon" {
				splitHorizonApp = &app
				break
			}
		}

		if splitHorizonApp == nil {
			t.Fatal("Split Horizon app not found in store")
		}

		t.Logf("Found Split Horizon app: version=%s, url=%s, installed=%v",
			splitHorizonApp.Version, splitHorizonApp.URL, splitHorizonApp.Installed)

		// Install the app if not already installed
		if !splitHorizonApp.Installed {
			t.Log("Installing Split Horizon app...")
			_, err := apiClient.DownloadAndInstallApp(ctx, "Split Horizon", splitHorizonApp.URL)
			if err != nil {
				t.Fatalf("Failed to install Split Horizon app: %v", err)
			}
			t.Log("Split Horizon app installed successfully")
		} else {
			t.Log("Split Horizon app already installed")
		}
	})

	t.Run("VerifyInstallation", func(t *testing.T) {
		// Verify the app is now installed
		apps, err := apiClient.ListApps(ctx)
		if err != nil {
			t.Fatalf("Failed to list installed apps: %v", err)
		}

		var splitHorizonApp *client.App
		for _, app := range apps {
			if app.Name == "Split Horizon" {
				splitHorizonApp = &app
				break
			}
		}

		if splitHorizonApp == nil {
			t.Fatal("Split Horizon app not found in installed apps")
		}

		t.Logf("Split Horizon app installed: version=%s, dns_apps_count=%d",
			splitHorizonApp.Version, len(splitHorizonApp.DNSApps))

		if len(splitHorizonApp.DNSApps) == 0 {
			t.Error("Split Horizon app should have at least one DNS app component")
		}
	})

	t.Run("SetInitialConfiguration", func(t *testing.T) {
		// Get the current configuration to understand the structure
		currentConfig, err := apiClient.GetAppConfig(ctx, "Split Horizon")
		if err != nil {
			t.Fatalf("Failed to get current app config: %v", err)
		}

		// Parse the current configuration
		var config map[string]interface{}
		if currentConfig != nil {
			err = json.Unmarshal([]byte(*currentConfig), &config)
			if err != nil {
				t.Fatalf("Failed to parse current config: %v", err)
			}
		} else {
			config = make(map[string]interface{})
		}

		// Add a simple modification - enable the app and add a custom network
		config["enabled"] = true
		config["testField"] = "integration-test"

		// If networks exist, modify them safely
		if networks, ok := config["networks"].(map[string]interface{}); ok {
			if customNetworks, exists := networks["custom-networks"].([]interface{}); exists {
				// Add a test network
				customNetworks = append(customNetworks, "203.0.113.0/24")
				networks["custom-networks"] = customNetworks
			}
		}

		configJSON, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal config: %v", err)
		}

		t.Logf("Setting Split Horizon configuration: %s", string(configJSON))

		err = apiClient.SetAppConfig(ctx, "Split Horizon", string(configJSON))
		if err != nil {
			t.Fatalf("Failed to set app config: %v", err)
		}

		t.Log("Split Horizon configuration set successfully")
	})

	t.Run("VerifyConfiguration", func(t *testing.T) {
		// Verify the configuration was set correctly
		config, err := apiClient.GetAppConfig(ctx, "Split Horizon")
		if err != nil {
			t.Fatalf("Failed to get app config: %v", err)
		}

		if config == nil {
			t.Fatal("Configuration should not be nil")
		}

		t.Logf("Retrieved configuration: %s", *config)

		// Parse and validate the configuration
		var parsedConfig map[string]interface{}
		err = json.Unmarshal([]byte(*config), &parsedConfig)
		if err != nil {
			t.Fatalf("Failed to parse configuration JSON: %v", err)
		}

		// Check for our test modifications
		if enabled, ok := parsedConfig["enabled"].(bool); !ok || !enabled {
			t.Error("Configuration should contain 'enabled' field set to true")
		}

		if testField, ok := parsedConfig["testField"].(string); !ok || testField != "integration-test" {
			t.Error("Configuration should contain 'testField' set to 'integration-test'")
		}

		// Check if networks were modified
		if networks, ok := parsedConfig["networks"].(map[string]interface{}); ok {
			if customNetworks, exists := networks["custom-networks"].([]interface{}); exists {
				found := false
				for _, network := range customNetworks {
					if networkStr, ok := network.(string); ok && networkStr == "203.0.113.0/24" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Custom network '203.0.113.0/24' should be present")
				}
			}
		}

		t.Log("Configuration verified successfully")
	})

	t.Run("UpdateConfiguration", func(t *testing.T) {
		// Get the current configuration
		currentConfig, err := apiClient.GetAppConfig(ctx, "Split Horizon")
		if err != nil {
			t.Fatalf("Failed to get current config for update: %v", err)
		}

		// Parse the current configuration
		var config map[string]interface{}
		err = json.Unmarshal([]byte(*currentConfig), &config)
		if err != nil {
			t.Fatalf("Failed to parse current config for update: %v", err)
		}

		// Update fields
		config["enabled"] = true
		config["testField"] = "updated-integration-test"
		config["updateCounter"] = 42

		// If networks exist, modify them safely
		if networks, ok := config["networks"].(map[string]interface{}); ok {
			if customNetworks, exists := networks["custom-networks"].([]interface{}); exists {
				// Add another test network
				customNetworks = append(customNetworks, "198.51.100.0/24")
				networks["custom-networks"] = customNetworks
			}
		}

		configJSON, err := json.Marshal(config)
		if err != nil {
			t.Fatalf("Failed to marshal updated config: %v", err)
		}

		t.Logf("Updating Split Horizon configuration: %s", string(configJSON))

		err = apiClient.SetAppConfig(ctx, "Split Horizon", string(configJSON))
		if err != nil {
			t.Fatalf("Failed to update app config: %v", err)
		}

		t.Log("Split Horizon configuration updated successfully")
	})

	t.Run("VerifyUpdatedConfiguration", func(t *testing.T) {
		// Add a small delay to ensure configuration is persisted
		time.Sleep(100 * time.Millisecond)

		config, err := apiClient.GetAppConfig(ctx, "Split Horizon")
		if err != nil {
			t.Fatalf("Failed to get updated app config: %v", err)
		}

		if config == nil {
			t.Fatal("Updated configuration should not be nil")
		}

		t.Logf("Retrieved updated configuration: %s", *config)

		// Parse and validate the updated configuration
		var parsedConfig map[string]interface{}
		err = json.Unmarshal([]byte(*config), &parsedConfig)
		if err != nil {
			t.Fatalf("Failed to parse updated configuration JSON: %v", err)
		}

		// Check for the updated fields
		if enabled, ok := parsedConfig["enabled"].(bool); !ok || !enabled {
			t.Error("Updated configuration should contain 'enabled' field set to true")
		}

		if testField, ok := parsedConfig["testField"].(string); !ok || testField != "updated-integration-test" {
			t.Error("Updated configuration should contain 'testField' set to 'updated-integration-test'")
		}

		if updateCounter, ok := parsedConfig["updateCounter"].(float64); !ok || updateCounter != 42 {
			t.Error("Updated configuration should contain 'updateCounter' set to 42")
		}

		// Check if networks were updated
		if networks, ok := parsedConfig["networks"].(map[string]interface{}); ok {
			if customNetworks, exists := networks["custom-networks"].([]interface{}); exists {
				found := false
				for _, network := range customNetworks {
					if networkStr, ok := network.(string); ok && networkStr == "198.51.100.0/24" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Updated custom network '198.51.100.0/24' should be present")
				}
			}
		}

		t.Log("Updated configuration verified successfully")
	})

	t.Run("ClearConfiguration", func(t *testing.T) {
		// Clear the configuration by setting it to empty string
		err := apiClient.SetAppConfig(ctx, "Split Horizon", "")
		if err != nil {
			t.Fatalf("Failed to clear app config: %v", err)
		}

		t.Log("Split Horizon configuration cleared successfully")
	})

	t.Run("VerifyConfigurationCleared", func(t *testing.T) {
		// Add a small delay to ensure configuration is persisted
		time.Sleep(100 * time.Millisecond)

		config, err := apiClient.GetAppConfig(ctx, "Split Horizon")
		if err != nil {
			t.Fatalf("Failed to get app config after clearing: %v", err)
		}

		// Configuration should be either nil or empty string
		if config != nil && *config != "" {
			t.Errorf("Configuration should be empty after clearing, got: %s", *config)
		}

		t.Log("Configuration successfully cleared")
	})

	t.Run("TestErrorCases", func(t *testing.T) {
		// Test setting configuration for non-existent app
		err := apiClient.SetAppConfig(ctx, "Non Existent App", `{"test": true}`)
		if err == nil {
			t.Error("Setting config for non-existent app should fail")
		}
		t.Logf("Expected error for non-existent app: %v", err)

		// Test getting configuration for non-existent app
		_, err = apiClient.GetAppConfig(ctx, "Non Existent App")
		if err == nil {
			t.Error("Getting config for non-existent app should fail")
		}
		t.Logf("Expected error for non-existent app: %v", err)

		// Test setting invalid JSON (this might be accepted by the server)
		err = apiClient.SetAppConfig(ctx, "Split Horizon", "invalid json")
		t.Logf("Setting invalid JSON config result: %v", err)
	})
}
