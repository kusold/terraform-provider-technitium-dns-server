package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestListApps(t *testing.T) {
	// Create mock response
	mockResponse := APIResponse{
		Status: "ok",
		Response: json.RawMessage(`{
			"apps": [
				{
					"name": "Test App",
					"version": "1.0",
					"dnsApps": [
						{
							"classPath": "TestApp.App",
							"description": "Test app description",
							"isAppRecordRequestHandler": true,
							"recordDataTemplate": null,
							"isRequestController": false,
							"isAuthoritativeRequestHandler": false,
							"isRequestBlockingHandler": false,
							"isQueryLogger": false,
							"isPostProcessor": false
						}
					]
				}
			]
		}`),
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/apps/list" {
			t.Errorf("Expected path /api/apps/list, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
		Token:      "test-token",
		retries:    1,
	}

	// Test ListApps
	apps, err := client.ListApps(context.Background())
	if err != nil {
		t.Fatalf("ListApps failed: %v", err)
	}

	if len(apps) != 1 {
		t.Fatalf("Expected 1 app, got %d", len(apps))
	}

	app := apps[0]
	if app.Name != "Test App" {
		t.Errorf("Expected app name 'Test App', got '%s'", app.Name)
	}
	if app.Version != "1.0" {
		t.Errorf("Expected app version '1.0', got '%s'", app.Version)
	}
	if len(app.DNSApps) != 1 {
		t.Fatalf("Expected 1 DNS app, got %d", len(app.DNSApps))
	}

	dnsApp := app.DNSApps[0]
	if dnsApp.ClassPath != "TestApp.App" {
		t.Errorf("Expected class path 'TestApp.App', got '%s'", dnsApp.ClassPath)
	}
	if !dnsApp.IsAppRecordRequestHandler {
		t.Error("Expected IsAppRecordRequestHandler to be true")
	}
}

func TestListStoreApps(t *testing.T) {
	// Create mock response
	mockResponse := APIResponse{
		Status: "ok",
		Response: json.RawMessage(`{
			"storeApps": [
				{
					"name": "Store App",
					"version": "2.0",
					"description": "A store app",
					"url": "https://example.com/app.zip",
					"size": "1.5 MB",
					"installed": false
				}
			]
		}`),
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/apps/listStoreApps" {
			t.Errorf("Expected path /api/apps/listStoreApps, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
		Token:      "test-token",
		retries:    1,
	}

	// Test ListStoreApps
	storeApps, err := client.ListStoreApps(context.Background())
	if err != nil {
		t.Fatalf("ListStoreApps failed: %v", err)
	}

	if len(storeApps) != 1 {
		t.Fatalf("Expected 1 store app, got %d", len(storeApps))
	}

	storeApp := storeApps[0]
	if storeApp.Name != "Store App" {
		t.Errorf("Expected store app name 'Store App', got '%s'", storeApp.Name)
	}
	if storeApp.Version != "2.0" {
		t.Errorf("Expected store app version '2.0', got '%s'", storeApp.Version)
	}
	if storeApp.Installed {
		t.Error("Expected store app to not be installed")
	}
}

func TestDownloadAndInstallApp(t *testing.T) {
	// Create mock response
	mockResponse := APIResponse{
		Status: "ok",
		Response: json.RawMessage(`{
			"installedApp": {
				"name": "Downloaded App",
				"version": "1.5",
				"dnsApps": [
					{
						"classPath": "DownloadedApp.App",
						"description": "Downloaded app description",
						"isAppRecordRequestHandler": false,
						"recordDataTemplate": null,
						"isRequestController": true,
						"isAuthoritativeRequestHandler": false,
						"isRequestBlockingHandler": false,
						"isQueryLogger": false,
						"isPostProcessor": false
					}
				]
			}
		}`),
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/apps/downloadAndInstall" {
			t.Errorf("Expected path /api/apps/downloadAndInstall, got %s", r.URL.Path)
		}

		name := r.URL.Query().Get("name")
		url := r.URL.Query().Get("url")

		if name != "test-app" {
			t.Errorf("Expected name 'test-app', got '%s'", name)
		}
		if url != "https://example.com/test.zip" {
			t.Errorf("Expected URL 'https://example.com/test.zip', got '%s'", url)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
		Token:      "test-token",
		retries:    1,
	}

	// Test DownloadAndInstallApp
	app, err := client.DownloadAndInstallApp(context.Background(), "test-app", "https://example.com/test.zip")
	if err != nil {
		t.Fatalf("DownloadAndInstallApp failed: %v", err)
	}

	if app.Name != "Downloaded App" {
		t.Errorf("Expected app name 'Downloaded App', got '%s'", app.Name)
	}
	if app.Version != "1.5" {
		t.Errorf("Expected app version '1.5', got '%s'", app.Version)
	}
}

func TestDownloadAndUpdateApp(t *testing.T) {
	// Create mock response
	mockResponse := APIResponse{
		Status: "ok",
		Response: json.RawMessage(`{
			"updatedApp": {
				"name": "Updated App",
				"version": "2.0",
				"dnsApps": [
					{
						"classPath": "UpdatedApp.App",
						"description": "Updated app description",
						"isAppRecordRequestHandler": false,
						"recordDataTemplate": null,
						"isRequestController": false,
						"isAuthoritativeRequestHandler": true,
						"isRequestBlockingHandler": false,
						"isQueryLogger": false,
						"isPostProcessor": false
					}
				]
			}
		}`),
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/apps/downloadAndUpdate" {
			t.Errorf("Expected path /api/apps/downloadAndUpdate, got %s", r.URL.Path)
		}

		name := r.URL.Query().Get("name")
		url := r.URL.Query().Get("url")

		if name != "test-app" {
			t.Errorf("Expected name 'test-app', got '%s'", name)
		}
		if url != "https://example.com/update.zip" {
			t.Errorf("Expected URL 'https://example.com/update.zip', got '%s'", url)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
		Token:      "test-token",
		retries:    1,
	}

	// Test DownloadAndUpdateApp
	app, err := client.DownloadAndUpdateApp(context.Background(), "test-app", "https://example.com/update.zip")
	if err != nil {
		t.Fatalf("DownloadAndUpdateApp failed: %v", err)
	}

	if app.Name != "Updated App" {
		t.Errorf("Expected app name 'Updated App', got '%s'", app.Name)
	}
	if app.Version != "2.0" {
		t.Errorf("Expected app version '2.0', got '%s'", app.Version)
	}
}

func TestInstallApp(t *testing.T) {
	// Create mock response
	mockResponse := APIResponse{
		Status: "ok",
		Response: json.RawMessage(`{
			"installedApp": {
				"name": "Uploaded App",
				"version": "1.0",
				"dnsApps": [
					{
						"classPath": "UploadedApp.App",
						"description": "Uploaded app description",
						"isAppRecordRequestHandler": true,
						"recordDataTemplate": "{\"value\":\"test\"}",
						"isRequestController": false,
						"isAuthoritativeRequestHandler": false,
						"isRequestBlockingHandler": false,
						"isQueryLogger": false,
						"isPostProcessor": false
					}
				]
			}
		}`),
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/apps/install" {
			t.Errorf("Expected path /api/apps/install, got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Check content type is multipart
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "multipart/form-data") {
			t.Errorf("Expected multipart/form-data content type, got %s", contentType)
		}

		// Parse multipart form
		err := r.ParseMultipartForm(32 << 20) // 32MB
		if err != nil {
			t.Fatalf("Failed to parse multipart form: %v", err)
		}

		name := r.FormValue("name")
		if name != "test-app" {
			t.Errorf("Expected name 'test-app', got '%s'", name)
		}

		// Check if file was uploaded
		_, _, err = r.FormFile("file")
		if err != nil {
			t.Errorf("Expected file upload, got error: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
		Token:      "test-token",
		retries:    1,
	}

	// Test InstallApp with mock app data
	appData := []byte("mock app archive data")
	app, err := client.InstallApp(context.Background(), "test-app", appData)
	if err != nil {
		t.Fatalf("InstallApp failed: %v", err)
	}

	if app.Name != "Uploaded App" {
		t.Errorf("Expected app name 'Uploaded App', got '%s'", app.Name)
	}
	if app.Version != "1.0" {
		t.Errorf("Expected app version '1.0', got '%s'", app.Version)
	}
	if len(app.DNSApps) != 1 {
		t.Fatalf("Expected 1 DNS app, got %d", len(app.DNSApps))
	}

	dnsApp := app.DNSApps[0]
	if dnsApp.RecordDataTemplate == nil {
		t.Error("Expected RecordDataTemplate to not be nil")
	} else if *dnsApp.RecordDataTemplate != "{\"value\":\"test\"}" {
		t.Errorf("Expected RecordDataTemplate '{\"value\":\"test\"}', got '%s'", *dnsApp.RecordDataTemplate)
	}
}

func TestUpdateApp(t *testing.T) {
	// Create mock response
	mockResponse := APIResponse{
		Status: "ok",
		Response: json.RawMessage(`{
			"updatedApp": {
				"name": "Updated Upload App",
				"version": "1.1",
				"dnsApps": [
					{
						"classPath": "UpdatedUploadApp.App",
						"description": "Updated upload app description",
						"isAppRecordRequestHandler": false,
						"recordDataTemplate": null,
						"isRequestController": false,
						"isAuthoritativeRequestHandler": false,
						"isRequestBlockingHandler": true,
						"isQueryLogger": false,
						"isPostProcessor": false
					}
				]
			}
		}`),
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/apps/update" {
			t.Errorf("Expected path /api/apps/update, got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Check content type is multipart
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "multipart/form-data") {
			t.Errorf("Expected multipart/form-data content type, got %s", contentType)
		}

		// Parse multipart form
		err := r.ParseMultipartForm(32 << 20) // 32MB
		if err != nil {
			t.Fatalf("Failed to parse multipart form: %v", err)
		}

		name := r.FormValue("name")
		if name != "test-app" {
			t.Errorf("Expected name 'test-app', got '%s'", name)
		}

		// Check if file was uploaded
		_, _, err = r.FormFile("file")
		if err != nil {
			t.Errorf("Expected file upload, got error: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
		Token:      "test-token",
		retries:    1,
	}

	// Test UpdateApp with mock app data
	appData := []byte("mock updated app archive data")
	app, err := client.UpdateApp(context.Background(), "test-app", appData)
	if err != nil {
		t.Fatalf("UpdateApp failed: %v", err)
	}

	if app.Name != "Updated Upload App" {
		t.Errorf("Expected app name 'Updated Upload App', got '%s'", app.Name)
	}
	if app.Version != "1.1" {
		t.Errorf("Expected app version '1.1', got '%s'", app.Version)
	}
	if len(app.DNSApps) != 1 {
		t.Fatalf("Expected 1 DNS app, got %d", len(app.DNSApps))
	}

	dnsApp := app.DNSApps[0]
	if !dnsApp.IsRequestBlockingHandler {
		t.Error("Expected IsRequestBlockingHandler to be true")
	}
}

func TestUninstallApp(t *testing.T) {
	// Create mock response
	mockResponse := APIResponse{
		Status:   "ok",
		Response: json.RawMessage(`{}`),
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/apps/uninstall" {
			t.Errorf("Expected path /api/apps/uninstall, got %s", r.URL.Path)
		}

		name := r.URL.Query().Get("name")
		if name != "test-app" {
			t.Errorf("Expected name 'test-app', got '%s'", name)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
		Token:      "test-token",
		retries:    1,
	}

	// Test UninstallApp
	err := client.UninstallApp(context.Background(), "test-app")
	if err != nil {
		t.Fatalf("UninstallApp failed: %v", err)
	}
}

func TestGetAppConfig(t *testing.T) {
	configData := `{"key": "value"}`

	// Create mock response with properly escaped JSON
	mockResponse := APIResponse{
		Status: "ok",
		Response: json.RawMessage(`{
			"config": "{\"key\": \"value\"}"
		}`),
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/apps/config/get" {
			t.Errorf("Expected path /api/apps/config/get, got %s", r.URL.Path)
		}

		name := r.URL.Query().Get("name")
		if name != "test-app" {
			t.Errorf("Expected name 'test-app', got '%s'", name)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
		Token:      "test-token",
		retries:    1,
	}

	// Test GetAppConfig
	config, err := client.GetAppConfig(context.Background(), "test-app")
	if err != nil {
		t.Fatalf("GetAppConfig failed: %v", err)
	}

	if config == nil {
		t.Fatal("Expected config to not be nil")
	}
	if *config != configData {
		t.Errorf("Expected config '%s', got '%s'", configData, *config)
	}
}

func TestSetAppConfig(t *testing.T) {
	// Create mock response
	mockResponse := APIResponse{
		Status:   "ok",
		Response: json.RawMessage(`{}`),
	}

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/apps/config/set" {
			t.Errorf("Expected path /api/apps/config/set, got %s", r.URL.Path)
		}

		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		// Parse form data
		err := r.ParseForm()
		if err != nil {
			t.Fatalf("Failed to parse form: %v", err)
		}

		name := r.FormValue("name")
		config := r.FormValue("config")

		if name != "test-app" {
			t.Errorf("Expected name 'test-app', got '%s'", name)
		}
		if config != "{\"setting\":\"value\"}" {
			t.Errorf("Expected config '{\"setting\":\"value\"}', got '%s'", config)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	}))
	defer server.Close()

	// Create client
	client := &Client{
		BaseURL:    server.URL,
		HTTPClient: server.Client(),
		Token:      "test-token",
		retries:    1,
	}

	// Test SetAppConfig
	err := client.SetAppConfig(context.Background(), "test-app", "{\"setting\":\"value\"}")
	if err != nil {
		t.Fatalf("SetAppConfig failed: %v", err)
	}
}
