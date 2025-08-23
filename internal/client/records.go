package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// DNSRecord represents a DNS record
type DNSRecord struct {
	Name         string        `json:"name"`
	Type         string        `json:"type"`
	TTL          int           `json:"ttl"`
	RData        DNSRecordData `json:"rData"`
	Disabled     bool          `json:"disabled"`
	DnssecStatus string        `json:"dnssecStatus"`
	Comments     string        `json:"comments,omitempty"`
	LastUsedOn   string        `json:"lastUsedOn,omitempty"`
}

// DNSRecordData represents the record-specific data for a DNS record
type DNSRecordData struct {
	// A record
	IPAddress string `json:"ipAddress,omitempty"`

	// AAAA record uses the same IPAddress field

	// CNAME record
	CNAME string `json:"cname,omitempty"`

	// MX record
	Exchange   string `json:"exchange,omitempty"`
	Preference int    `json:"preference,omitempty"`

	// TXT record
	Text string `json:"text,omitempty"`

	// PTR record
	PTRName string `json:"ptrName,omitempty"`

	// NS record
	NameServer string `json:"nameServer,omitempty"`

	// SRV record
	Priority int    `json:"priority,omitempty"`
	Weight   int    `json:"weight,omitempty"`
	Port     int    `json:"port,omitempty"`
	Target   string `json:"target,omitempty"`

	// SOA record
	PrimaryNameServer string `json:"primaryNameServer,omitempty"`
	ResponsiblePerson string `json:"responsiblePerson,omitempty"`
	Serial            uint32 `json:"serial,omitempty"`
	Refresh           int    `json:"refresh,omitempty"`
	Retry             int    `json:"retry,omitempty"`
	Expire            int    `json:"expire,omitempty"`
	Minimum           int    `json:"minimum,omitempty"`
}

// AddRecordResponse represents the API response when adding a DNS record
type AddRecordResponse struct {
	Zone        ZoneInfo  `json:"zone"`
	AddedRecord DNSRecord `json:"addedRecord"`
}

// UpdateRecordResponse represents the API response when updating a DNS record
type UpdateRecordResponse struct {
	Zone          ZoneInfo  `json:"zone"`
	UpdatedRecord DNSRecord `json:"updatedRecord"`
}

// GetRecordsResponse represents the API response for listing DNS records
type GetRecordsResponse struct {
	Zone    ZoneInfo    `json:"zone"`
	Records []DNSRecord `json:"records"`
}

// AddRecord adds a new DNS record
func (c *Client) AddRecord(ctx context.Context, zone, domain, recordType string, ttl int, options map[string]string) (*AddRecordResponse, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("domain", domain)
	params.Set("zone", zone)
	params.Set("type", recordType)
	params.Set("ttl", fmt.Sprintf("%d", ttl))

	// Add additional options based on record type
	for key, value := range options {
		params.Set(key, value)
	}

	endpoint := "/api/zones/records/add?" + params.Encode()

	var response AddRecordResponse
	if err := c.DoRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to add DNS record: %w", err)
	}

	return &response, nil
}

// GetRecords retrieves DNS records for a zone or domain
func (c *Client) GetRecords(ctx context.Context, zone, domain string, listZone bool) (*GetRecordsResponse, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("domain", domain)
	params.Set("zone", zone)

	if listZone {
		params.Set("listZone", "true")
	}

	endpoint := "/api/zones/records/get?" + params.Encode()

	var response GetRecordsResponse
	if err := c.DoRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get DNS records: %w", err)
	}

	return &response, nil
}

// UpdateRecord updates an existing DNS record
func (c *Client) UpdateRecord(ctx context.Context, zone, domain, recordType string, options map[string]string) (*UpdateRecordResponse, error) {
	if err := c.Authenticate(ctx); err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("domain", domain)
	params.Set("zone", zone)
	params.Set("type", recordType)

	// Add additional options based on record type and update operation
	for key, value := range options {
		params.Set(key, value)
	}

	endpoint := "/api/zones/records/update?" + params.Encode()

	var response UpdateRecordResponse
	if err := c.DoRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to update DNS record: %w", err)
	}

	return &response, nil
}

// DeleteRecord deletes a DNS record
func (c *Client) DeleteRecord(ctx context.Context, zone, domain, recordType string, options map[string]string) error {
	if err := c.Authenticate(ctx); err != nil {
		return err
	}

	params := url.Values{}
	params.Set("domain", domain)
	params.Set("zone", zone)
	params.Set("type", recordType)

	// Add record-specific options required for deletion
	for key, value := range options {
		params.Set(key, value)
	}

	endpoint := "/api/zones/records/delete?" + params.Encode()

	if err := c.DoRequest(ctx, http.MethodGet, endpoint, nil, nil); err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}

	return nil
}
