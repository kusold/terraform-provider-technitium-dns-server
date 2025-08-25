package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

// DNSApp represents a single DNS application within an app package
type DNSApp struct {
	ClassPath                     string  `json:"classPath"`
	Description                   string  `json:"description"`
	IsAppRecordRequestHandler     bool    `json:"isAppRecordRequestHandler"`
	RecordDataTemplate            *string `json:"recordDataTemplate"`
	IsRequestController           bool    `json:"isRequestController"`
	IsAuthoritativeRequestHandler bool    `json:"isAuthoritativeRequestHandler"`
	IsRequestBlockingHandler      bool    `json:"isRequestBlockingHandler"`
	IsQueryLogger                 bool    `json:"isQueryLogger"`
	IsPostProcessor               bool    `json:"isPostProcessor"`
}

// App represents an installed DNS application
type App struct {
	Name    string   `json:"name"`
	Version string   `json:"version"`
	DNSApps []DNSApp `json:"dnsApps"`
}

// StoreApp represents an app available in the DNS App Store
type StoreApp struct {
	Name             string `json:"name"`
	Version          string `json:"version"`
	Description      string `json:"description"`
	URL              string `json:"url"`
	Size             string `json:"size"`
	Installed        bool   `json:"installed"`
	InstalledVersion string `json:"installedVersion,omitempty"`
	UpdateAvailable  bool   `json:"updateAvailable,omitempty"`
}

// ListAppsResponse represents the response from the list apps API
type ListAppsResponse struct {
	Apps []App `json:"apps"`
}

// ListStoreAppsResponse represents the response from the list store apps API
type ListStoreAppsResponse struct {
	StoreApps []StoreApp `json:"storeApps"`
}

// InstallAppResponse represents the response from install/update app APIs
type InstallAppResponse struct {
	InstalledApp App `json:"installedApp,omitempty"`
	UpdatedApp   App `json:"updatedApp,omitempty"`
}

// GetAppConfigResponse represents the response from get app config API
type GetAppConfigResponse struct {
	Config *string `json:"config"`
}

// ListApps lists all installed apps on the DNS server
func (c *Client) ListApps(ctx context.Context) ([]App, error) {
	endpoint := "/api/apps/list"

	var response ListAppsResponse
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}

	return response.Apps, nil
}

// ListStoreApps lists all available apps on the DNS App Store
func (c *Client) ListStoreApps(ctx context.Context) ([]StoreApp, error) {
	endpoint := "/api/apps/listStoreApps"

	var response ListStoreAppsResponse
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list store apps: %w", err)
	}

	return response.StoreApps, nil
}

// DownloadAndInstallApp downloads an app zip file from URL and installs it
func (c *Client) DownloadAndInstallApp(ctx context.Context, name, appURL string) (*App, error) {
	params := url.Values{}
	params.Set("name", name)
	params.Set("url", appURL)

	endpoint := "/api/apps/downloadAndInstall?" + params.Encode()

	var response InstallAppResponse
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to download and install app: %w", err)
	}

	return &response.InstalledApp, nil
}

// DownloadAndUpdateApp downloads an app zip file from URL and updates an existing app
func (c *Client) DownloadAndUpdateApp(ctx context.Context, name, appURL string) (*App, error) {
	params := url.Values{}
	params.Set("name", name)
	params.Set("url", appURL)

	endpoint := "/api/apps/downloadAndUpdate?" + params.Encode()

	var response InstallAppResponse
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to download and update app: %w", err)
	}

	return &response.UpdatedApp, nil
}

// InstallApp installs a DNS application from uploaded zip file
func (c *Client) InstallApp(ctx context.Context, name string, appData []byte) (*App, error) {
	params := url.Values{}
	params.Set("name", name)

	endpoint := "/api/apps/install?" + params.Encode()

	// Add token to URL if we have one
	if c.Token != "" {
		endpoint += "&token=" + url.QueryEscape(c.Token)
	}

	var response InstallAppResponse
	if err := c.makeMultipartRequest(ctx, "POST", endpoint, "app.zip", appData, &response); err != nil {
		return nil, fmt.Errorf("failed to install app: %w", err)
	}

	return &response.InstalledApp, nil
}

