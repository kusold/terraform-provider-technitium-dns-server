package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Zone represents a DNS zone
type Zone struct {
	Name             string `json:"name"`
	Type             string `json:"type"`
	Internal         bool   `json:"internal"`
	Disabled         bool   `json:"disabled"`
	DnssecStatus     string `json:"dnssecStatus,omitempty"`
	NotifyFailed     bool   `json:"notifyFailed,omitempty"`
	Expiry           string `json:"expiry,omitempty"`
	LastModified     string `json:"lastModified,omitempty"`
}

// ZoneInfo represents detailed zone information
type ZoneInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	TypeName     string `json:"typeName"`
	Disabled     bool   `json:"disabled"`
	Internal     bool   `json:"internal"`
	DnssecStatus string `json:"dnssecStatus"`
	NotifyFailed bool   `json:"notifyFailed"`
	Expiry       string `json:"expiry,omitempty"`
	LastModified string `json:"lastModified"`
}

// ZoneListResponse represents the response from zones/list API
type ZoneListResponse struct {
	PageNumber   int    `json:"pageNumber"`
	TotalPages   int    `json:"totalPages"`
	TotalZones   int    `json:"totalZones"`
	Zones        []Zone `json:"zones"`
}

// CreateZoneRequest represents the request to create a zone
type CreateZoneRequest struct {
	Zone string `json:"zone"`
	Type string `json:"type"`
}

// ListZones retrieves all zones from the DNS server
func (c *Client) ListZones(ctx context.Context) ([]Zone, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}

	endpoint := "/api/zones/list"
	
	var response ZoneListResponse
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to list zones: %w", err)
	}

	return response.Zones, nil
}

// GetZone retrieves information about a specific zone
func (c *Client) GetZone(ctx context.Context, zoneName string) (*ZoneInfo, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}

	// For getting zone details, we can use the list endpoint with pagination
	// to find our specific zone, or we can use zone/options to get zone info
	params := url.Values{}
	params.Set("zone", zoneName)
	
	endpoint := "/api/zones/options/get?" + params.Encode()
	
	var response ZoneInfo
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get zone %s: %w", zoneName, err)
	}

	return &response, nil
}

// CreateZone creates a new DNS zone
func (c *Client) CreateZone(ctx context.Context, zoneName, zoneType string) error {
	if err := c.Authenticate(ctx); err != nil {
		return err
	}

	params := url.Values{}
	params.Set("zone", zoneName)
	params.Set("type", zoneType)
	
	endpoint := "/api/zones/create?" + params.Encode()
	
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, nil); err != nil {
		return fmt.Errorf("failed to create zone %s: %w", zoneName, err)
	}

	return nil
}

// DeleteZone deletes a DNS zone
func (c *Client) DeleteZone(ctx context.Context, zoneName string) error {
	if err := c.Authenticate(ctx); err != nil {
		return err
	}

	params := url.Values{}
	params.Set("zone", zoneName)
	
	endpoint := "/api/zones/delete?" + params.Encode()
	
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, nil); err != nil {
		return fmt.Errorf("failed to delete zone %s: %w", zoneName, err)
	}

	return nil
}

// EnableZone enables a DNS zone
func (c *Client) EnableZone(ctx context.Context, zoneName string) error {
	if err := c.Authenticate(ctx); err != nil {
		return err
	}

	params := url.Values{}
	params.Set("zone", zoneName)
	
	endpoint := "/api/zones/enable?" + params.Encode()
	
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, nil); err != nil {
		return fmt.Errorf("failed to enable zone %s: %w", zoneName, err)
	}

	return nil
}

// DisableZone disables a DNS zone
func (c *Client) DisableZone(ctx context.Context, zoneName string) error {
	if err := c.Authenticate(ctx); err != nil {
		return err
	}

	params := url.Values{}
	params.Set("zone", zoneName)
	
	endpoint := "/api/zones/disable?" + params.Encode()
	
	if err := c.doRequest(ctx, http.MethodGet, endpoint, nil, nil); err != nil {
		return fmt.Errorf("failed to disable zone %s: %w", zoneName, err)
	}

	return nil
}

// ZoneExists checks if a zone exists
func (c *Client) ZoneExists(ctx context.Context, zoneName string) (bool, error) {
	zones, err := c.ListZones(ctx)
	if err != nil {
		return false, err
	}

	for _, zone := range zones {
		if strings.EqualFold(zone.Name, zoneName) {
			return true, nil
		}
	}

	return false, nil
}