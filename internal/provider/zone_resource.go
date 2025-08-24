package provider

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ZoneResource{}
var _ resource.ResourceWithImportState = &ZoneResource{}

func NewZoneResource() resource.Resource {
	return &ZoneResource{}
}

// ZoneResource defines the resource implementation.
type ZoneResource struct {
	client *client.Client
}

// ZoneResourceModel describes the resource data model.
type ZoneResourceModel struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
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

func (r *ZoneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (r *ZoneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Technitium DNS Server zone resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the zone resource.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The domain name for the zone. Can be a valid domain name, IP address, or network address in CIDR format for reverse zones.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of zone to create. Valid values are: Primary, Secondary, Stub, Forwarder, SecondaryForwarder, Catalog, SecondaryCatalog.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("Primary", "Secondary", "Stub", "Forwarder", "SecondaryForwarder", "Catalog", "SecondaryCatalog"),
				},
			},
			"catalog": schema.StringAttribute{
				MarkdownDescription: "The name of the catalog zone to become its member zone. Valid only for Primary, Stub, and Forwarder zones.",
				Optional:            true,
			},
			"use_soa_serial_date_scheme": schema.BoolAttribute{
				MarkdownDescription: "Set to true to enable using date scheme for SOA serial. Valid for Primary, Forwarder, and Catalog zones.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"primary_name_server_addresses": schema.StringAttribute{
				MarkdownDescription: "Comma separated list of IP addresses or domain names of the primary name server. Used only with Secondary, SecondaryForwarder, SecondaryCatalog, and Stub zones.",
				Optional:            true,
			},
			"zone_transfer_protocol": schema.StringAttribute{
				MarkdownDescription: "The zone transfer protocol to be used. Valid values are: Tcp, Tls, Quic. Used by Secondary, SecondaryForwarder, and SecondaryCatalog zones.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Tcp"),
				Validators: []validator.String{
					stringvalidator.OneOf("Tcp", "Tls", "Quic"),
				},
			},
			"tsig_key_name": schema.StringAttribute{
				MarkdownDescription: "The TSIG key name to be used. Used by Secondary, SecondaryForwarder, and SecondaryCatalog zones.",
				Optional:            true,
			},
			"validate_zone": schema.BoolAttribute{
				MarkdownDescription: "Set to true to enable ZONEMD validation. Valid only for Secondary zones.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"initialize_forwarder": schema.BoolAttribute{
				MarkdownDescription: "Set to true to initialize the Conditional Forwarder zone with an FWD record. Valid for Forwarder zones.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "The DNS transport protocol to be used by the Conditional Forwarder zone. Valid values are: Udp, Tcp, Tls, Https, Quic.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Udp"),
				Validators: []validator.String{
					stringvalidator.OneOf("Udp", "Tcp", "Tls", "Https", "Quic"),
				},
			},
			"forwarder": schema.StringAttribute{
				MarkdownDescription: "The address of the DNS server to be used as a forwarder. Use 'this-server' to forward internally. Required for Conditional Forwarder zones.",
				Optional:            true,
			},
			"dnssec_validation": schema.BoolAttribute{
				MarkdownDescription: "Set to true to indicate if DNSSEC validation must be done. Used with Conditional Forwarder zones.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"proxy_type": schema.StringAttribute{
				MarkdownDescription: "The type of proxy for conditional forwarding. Valid values are: NoProxy, DefaultProxy, Http, Socks5.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("DefaultProxy"),
				Validators: []validator.String{
					stringvalidator.OneOf("NoProxy", "DefaultProxy", "Http", "Socks5"),
				},
			},
			"proxy_address": schema.StringAttribute{
				MarkdownDescription: "The proxy server address to use when proxy_type is configured.",
				Optional:            true,
			},
			"proxy_port": schema.Int64Attribute{
				MarkdownDescription: "The proxy server port to use when proxy_type is configured.",
				Optional:            true,
			},
			"proxy_username": schema.StringAttribute{
				MarkdownDescription: "The proxy server username to use when proxy_type is configured.",
				Optional:            true,
			},
			"proxy_password": schema.StringAttribute{
				MarkdownDescription: "The proxy server password to use when proxy_type is configured.",
				Optional:            true,
				Sensitive:           true,
			},

			// Computed attributes
			"internal": schema.BoolAttribute{
				MarkdownDescription: "Indicates if this is an internal zone.",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dnssec_status": schema.StringAttribute{
				MarkdownDescription: "The DNSSEC status of the zone.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Indicates if the zone is disabled.",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"soa_serial": schema.Int64Attribute{
				MarkdownDescription: "The SOA serial number of the zone.",
				Computed:            true,
			},
		},
	}
}

