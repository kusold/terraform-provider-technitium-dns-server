package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ZoneDataSource{}

func NewZoneDataSource() datasource.DataSource {
	return &ZoneDataSource{}
}

// ZoneDataSource defines the data source implementation.
type ZoneDataSource struct {
	client *client.Client
}

// ZoneDataSourceModel describes the data source data model.
type ZoneDataSourceModel struct {
	// Required inputs
	Name types.String `tfsdk:"name"`

	// Computed outputs - using the same structure as ZoneResourceModel for consistency
	ID                         types.String `tfsdk:"id"`
	Type                       types.String `tfsdk:"type"`
	Catalog                    types.String `tfsdk:"catalog"`
	UseSoaSerialDateScheme     types.Bool   `tfsdk:"use_soa_serial_date_scheme"`
	PrimaryNameServerAddresses types.String `tfsdk:"primary_name_server_addresses"`
	ZoneTransferProtocol       types.String `tfsdk:"zone_transfer_protocol"`
	TsigKeyName                types.String `tfsdk:"tsig_key_name"`
	ValidateZone               types.Bool   `tfsdk:"validate_zone"`
	InitializeForwarder        types.Bool   `tfsdk:"initialize_forwarder"`
	Protocol                   types.String `tfsdk:"protocol"`
	Forwarder                  types.String `tfsdk:"forwarder"`
	DnssecValidation           types.Bool   `tfsdk:"dnssec_validation"`
	ProxyType                  types.String `tfsdk:"proxy_type"`
	ProxyAddress               types.String `tfsdk:"proxy_address"`
	ProxyPort                  types.Int64  `tfsdk:"proxy_port"`
	ProxyUsername              types.String `tfsdk:"proxy_username"`
	ProxyPassword              types.String `tfsdk:"proxy_password"`

	// Read-only computed attributes
	Internal     types.Bool   `tfsdk:"internal"`
	DnssecStatus types.String `tfsdk:"dnssec_status"`
	Disabled     types.Bool   `tfsdk:"disabled"`
	SoaSerial    types.Int64  `tfsdk:"soa_serial"`
}

