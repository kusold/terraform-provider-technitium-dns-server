package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Client represents the Technitium DNS API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      string
	username   string
	password   string
	retries    int
}

// Config holds the configuration for creating a new client
type Config struct {
	Host               string
	Username           string
	Password           string
	Token              string
	TimeoutSeconds     int64
	RetryAttempts      int64
	InsecureSkipVerify bool
}

// APIResponse represents the standard API response format
type APIResponse struct {
	Status   string          `json:"status"`
	Response json.RawMessage `json:"response,omitempty"`
	// Error fields
	ErrorMessage string `json:"errorMessage,omitempty"`
	Error        string `json:"error,omitempty"`
}

// LoginResponse represents the login API response
type LoginResponse struct {
	DisplayName string `json:"displayName"`
	Username    string `json:"username"`
	Token       string `json:"token"`
}

// NewClient creates a new Technitium DNS API client
func NewClient(config Config) (*Client, error) {
	// Set defaults
	if config.TimeoutSeconds == 0 {
		config.TimeoutSeconds = 30
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}

	// Validate configuration
	if config.Host == "" {
		return nil, fmt.Errorf("host is required")
	}

	// Ensure we have authentication
	if config.Token == "" && (config.Username == "" || config.Password == "") {
		return nil, fmt.Errorf("either token or username/password must be provided")
	}

	// Create HTTP client
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			//nolint:gosec // G402: InsecureSkipVerify is an intentional user-configurable option for development/testing
			InsecureSkipVerify: config.InsecureSkipVerify,
		},
	}

	httpClient := &http.Client{
		Timeout:   time.Duration(config.TimeoutSeconds) * time.Second,
		Transport: transport,
	}

	client := &Client{
		BaseURL:    strings.TrimSuffix(config.Host, "/"),
		HTTPClient: httpClient,
		Token:      config.Token,
		username:   config.Username,
		password:   config.Password,
		retries:    int(config.RetryAttempts),
	}

	return client, nil
}

// Login authenticates with the Technitium DNS server using username/password
func (c *Client) Login(ctx context.Context) error {
	if c.username == "" || c.password == "" {
		return fmt.Errorf("username and password are required for login")
	}

	params := url.Values{}
	params.Set("user", c.username)
	params.Set("pass", c.password)
	params.Set("includeInfo", "true")

	endpoint := "/api/user/login?" + params.Encode()

	tflog.Debug(ctx, "Attempting login to", map[string]interface{}{
		"endpoint": endpoint,
		"username": c.username,
	})

	// Login endpoint returns data directly, not wrapped in APIResponse
	var response LoginResponse
	if err := c.makeLoginRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	tflog.Debug(ctx, "Login response received", map[string]interface{}{
		"username":    response.Username,
		"displayName": response.DisplayName,
		"token":       response.Token,
		"token_empty": response.Token == "",
	})

	c.Token = response.Token
	tflog.Debug(ctx, "Successfully authenticated with Technitium DNS server", map[string]interface{}{
		"username":     response.Username,
		"displayName":  response.DisplayName,
		"token":        response.Token,
		"token_length": len(response.Token),
	})

	return nil
}

// doRequest performs an HTTP request with retry logic
func (c *Client) doRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	var lastErr error

	for attempt := 0; attempt <= c.retries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(attempt) * time.Second
			tflog.Debug(ctx, "Retrying request after backoff", map[string]interface{}{
				"attempt":  attempt,
				"backoff":  backoff.String(),
				"endpoint": endpoint,
			})

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := c.makeRequest(ctx, method, endpoint, body, result)
		if err == nil {
			return nil
		}

		lastErr = err
		tflog.Debug(ctx, "Request failed", map[string]interface{}{
			"attempt":  attempt + 1,
			"error":    err.Error(),
			"endpoint": endpoint,
		})

		// Don't retry on certain errors
		if strings.Contains(err.Error(), "invalid-token") && c.username != "" && c.password != "" {
			// Try to re-authenticate
			if loginErr := c.Login(ctx); loginErr != nil {
				return fmt.Errorf("authentication failed: %w", loginErr)
			}
			continue
		}
	}

	return lastErr
}

// makeLoginRequest performs a single HTTP request for login (which returns data directly)
func (c *Client) makeLoginRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	// Prepare request URL
	requestURL := c.BaseURL + endpoint

	// Prepare request body
	var requestBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		requestBody = bytes.NewBuffer(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, requestURL, requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Log request
	tflog.Debug(ctx, "Making login API request", map[string]interface{}{
		"method": method,
		"url":    requestURL,
	})

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

	// Log response
	tflog.Debug(ctx, "Received login API response", map[string]interface{}{
		"status_code":     resp.StatusCode,
		"response_length": len(responseBody),
		"response_body":   string(responseBody),
	})

	// Check HTTP status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Login endpoint returns data directly, not wrapped in APIResponse
	if result != nil {
		if err := json.Unmarshal(responseBody, result); err != nil {
			return fmt.Errorf("failed to parse login response: %w", err)
		}
	}

	return nil
}

// makeRequest performs a single HTTP request
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	// Prepare request URL
	requestURL := c.BaseURL + endpoint

	// Add token to URL if we have one and it's not already in the endpoint
	if c.Token != "" && !strings.Contains(endpoint, "token=") {
		separator := "?"
		if strings.Contains(endpoint, "?") {
			separator = "&"
		}
		requestURL += separator + "token=" + url.QueryEscape(c.Token)
	}

	// Prepare request body
	var requestBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		requestBody = bytes.NewBuffer(jsonBody)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, method, requestURL, requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Log request
	tflog.Debug(ctx, "Making API request", map[string]interface{}{
		"method": method,
		"url":    requestURL,
	})

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

	// Log response
	tflog.Debug(ctx, "Received API response", map[string]interface{}{
		"status_code":     resp.StatusCode,
		"response_length": len(responseBody),
	})

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

// Authenticate ensures the client is authenticated
func (c *Client) Authenticate(ctx context.Context) error {
	// If we already have a token, we're good
	if c.Token != "" {
		return nil
	}

	// If we have username/password, login
	if c.username != "" && c.password != "" {
		return c.Login(ctx)
	}

	return fmt.Errorf("no authentication method available")
}

// DoRequest performs an HTTP request with authentication and retry logic
func (c *Client) DoRequest(ctx context.Context, method, endpoint string, body interface{}, result interface{}) error {
	if err := c.Authenticate(ctx); err != nil {
		return err
	}
	return c.doRequest(ctx, method, endpoint, body, result)
}