func (r *ZoneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ZoneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ZoneResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating zone", map[string]interface{}{
		"name": data.Name.ValueString(),
		"type": data.Type.ValueString(),
	})

	// Create zone using the API
	if err := r.createZone(ctx, &data); err != nil {
		resp.Diagnostics.AddError(
			"Error creating zone",
			fmt.Sprintf("Could not create zone %s: %s", data.Name.ValueString(), err.Error()),
		)
		return
	}

	// Set the ID for the resource (zone name serves as the ID)
	data.ID = data.Name

	// Read the zone back to get computed values
	if err := r.readZone(ctx, &data); err != nil {
		resp.Diagnostics.AddError(
			"Error reading zone after creation",
			fmt.Sprintf("Could not read zone %s after creation: %s", data.Name.ValueString(), err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Created zone successfully", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ZoneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ZoneResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read zone from API
	if err := r.readZone(ctx, &data); err != nil {
		if strings.Contains(err.Error(), "not found") {
			// Zone doesn't exist, remove from state
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading zone",
			fmt.Sprintf("Could not read zone %s: %s", data.Name.ValueString(), err.Error()),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ZoneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ZoneResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating zone", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	// Update zone options using the API
	if err := r.updateZone(ctx, &data); err != nil {
		resp.Diagnostics.AddError(
			"Error updating zone",
			fmt.Sprintf("Could not update zone %s: %s", data.Name.ValueString(), err.Error()),
		)
		return
	}

	// Read the zone back to get updated values
	if err := r.readZone(ctx, &data); err != nil {
		resp.Diagnostics.AddError(
			"Error reading zone after update",
			fmt.Sprintf("Could not read zone %s after update: %s", data.Name.ValueString(), err.Error()),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ZoneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ZoneResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting zone", map[string]interface{}{
		"name": data.Name.ValueString(),
	})

	// Delete zone using the API
	if err := r.deleteZone(ctx, data.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting zone",
			fmt.Sprintf("Could not delete zone %s: %s", data.Name.ValueString(), err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "Deleted zone successfully", map[string]interface{}{
		"name": data.Name.ValueString(),
	})
}

func (r *ZoneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Set both ID and name to the import ID (zone name)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), req.ID)...)
}

// createZone creates a new zone via the API
func (r *ZoneResource) createZone(ctx context.Context, data *ZoneResourceModel) error {
	params := url.Values{}
	params.Set("zone", data.Name.ValueString())
	params.Set("type", data.Type.ValueString())

	// Add optional parameters based on zone type and configuration
	if !data.Catalog.IsNull() && !data.Catalog.IsUnknown() {
		params.Set("catalog", data.Catalog.ValueString())
	}

	if !data.UseSoaSerialDateScheme.IsNull() && !data.UseSoaSerialDateScheme.IsUnknown() {
		params.Set("useSoaSerialDateScheme", fmt.Sprintf("%t", data.UseSoaSerialDateScheme.ValueBool()))
	}

	if !data.PrimaryNameServerAddresses.IsNull() && !data.PrimaryNameServerAddresses.IsUnknown() {
		params.Set("primaryNameServerAddresses", data.PrimaryNameServerAddresses.ValueString())
	}

	if !data.ZoneTransferProtocol.IsNull() && !data.ZoneTransferProtocol.IsUnknown() {
		params.Set("zoneTransferProtocol", data.ZoneTransferProtocol.ValueString())
	}

	if !data.TsigKeyName.IsNull() && !data.TsigKeyName.IsUnknown() {
		params.Set("tsigKeyName", data.TsigKeyName.ValueString())
	}

	if !data.ValidateZone.IsNull() && !data.ValidateZone.IsUnknown() {
		params.Set("validateZone", fmt.Sprintf("%t", data.ValidateZone.ValueBool()))
	}

	if !data.InitializeForwarder.IsNull() && !data.InitializeForwarder.IsUnknown() {
		params.Set("initializeForwarder", fmt.Sprintf("%t", data.InitializeForwarder.ValueBool()))
	}

	if !data.Protocol.IsNull() && !data.Protocol.IsUnknown() {
		params.Set("protocol", data.Protocol.ValueString())
	}

	if !data.Forwarder.IsNull() && !data.Forwarder.IsUnknown() {
		params.Set("forwarder", data.Forwarder.ValueString())
	}

	if !data.DnssecValidation.IsNull() && !data.DnssecValidation.IsUnknown() {
		params.Set("dnssecValidation", fmt.Sprintf("%t", data.DnssecValidation.ValueBool()))
	}

	if !data.ProxyType.IsNull() && !data.ProxyType.IsUnknown() {
		params.Set("proxyType", data.ProxyType.ValueString())
	}

	if !data.ProxyAddress.IsNull() && !data.ProxyAddress.IsUnknown() {
		params.Set("proxyAddress", data.ProxyAddress.ValueString())
	}

	if !data.ProxyPort.IsNull() && !data.ProxyPort.IsUnknown() {
		params.Set("proxyPort", fmt.Sprintf("%d", data.ProxyPort.ValueInt64()))
	}

	if !data.ProxyUsername.IsNull() && !data.ProxyUsername.IsUnknown() {
		params.Set("proxyUsername", data.ProxyUsername.ValueString())
	}

	if !data.ProxyPassword.IsNull() && !data.ProxyPassword.IsUnknown() {
		params.Set("proxyPassword", data.ProxyPassword.ValueString())
	}

	endpoint := "/api/zones/create?" + params.Encode()

	var response struct {
		Domain string `json:"domain"`
	}

	return r.client.DoRequest(ctx, "GET", endpoint, nil, &response)
}

// readZone reads zone information from the API
func (r *ZoneResource) readZone(ctx context.Context, data *ZoneResourceModel) error {
	// First, get the zone options
	params := url.Values{}
	params.Set("zone", data.Name.ValueString())
	endpoint := "/api/zones/options/get?" + params.Encode()

	var optionsResponse ZoneOptionsResponse
	if err := r.client.DoRequest(ctx, "GET", endpoint, nil, &optionsResponse); err != nil {
		return fmt.Errorf("failed to get zone options: %w", err)
	}

	// Ensure ID is set (zone name serves as the ID)
	data.ID = data.Name

	// Update the data model with the response
	data.Type = types.StringValue(optionsResponse.Type)
	data.Internal = types.BoolValue(optionsResponse.Internal)
	data.DnssecStatus = types.StringValue(optionsResponse.DnssecStatus)
	data.Disabled = types.BoolValue(optionsResponse.Disabled)

	// Set computed attributes - ensure all computed attributes get explicit values
	if optionsResponse.UseSoaSerialDateScheme != nil {
		data.UseSoaSerialDateScheme = types.BoolValue(*optionsResponse.UseSoaSerialDateScheme)
	} else {
		// API doesn't return this field, preserve the value from create operation
		// If this is a new resource without a state value, default to false
		if data.UseSoaSerialDateScheme.IsNull() || data.UseSoaSerialDateScheme.IsUnknown() {
			data.UseSoaSerialDateScheme = types.BoolValue(false)
		}
		// Otherwise, preserve the existing state value (don't modify it)
	}

	if optionsResponse.Catalog != "" {
		data.Catalog = types.StringValue(optionsResponse.Catalog)
	}

	if len(optionsResponse.PrimaryNameServerAddresses) > 0 {
		data.PrimaryNameServerAddresses = types.StringValue(strings.Join(optionsResponse.PrimaryNameServerAddresses, ","))
	}

	if optionsResponse.PrimaryZoneTransferProtocol != "" {
		data.ZoneTransferProtocol = types.StringValue(optionsResponse.PrimaryZoneTransferProtocol)
	} else {
		// Set default if not provided by API
		data.ZoneTransferProtocol = types.StringValue("Tcp")
	}

	if optionsResponse.PrimaryZoneTransferTsigKeyName != "" {
		data.TsigKeyName = types.StringValue(optionsResponse.PrimaryZoneTransferTsigKeyName)
	}

	if optionsResponse.ValidateZone != nil {
		data.ValidateZone = types.BoolValue(*optionsResponse.ValidateZone)
	} else {
		data.ValidateZone = types.BoolValue(false)
	}

	// Set computed attributes that need explicit defaults
	// Preserve InitializeForwarder value if already set, otherwise default to false
	if data.InitializeForwarder.IsNull() || data.InitializeForwarder.IsUnknown() {
		data.InitializeForwarder = types.BoolValue(false)
	}

	// Preserve DnssecValidation value if already set, otherwise default to false
	if data.DnssecValidation.IsNull() || data.DnssecValidation.IsUnknown() {
		data.DnssecValidation = types.BoolValue(false)
	}

	// Set default values for schema attributes with defaults
	data.Protocol = types.StringValue("Udp")
	data.ProxyType = types.StringValue("DefaultProxy")

	// Get zone records to extract SOA serial
	recordsParams := url.Values{}
	recordsParams.Set("domain", data.Name.ValueString())
	recordsParams.Set("zone", data.Name.ValueString())
	recordsParams.Set("listZone", "true")
	recordsEndpoint := "/api/zones/records/get?" + recordsParams.Encode()

	var recordsResponse ZoneRecordsResponse
	if err := r.client.DoRequest(ctx, "GET", recordsEndpoint, nil, &recordsResponse); err != nil {
		// Don't fail if records can't be read, just log it
		tflog.Warn(ctx, "Failed to read zone records for SOA serial", map[string]interface{}{
			"zone":  data.Name.ValueString(),
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

	return nil
}

// updateZone updates zone options via the API
func (r *ZoneResource) updateZone(ctx context.Context, data *ZoneResourceModel) error {
	params := url.Values{}
	params.Set("zone", data.Name.ValueString())

	// Add parameters that can be updated
	if !data.Catalog.IsNull() && !data.Catalog.IsUnknown() {
		params.Set("catalog", data.Catalog.ValueString())
	}

	// Note: useSoaSerialDateScheme cannot be updated after zone creation
	// This attribute requires zone replacement (handled by RequiresReplace plan modifier)

	if !data.PrimaryNameServerAddresses.IsNull() && !data.PrimaryNameServerAddresses.IsUnknown() {
		params.Set("primaryNameServerAddresses", data.PrimaryNameServerAddresses.ValueString())
	}

	if !data.ZoneTransferProtocol.IsNull() && !data.ZoneTransferProtocol.IsUnknown() {
		params.Set("primaryZoneTransferProtocol", data.ZoneTransferProtocol.ValueString())
	}

	if !data.TsigKeyName.IsNull() && !data.TsigKeyName.IsUnknown() {
		params.Set("primaryZoneTransferTsigKeyName", data.TsigKeyName.ValueString())
	}

	if !data.ValidateZone.IsNull() && !data.ValidateZone.IsUnknown() {
		params.Set("validateZone", fmt.Sprintf("%t", data.ValidateZone.ValueBool()))
	}

	endpoint := "/api/zones/options/set?" + params.Encode()

	return r.client.DoRequest(ctx, "GET", endpoint, nil, nil)
}

// deleteZone deletes a zone via the API
func (r *ZoneResource) deleteZone(ctx context.Context, zoneName string) error {
	params := url.Values{}
	params.Set("zone", zoneName)
	endpoint := "/api/zones/delete?" + params.Encode()
	return r.client.DoRequest(ctx, "GET", endpoint, nil, nil)
}

// ZoneOptionsResponse represents the API response for zone options
type ZoneOptionsResponse struct {
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

// ZoneRecordsResponse represents the API response for zone records
type ZoneRecordsResponse struct {
	Zone    ZoneInfo     `json:"zone"`
	Records []ZoneRecord `json:"records"`
}

type ZoneInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Internal     bool   `json:"internal"`
	DnssecStatus string `json:"dnssecStatus"`
	Disabled     bool   `json:"disabled"`
}

type ZoneRecord struct {
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	TTL      int             `json:"ttl"`
	RData    ZoneRecordRData `json:"rData"`
	Disabled bool            `json:"disabled"`
}

type ZoneRecordRData struct {
	SoaRecord *SoaRecordData `json:"soaRecord,omitempty"`
}

type SoaRecordData struct {
	Serial uint32 `json:"serial"`
}