// UpdateApp updates an installed app using a provided app zip file
func (c *Client) UpdateApp(ctx context.Context, name string, appData []byte) (*App, error) {
	params := url.Values{}
	params.Set("name", name)

	endpoint := "/api/apps/update?" + params.Encode()

	// Add token to URL if we have one
	if c.Token != "" {
		endpoint += "&token=" + url.QueryEscape(c.Token)
	}

	var response InstallAppResponse
	if err := c.makeMultipartRequest(ctx, "POST", endpoint, "app.zip", appData, &response); err != nil {
		return nil, fmt.Errorf("failed to update app: %w", err)
	}

	return &response.UpdatedApp, nil
}

// UninstallApp uninstalls an app from the DNS server
func (c *Client) UninstallApp(ctx context.Context, name string) error {
	params := url.Values{}
	params.Set("name", name)

	endpoint := "/api/apps/uninstall?" + params.Encode()

	if err := c.DoRequest(ctx, "GET", endpoint, nil, nil); err != nil {
		return fmt.Errorf("failed to uninstall app: %w", err)
	}

	return nil
}

// GetAppConfig retrieves the DNS application config from the dnsApp.config file
func (c *Client) GetAppConfig(ctx context.Context, name string) (*string, error) {
	params := url.Values{}
	params.Set("name", name)

	endpoint := "/api/apps/config/get?" + params.Encode()

	var response GetAppConfigResponse
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get app config: %w", err)
	}

	return response.Config, nil
}

// SetAppConfig saves the provided DNS application config into the dnsApp.config file
func (c *Client) SetAppConfig(ctx context.Context, name, config string) error {
	params := url.Values{}
	params.Set("name", name)

	endpoint := "/api/apps/config/set?" + params.Encode()

	// Add token to URL if we have one
	if c.Token != "" {
		endpoint += "&token=" + url.QueryEscape(c.Token)
	}

	// Pretty-format the JSON config with 2-space indentation before sending
	formattedConfig := config
	if config != "" {
		var jsonData interface{}
		if err := json.Unmarshal([]byte(config), &jsonData); err == nil {
			if prettyJSON, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
				formattedConfig = string(prettyJSON)
			}
			// If formatting fails, we'll use the original config
		}
	}

	// Create form data
	formData := url.Values{}
	formData.Set("config", formattedConfig)

	if err := c.makeFormRequest(ctx, "POST", endpoint, formData, nil); err != nil {
		return fmt.Errorf("failed to set app config: %w", err)
	}

	return nil
}

// makeMultipartRequest performs a multipart form-data HTTP request for file uploads
func (c *Client) makeMultipartRequest(ctx context.Context, method, endpoint, fileName string, fileData []byte, result interface{}) error {
	// Prepare request URL
	requestURL := c.BaseURL + endpoint

	// Create multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add file part
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(fileData); err != nil {
		return fmt.Errorf("failed to write file data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, requestURL, &body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type header
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Make request using the same pattern as other requests
	return c.executeRequest(ctx, req, result)
}

// makeFormRequest performs a form-encoded HTTP request
func (c *Client) makeFormRequest(ctx context.Context, method, endpoint string, formData url.Values, result interface{}) error {
	// Prepare request URL
	requestURL := c.BaseURL + endpoint

	// Create request body
	requestBody := bytes.NewBufferString(formData.Encode())

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, requestURL, requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set content type header
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make request using the same pattern as other requests
	return c.executeRequest(ctx, req, result)
}

// executeRequest executes an HTTP request and handles the response
func (c *Client) executeRequest(ctx context.Context, req *http.Request, result interface{}) error {
	// Make request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Parse API response
	var apiResp APIResponse
	if err := json.Unmarshal(responseBody, &apiResp); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}

	// Check API status
	switch apiResp.Status {
	case "ok":
		// Success - unmarshal the response into result if provided
		if result != nil && apiResp.Response != nil {
			if err := json.Unmarshal(apiResp.Response, result); err != nil {
				return fmt.Errorf("failed to parse response data: %w", err)
			}
		}
		return nil
	case "error":
		errorMsg := apiResp.ErrorMessage
		if errorMsg == "" {
			errorMsg = apiResp.Error
		}
		if errorMsg == "" {
			errorMsg = "unknown error"
		}
		return fmt.Errorf("API error: %s", errorMsg)
	case "invalid-token":
		return fmt.Errorf("invalid-token: session expired or invalid token")
	default:
		return fmt.Errorf("unexpected API status: %s", apiResp.Status)
	}
}