func (d *ZoneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (d *ZoneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Technitium DNS Server zone data source",
		MarkdownDescription: "Technitium DNS Server zone data source",

		Attributes: map[string]schema.Attribute{
			// Required input
			"name": schema.StringAttribute{
				MarkdownDescription: "The domain name for the zone to retrieve.",
				Required:            true,
			},

			// Output attributes - using the same structure as ZoneResource for consistency
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the zone resource.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of zone. Valid values are: Primary, Secondary, Stub, Forwarder, SecondaryForwarder, Catalog, SecondaryCatalog.",
				Computed:            true,
			},
			"catalog": schema.StringAttribute{
				MarkdownDescription: "The name of the catalog zone to become its member zone. Valid only for Primary, Stub, and Forwarder zones.",
				Computed:            true,
			},
			"use_soa_serial_date_scheme": schema.BoolAttribute{
				MarkdownDescription: "Indicates if the zone uses date scheme for SOA serial.",
				Computed:            true,
			},
			"primary_name_server_addresses": schema.StringAttribute{
				MarkdownDescription: "Comma separated list of IP addresses or domain names of the primary name server. Used only with Secondary, SecondaryForwarder, SecondaryCatalog, and Stub zones.",
				Computed:            true,
			},
			"zone_transfer_protocol": schema.StringAttribute{
				MarkdownDescription: "The zone transfer protocol used. Valid values are: Tcp, Tls, Quic. Used by Secondary, SecondaryForwarder, and SecondaryCatalog zones.",
				Computed:            true,
			},
			"tsig_key_name": schema.StringAttribute{
				MarkdownDescription: "The TSIG key name used. Used by Secondary, SecondaryForwarder, and SecondaryCatalog zones.",
				Computed:            true,
			},
			"validate_zone": schema.BoolAttribute{
				MarkdownDescription: "Indicates if ZONEMD validation is enabled. Valid only for Secondary zones.",
				Computed:            true,
			},
			"initialize_forwarder": schema.BoolAttribute{
				MarkdownDescription: "Indicates if the Conditional Forwarder zone is initialized with an FWD record. Valid for Forwarder zones.",
				Computed:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "The DNS transport protocol used by the Conditional Forwarder zone. Valid values are: Udp, Tcp, Tls, Https, Quic.",
				Computed:            true,
			},
			"forwarder": schema.StringAttribute{
				MarkdownDescription: "The address of the DNS server used as a forwarder. 'this-server' means forwarding internally.",
				Computed:            true,
			},
			"dnssec_validation": schema.BoolAttribute{
				MarkdownDescription: "Indicates if DNSSEC validation is done. Used with Conditional Forwarder zones.",
				Computed:            true,
			},
			"proxy_type": schema.StringAttribute{
				MarkdownDescription: "The type of proxy for conditional forwarding. Valid values are: NoProxy, DefaultProxy, Http, Socks5.",
				Computed:            true,
			},
			"proxy_address": schema.StringAttribute{
				MarkdownDescription: "The proxy server address used.",
				Computed:            true,
				Sensitive:           true,
			},
			"proxy_port": schema.Int64Attribute{
				MarkdownDescription: "The proxy server port used.",
				Computed:            true,
			},
			"proxy_username": schema.StringAttribute{
				MarkdownDescription: "The proxy server username used.",
				Computed:            true,
				Sensitive:           true,
			},
			"proxy_password": schema.StringAttribute{
				MarkdownDescription: "The proxy server password used.",
				Computed:            true,
				Sensitive:           true,
			},

			// Computed attributes
			"internal": schema.BoolAttribute{
				MarkdownDescription: "Indicates if this is an internal zone.",
				Computed:            true,
			},
			"dnssec_status": schema.StringAttribute{
				MarkdownDescription: "The DNSSEC status of the zone.",
				Computed:            true,
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Indicates if the zone is disabled.",
				Computed:            true,
			},
			"soa_serial": schema.Int64Attribute{
				MarkdownDescription: "The SOA serial number of the zone.",
				Computed:            true,
			},
		},
	}
}

