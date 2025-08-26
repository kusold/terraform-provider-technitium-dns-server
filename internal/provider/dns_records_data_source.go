package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &DNSRecordsDataSource{}

func NewDNSRecordsDataSource() datasource.DataSource {
	return &DNSRecordsDataSource{}
}

// DNSRecordsDataSource defines the data source implementation.
type DNSRecordsDataSource struct {
	client *client.Client
}

// DNSRecordsDataSourceModel describes the data source data model.
type DNSRecordsDataSourceModel struct {
	// Required inputs
	Zone types.String `tfsdk:"zone"`

	// Optional inputs
	Domain      types.String   `tfsdk:"domain"`
	RecordTypes []types.String `tfsdk:"record_types"`

	// Computed outputs
	ID      types.String        `tfsdk:"id"`
	Records []DNSRecordDataItem `tfsdk:"records"`
}

// DNSRecordDataItem represents an individual DNS record
type DNSRecordDataItem struct {
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	TTL      types.Int64  `tfsdk:"ttl"`
	Data     types.String `tfsdk:"data"`
	Disabled types.Bool   `tfsdk:"disabled"`
	Comments types.String `tfsdk:"comments"`
}

func (d *DNSRecordsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_records"
}

func (d *DNSRecordsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Data source to retrieve and filter DNS records from a Technitium DNS zone",
		MarkdownDescription: "Data source to retrieve and filter DNS records from a Technitium DNS zone",

		Attributes: map[string]schema.Attribute{
			// Required inputs
			"zone": schema.StringAttribute{
				MarkdownDescription: "The zone name to retrieve DNS records from (e.g., 'example.com').",
				Required:            true,
			},

			// Optional inputs
			"domain": schema.StringAttribute{
				MarkdownDescription: "The specific domain to retrieve records for. If not specified, all records in the zone will be returned.",
				Optional:            true,
			},
			"record_types": schema.ListAttribute{
				MarkdownDescription: "Filter records by type (e.g., ['A', 'AAAA', 'CNAME']). If not specified, all record types will be returned.",
				Optional:            true,
				ElementType:         types.StringType,
			},

			// Computed outputs
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the data source.",
				Computed:            true,
			},
			"records": schema.ListNestedAttribute{
				MarkdownDescription: "List of DNS records in the zone.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The DNS record name.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The DNS record type (A, AAAA, CNAME, MX, TXT, etc.).",
							Computed:            true,
						},
						"ttl": schema.Int64Attribute{
							MarkdownDescription: "Time-to-live value for the record in seconds.",
							Computed:            true,
						},
						"data": schema.StringAttribute{
							MarkdownDescription: "The record data, formatted according to the record type.",
							Computed:            true,
						},
						"disabled": schema.BoolAttribute{
							MarkdownDescription: "Whether the record is disabled.",
							Computed:            true,
						},
						"comments": schema.StringAttribute{
							MarkdownDescription: "Any comments attached to the record.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *DNSRecordsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DNSRecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DNSRecordsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	zoneName := data.Zone.ValueString()
	domain := zoneName // Default to the zone name

	if !data.Domain.IsNull() {
		domain = data.Domain.ValueString()
	}

	// Determine if we need to list all records in the zone
	listZone := (domain == zoneName)

	tflog.Debug(ctx, "Reading DNS records data source", map[string]interface{}{
		"zone":     zoneName,
		"domain":   domain,
		"listZone": listZone,
	})

	// Get DNS records from the API
	recordsResponse, err := d.client.GetRecords(ctx, zoneName, domain, listZone)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading DNS records",
			fmt.Sprintf("Could not read DNS records for zone %s: %s", zoneName, err.Error()),
		)
		return
	}

	// Create a set to check if a record type should be included
	includeRecordTypes := make(map[string]bool)
	if len(data.RecordTypes) > 0 {
		for _, recordType := range data.RecordTypes {
			includeRecordTypes[recordType.ValueString()] = true
		}
	}

	// Process records and convert to Terraform model
	records := make([]DNSRecordDataItem, 0)
	for _, record := range recordsResponse.Records {
		// Skip record if type filtering is enabled and this type isn't in the filter
		if len(includeRecordTypes) > 0 && !includeRecordTypes[record.Type] {
			continue
		}

		// Format record data based on the record type
		formattedData := formatRecordData(record)

		recordItem := DNSRecordDataItem{
			Name:     types.StringValue(record.Name),
			Type:     types.StringValue(record.Type),
			TTL:      types.Int64Value(int64(record.TTL)),
			Data:     types.StringValue(formattedData),
			Disabled: types.BoolValue(record.Disabled),
			Comments: types.StringValue(record.Comments),
		}

		records = append(records, recordItem)
	}

	data.ID = types.StringValue(zoneName)
	data.Records = records

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// formatRecordData formats the record data based on the record type
func formatRecordData(record client.DNSRecord) string {
	switch record.Type {
	case "A", "AAAA":
		return record.RData.IPAddress
	case "CNAME":
		return record.RData.CNAME
	case "MX":
		return fmt.Sprintf("%d %s", record.RData.Preference, record.RData.Exchange)
	case "TXT":
		return record.RData.Text
	case "PTR":
		return record.RData.PTRName
	case "NS":
		return record.RData.NameServer
	case "SRV":
		return fmt.Sprintf("%d %d %d %s", record.RData.Priority, record.RData.Weight, record.RData.Port, record.RData.Target)
	case "SOA":
		return fmt.Sprintf("%s %s %d %d %d %d %d",
			record.RData.PrimaryNameServer,
			record.RData.ResponsiblePerson,
			record.RData.Serial,
			record.RData.Refresh,
			record.RData.Retry,
			record.RData.Expire,
			record.RData.Minimum)
	default:
		// For other record types, return an empty string as they have complex structures
		return fmt.Sprintf("[%s record]", record.Type)
	}
}