func (d *ZoneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

// Using the struct definitions from zone_resource.go

func (d *ZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ZoneDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneName := data.Name.ValueString()
	tflog.Debug(ctx, "Reading zone data source", map[string]interface{}{
		"name": zoneName,
	})

	// Get zone info from the API
	zoneInfo, err := d.client.GetZone(ctx, zoneName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading zone",
			fmt.Sprintf("Could not read zone %s: %s", zoneName, err.Error()),
		)
		return
	}

	// Set ID (same as name)
	data.ID = types.StringValue(zoneName)
	data.Type = types.StringValue(zoneInfo.Type)
	data.Internal = types.BoolValue(zoneInfo.Internal)
	data.DnssecStatus = types.StringValue(zoneInfo.DnssecStatus)
	data.Disabled = types.BoolValue(zoneInfo.Disabled)

	// Get zone options directly from the API
	params := url.Values{}
	params.Set("zone", zoneName)
	endpoint := "/api/zones/options/get?" + params.Encode()

	type zoneOptionsResponse struct {
		Name                           string   `json:"name"`
		Type                           string   `json:"type"`
		Internal                       bool     `json:"internal"`
		DnssecStatus                   string   `json:"dnssecStatus"`
		Disabled                       bool     `json:"disabled"`
		Catalog                        string   `json:"catalog,omitempty"`
		UseSoaSerialDateScheme         *bool    `json:"useSoaSerialDateScheme,omitempty"`
		PrimaryNameServerAddresses     []string `json:"primaryNameServerAddresses,omitempty"`
		PrimaryZoneTransferProtocol    string   `json:"primaryZoneTransferProtocol,omitempty"`
		PrimaryZoneTransferTsigKeyName string   `json:"primaryZoneTransferTsigKeyName,omitempty"`
		ValidateZone                   *bool    `json:"validateZone,omitempty"`
	}

	var options zoneOptionsResponse
	if err := d.client.DoRequest(ctx, "GET", endpoint, nil, &options); err != nil {
		resp.Diagnostics.AddError(
			"Error reading zone options",
			fmt.Sprintf("Could not read options for zone %s: %s", zoneName, err.Error()),
		)
		return
	}

	// Update model with zone options
	if options.Catalog != "" {
		data.Catalog = types.StringValue(options.Catalog)
	}

	if options.UseSoaSerialDateScheme != nil {
		data.UseSoaSerialDateScheme = types.BoolValue(*options.UseSoaSerialDateScheme)
	} else {
		data.UseSoaSerialDateScheme = types.BoolValue(false)
	}

	if len(options.PrimaryNameServerAddresses) > 0 {
		data.PrimaryNameServerAddresses = types.StringValue(strings.Join(options.PrimaryNameServerAddresses, ","))
	}

	if options.PrimaryZoneTransferProtocol != "" {
		data.ZoneTransferProtocol = types.StringValue(options.PrimaryZoneTransferProtocol)
	} else {
		data.ZoneTransferProtocol = types.StringValue("Tcp")
	}

	if options.PrimaryZoneTransferTsigKeyName != "" {
		data.TsigKeyName = types.StringValue(options.PrimaryZoneTransferTsigKeyName)
	}

	if options.ValidateZone != nil {
		data.ValidateZone = types.BoolValue(*options.ValidateZone)
	} else {
		data.ValidateZone = types.BoolValue(false)
	}

	// Set default values for computed fields
	if data.InitializeForwarder.IsNull() || data.InitializeForwarder.IsUnknown() {
		data.InitializeForwarder = types.BoolValue(false)
	}

	if data.DnssecValidation.IsNull() || data.DnssecValidation.IsUnknown() {
		data.DnssecValidation = types.BoolValue(false)
	}

	// Default protocol and proxy type
	data.Protocol = types.StringValue("Udp")
	data.ProxyType = types.StringValue("DefaultProxy")

	// Get zone records to extract SOA serial
	// Use the client's DoRequest method directly since the API has specific formats for each record type
	recordsParams := url.Values{}
	recordsParams.Set("domain", zoneName)
	recordsParams.Set("zone", zoneName)
	recordsParams.Set("listZone", "true")
	recordsEndpoint := "/api/zones/records/get?" + recordsParams.Encode()

	// Define a simple structure for SOA record responses
	type soaRData struct {
		Serial uint32 `json:"serial"`
	}

	type recordRData struct {
		SoaRecord *soaRData `json:"soaRecord,omitempty"`
	}

	type zoneRecord struct {
		Type  string      `json:"type"`
		RData recordRData `json:"rData"`
	}

	type zoneRecordsResponse struct {
		Records []zoneRecord `json:"records"`
	}

	var recordsResponse zoneRecordsResponse
	if err := d.client.DoRequest(ctx, "GET", recordsEndpoint, nil, &recordsResponse); err != nil {
		// Don't fail if records can't be read, just log it
		tflog.Warn(ctx, "Failed to read zone records for SOA serial", map[string]interface{}{
			"zone":  zoneName,
			"error": err.Error(),
		})
	} else {
		// Find SOA record to get serial
		soaFound := false
		for _, record := range recordsResponse.Records {
			if record.Type == "SOA" && record.RData.SoaRecord != nil {
				data.SoaSerial = types.Int64Value(int64(record.RData.SoaRecord.Serial))
				soaFound = true
				break
			}
		}
		if !soaFound {
			// Default SOA serial if not found
			data.SoaSerial = types.Int64Value(1)
		}
	}

	// Ensure SoaSerial is set even if records couldn't be read
	if data.SoaSerial.IsNull() || data.SoaSerial.IsUnknown() {
		data.SoaSerial = types.Int64Value(1)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
