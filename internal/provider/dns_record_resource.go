package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DNSRecordResource{}
var _ resource.ResourceWithImportState = &DNSRecordResource{}

func NewDNSRecordResource() resource.Resource {
	return &DNSRecordResource{}
}

// DNSRecordResource defines the resource implementation.
type DNSRecordResource struct {
	client *client.Client
}

// DNSRecordResourceModel describes the resource data model.
type DNSRecordResourceModel struct {
	ID       types.String `tfsdk:"id"`
	Zone     types.String `tfsdk:"zone"`
	Name     types.String `tfsdk:"name"`
	Type     types.String `tfsdk:"type"`
	TTL      types.Int64  `tfsdk:"ttl"`
	Data     types.String `tfsdk:"data"`     // Holds the main record data (varies by type)
	Priority types.Int64  `tfsdk:"priority"` // For MX and SRV records
	Weight   types.Int64  `tfsdk:"weight"`   // For SRV records
	Port     types.Int64  `tfsdk:"port"`     // For SRV records
	Comments types.String `tfsdk:"comments"` // Optional comments

	// Computed attributes
	Disabled     types.Bool   `tfsdk:"disabled"`
	DnssecStatus types.String `tfsdk:"dnssec_status"`
	LastUsedOn   types.String `tfsdk:"last_used_on"`
}

func (r *DNSRecordResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (r *DNSRecordResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Technitium DNS Server record resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"zone": schema.StringAttribute{
				MarkdownDescription: "The zone in which to create the DNS record",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The record name (e.g., 'www' for www.example.com)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The DNS record type (A, AAAA, CNAME, MX, TXT, etc.)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(
						"A", "AAAA", "CNAME", "MX", "TXT",
						"PTR", "NS", "SRV",
					),
				},
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "Time-to-live value in seconds",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"data": schema.StringAttribute{
				MarkdownDescription: "Record data (depends on record type: IP address for A/AAAA, domain for CNAME, text for TXT, etc.)",
				Required:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Priority value (used for MX and SRV records)",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"weight": schema.Int64Attribute{
				MarkdownDescription: "Weight value (used for SRV records)",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Port value (used for SRV records)",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"comments": schema.StringAttribute{
				MarkdownDescription: "Optional comments for the DNS record",
				Optional:            true,
			},

			// Computed attributes
			"disabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the record is disabled",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dnssec_status": schema.StringAttribute{
				MarkdownDescription: "DNSSEC status of the record",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_used_on": schema.StringAttribute{
				MarkdownDescription: "When the record was last used",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DNSRecordResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSRecordResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create options map for record creation
	options := r.buildRecordOptions(ctx, &data, "create")

	// Validate based on record type
	if err := r.validateRecord(&data, options); err != nil {
		resp.Diagnostics.AddError(
			"Invalid DNS record configuration",
			err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "Creating DNS record", map[string]interface{}{
		"zone": data.Zone.ValueString(),
		"name": data.Name.ValueString(),
		"type": data.Type.ValueString(),
	})

	// In Technitium DNS, if the record name doesn't match certain patterns,
	// we need to use the fully qualified domain name (FQDN)
	recordName := data.Name.ValueString()
	zoneName := data.Zone.ValueString()

	// If the record name is not "@" (root), not already the zone name, and doesn't end with the zone name,
	// we need to append the zone name to create a proper FQDN for Technitium
	if recordName != "@" && recordName != zoneName {
		// For short names like "www", we need to append the zone name to make "www.example.com"
		// But don't do this if it already has a trailing dot or already includes the zone name
		if !strings.HasSuffix(recordName, ".") && !strings.HasSuffix(recordName, "."+zoneName) && !strings.HasSuffix(recordName, zoneName) {
			recordName = recordName + "." + zoneName
		}
	}

	tflog.Debug(ctx, "Creating DNS record with formatted name", map[string]interface{}{
		"zone":           zoneName,
		"original_name":  data.Name.ValueString(),
		"formatted_name": recordName,
	})

	// Create the record via the API
	recordResp, err := r.client.AddRecord(
		ctx,
		zoneName,
		recordName,
		data.Type.ValueString(),
		int(data.TTL.ValueInt64()),
		options,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating DNS record",
			fmt.Sprintf("Could not create %s record %s: %s", data.Type.ValueString(), data.Name.ValueString(), err.Error()),
		)
		return
	}

	// Generate a unique ID for the record
	recordID := fmt.Sprintf("%s:%s:%s",
		data.Zone.ValueString(),
		data.Name.ValueString(),
		data.Type.ValueString(),
	)

	// For records like MX and SRV that need additional data in the ID to be unique
	if !data.Priority.IsNull() && !data.Priority.IsUnknown() {
		recordID += fmt.Sprintf(":%d", data.Priority.ValueInt64())
	}

	// For data value that might be crucial to uniquely identify the record
	// Special handling for TXT records
	if data.Type.ValueString() == "TXT" {
		// TXT records may have spaces and special characters that make IDs problematic
		// For TXT records, exclude the data from the ID to avoid issues with special characters
		// The combination of zone, name, and type should be unique enough

		// Log TXT record ID generation without including the data
		tflog.Info(ctx, "Generated TXT record ID without data field", map[string]interface{}{
			"record_id": recordID,
			"txt_value": data.Data.ValueString(),
		})
	} else if data.Data.ValueString() != "" {
		// For other record types, include the data in the ID
		recordID += fmt.Sprintf(":%s", data.Data.ValueString())
	}

	data.ID = types.StringValue(recordID)

	// Update model with any computed fields from response
	data.Disabled = types.BoolValue(recordResp.AddedRecord.Disabled)
	data.DnssecStatus = types.StringValue(recordResp.AddedRecord.DnssecStatus)

	// Set default values for computed fields
	if data.Priority.IsNull() || data.Priority.IsUnknown() {
		data.Priority = types.Int64Value(0)
	}

	if data.Weight.IsNull() || data.Weight.IsUnknown() {
		data.Weight = types.Int64Value(0)
	}

	if data.Port.IsNull() || data.Port.IsUnknown() {
		data.Port = types.Int64Value(0)
	}

	if recordResp.AddedRecord.LastUsedOn != "" {
		data.LastUsedOn = types.StringValue(recordResp.AddedRecord.LastUsedOn)
	} else {
		data.LastUsedOn = types.StringValue("")
	}

	tflog.Debug(ctx, "DNS record created successfully", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSRecordResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract record details from ID (format: zone:name:type[:priority][:data])
	idParts := strings.Split(data.ID.ValueString(), ":")
	if len(idParts) < 3 {
		resp.Diagnostics.AddError(
			"Invalid ID format",
			fmt.Sprintf("Expected at least 3 parts in ID (zone:name:type), got: %s", data.ID.ValueString()),
		)
		return
	}

	zone := idParts[0]
	name := idParts[1]
	recordType := idParts[2]

	// Add extra logging for TXT records
	if recordType == "TXT" {
		tflog.Info(ctx, "Reading TXT record in Read method", map[string]interface{}{
			"id":          data.ID.ValueString(),
			"zone":        zone,
			"name":        name,
			"type":        recordType,
			"parts_count": len(idParts),
		})

		// For TXT records, we don't include data in the ID, so we need to be careful
		// when matching records. We'll primarily match on zone, name, and type.
	}


	// Format the name properly for Technitium DNS
	recordName := name
	zoneName := zone

	// If the record name is not "@" (root), not already the zone name, and doesn't end with the zone name,
	// we need to append the zone name to create a proper FQDN for Technitium
	if recordName != "@" && recordName != zoneName {
		// For short names like "www", we need to append the zone name to make "www.example.com"
		// But don't do this if it already has a trailing dot or already includes the zone name
		if !strings.HasSuffix(recordName, ".") && !strings.HasSuffix(recordName, "."+zoneName) && !strings.HasSuffix(recordName, zoneName) {
			recordName = recordName + "." + zoneName
		}
	}

	// Priority or data may be part of the ID for certain record types
	var priority int64
	var recordData string

	if len(idParts) > 3 {
		// Try to parse as priority first
		if p, err := strconv.ParseInt(idParts[3], 10, 64); err == nil {
			priority = p
		} else {
			recordData = idParts[3]
		}
	}

	if len(idParts) > 4 {
		recordData = idParts[4]
	}

	// Fetch records for this domain in this zone
	recordsResp, err := r.client.GetRecords(ctx, zone, recordName, false)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading DNS record",
			fmt.Sprintf("Could not read %s record %s in zone %s: %s", recordType, recordName, zone, err.Error()),
		)
		return
	}

	// Debug log for TXT records
	if recordType == "TXT" {
		tflog.Debug(ctx, "Reading TXT record details", map[string]interface{}{
			"zone":        zone,
			"name":        name,
			"recordName":  recordName,
			"recordData":  recordData,
			"recordCount": len(recordsResp.Records),
		})

		for i, record := range recordsResp.Records {
			tflog.Debug(ctx, fmt.Sprintf("Record %d details", i), map[string]interface{}{
				"type":     record.Type,
				"name":     record.Name,
				"ttl":      record.TTL,
				"text":     record.RData.Text,
				"disabled": record.Disabled,
			})
		}
	}

	// Find the specific record we're looking for
	var found bool
	for _, record := range recordsResp.Records {
		// Match on type first
		if record.Type != recordType {
			continue
		}

		// For MX records, match on priority and data
		if recordType == "MX" {
			if (priority > 0 && record.RData.Preference != int(priority)) ||
				(recordData != "" && record.RData.Exchange != recordData) {
				continue
			}
		} else if recordType == "A" || recordType == "AAAA" {
			if recordData != "" && record.RData.IPAddress != recordData {
				continue
			}
		} else if recordType == "CNAME" {
			if recordData != "" && record.RData.CNAME != recordData {
				continue
			}
		} else if recordType == "TXT" {
			// Debug log for TXT record comparison
			tflog.Debug(ctx, "TXT record comparison in Read", map[string]interface{}{
				"expected":  recordData,
				"actual":    record.RData.Text,
				"match":     record.RData.Text == recordData,
				"record_id": data.ID.ValueString(),
			})

			// Special handling for TXT records - they might have quotes or special handling
			if recordData != "" {
				// Try both with and without quotes for matching
				cleanExpected := strings.Trim(recordData, "\"")
				cleanActual := strings.Trim(record.RData.Text, "\"")

				tflog.Debug(ctx, "TXT record cleaned comparison", map[string]interface{}{
					"clean_expected": cleanExpected,
					"clean_actual":   cleanActual,
					"clean_match":    cleanExpected == cleanActual,
				})

				// Skip only if neither raw nor cleaned comparison matches
				if record.RData.Text != recordData && cleanActual != cleanExpected {
					continue
				}
			}
		}

		// If we reach here, we've found a match
		found = true

		// Update the model with values from the record
		data.Zone = types.StringValue(zone)
		data.Name = types.StringValue(name)
		data.Type = types.StringValue(recordType)
		data.TTL = types.Int64Value(int64(record.TTL))
		data.Disabled = types.BoolValue(record.Disabled)
		data.DnssecStatus = types.StringValue(record.DnssecStatus)

		// Set default values for computed fields
		if data.Priority.IsNull() || data.Priority.IsUnknown() {
			data.Priority = types.Int64Value(0)
		}

		if data.Weight.IsNull() || data.Weight.IsUnknown() {
			data.Weight = types.Int64Value(0)
		}

		if data.Port.IsNull() || data.Port.IsUnknown() {
			data.Port = types.Int64Value(0)
		}

		if record.LastUsedOn != "" {
			data.LastUsedOn = types.StringValue(record.LastUsedOn)
		} else {
			data.LastUsedOn = types.StringValue("")
		}

		// Set record-specific fields
		switch recordType {
		case "A", "AAAA":
			data.Data = types.StringValue(record.RData.IPAddress)
		case "CNAME":
			data.Data = types.StringValue(record.RData.CNAME)
		case "MX":
			data.Data = types.StringValue(record.RData.Exchange)
			data.Priority = types.Int64Value(int64(record.RData.Preference))
		case "TXT":
			// Special handling for TXT record data
			txtValue := record.RData.Text

			// Log the raw value received from the API
			tflog.Debug(ctx, "TXT record data from API", map[string]interface{}{
				"raw_value": txtValue,
			})

			// Remove quotes if they're present
			txtValue = strings.Trim(txtValue, "\"")

			data.Data = types.StringValue(txtValue)
		case "PTR":
			data.Data = types.StringValue(record.RData.PTRName)
		case "NS":
			data.Data = types.StringValue(record.RData.NameServer)
		case "SRV":
			data.Data = types.StringValue(record.RData.Target)
			data.Priority = types.Int64Value(int64(record.RData.Priority))
			data.Weight = types.Int64Value(int64(record.RData.Weight))
			data.Port = types.Int64Value(int64(record.RData.Port))
		}

		break
	}

	if !found {
		// Record not found, remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSRecordResourceModel
	var oldData DNSRecordResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &oldData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create options map for record update
	options := r.buildRecordOptions(ctx, &oldData, "current")
	updateOptions := r.buildRecordOptions(ctx, &data, "new")

	// Merge options for the update call
	for k, v := range updateOptions {
		options[k] = v
	}

	// Add TTL to options
	options["ttl"] = strconv.FormatInt(data.TTL.ValueInt64(), 10)

	// Add comments if provided
	if !data.Comments.IsNull() && !data.Comments.IsUnknown() {
		options["comments"] = data.Comments.ValueString()
	}

	// Format the name properly for Technitium DNS
	recordName := data.Name.ValueString()
	zoneName := data.Zone.ValueString()

	// If the record name is not "@" (root), not already the zone name, and doesn't end with the zone name,
	// we need to append the zone name to create a proper FQDN for Technitium
	if recordName != "@" && recordName != zoneName {
		// For short names like "www", we need to append the zone name to make "www.example.com"
		// But don't do this if it already has a trailing dot or already includes the zone name
		if !strings.HasSuffix(recordName, ".") && !strings.HasSuffix(recordName, "."+zoneName) && !strings.HasSuffix(recordName, zoneName) {
			recordName = recordName + "." + zoneName
		}
	}

	tflog.Debug(ctx, "Updating DNS record", map[string]interface{}{
		"id":             data.ID.ValueString(),
		"zone":           zoneName,
		"original_name":  data.Name.ValueString(),
		"formatted_name": recordName,
		"type":           data.Type.ValueString(),
	})

	// Update the record via the API
	recordResp, err := r.client.UpdateRecord(
		ctx,
		zoneName,
		recordName,
		data.Type.ValueString(),
		options,
	)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating DNS record",
			fmt.Sprintf("Could not update %s record %s: %s", data.Type.ValueString(), data.Name.ValueString(), err.Error()),
		)
		return
	}

	// Update model with any computed fields from response
	data.Disabled = types.BoolValue(recordResp.UpdatedRecord.Disabled)
	data.DnssecStatus = types.StringValue(recordResp.UpdatedRecord.DnssecStatus)

	// Set default values for computed fields
	if data.Priority.IsNull() || data.Priority.IsUnknown() {
		data.Priority = types.Int64Value(0)
	}

	if data.Weight.IsNull() || data.Weight.IsUnknown() {
		data.Weight = types.Int64Value(0)
	}

	if data.Port.IsNull() || data.Port.IsUnknown() {
		data.Port = types.Int64Value(0)
	}

	if recordResp.UpdatedRecord.LastUsedOn != "" {
		data.LastUsedOn = types.StringValue(recordResp.UpdatedRecord.LastUsedOn)
	} else {
		data.LastUsedOn = types.StringValue("")
	}

	tflog.Debug(ctx, "DNS record updated successfully", map[string]interface{}{
		"id": data.ID.ValueString(),
	})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSRecordResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create options map for record deletion
	options := r.buildRecordOptions(ctx, &data, "delete")

	// Format the name properly for Technitium DNS
	recordName := data.Name.ValueString()
	zoneName := data.Zone.ValueString()

	// If the record name is not "@" (root), not already the zone name, and doesn't end with the zone name,
	// we need to append the zone name to create a proper FQDN for Technitium
	if recordName != "@" && recordName != zoneName {
		// For short names like "www", we need to append the zone name to make "www.example.com"
		// But don't do this if it already has a trailing dot or already includes the zone name
		if !strings.HasSuffix(recordName, ".") && !strings.HasSuffix(recordName, "."+zoneName) && !strings.HasSuffix(recordName, zoneName) {
			recordName = recordName + "." + zoneName
		}
	}

	tflog.Debug(ctx, "Deleting DNS record", map[string]interface{}{
		"id":             data.ID.ValueString(),
		"zone":           zoneName,
		"original_name":  data.Name.ValueString(),
		"formatted_name": recordName,
		"type":           data.Type.ValueString(),
	})

	// Delete the record via the API
	if err := r.client.DeleteRecord(
		ctx,
		zoneName,
		recordName,
		data.Type.ValueString(),
		options,
	); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting DNS record",
			fmt.Sprintf("Could not delete %s record %s: %s", data.Type.ValueString(), data.Name.ValueString(), err.Error()),
		)
		return
	}

	tflog.Debug(ctx, "DNS record deleted successfully", map[string]interface{}{
		"id": data.ID.ValueString(),
	})
}

func (r *DNSRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: zone:name:type[:priority][:data]
	idParts := strings.Split(req.ID, ":")
	if len(idParts) < 3 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			"Import ID must be in the format zone:name:type or zone:name:type:priority:data",
		)
		return
	}

	// Set ID and core attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("zone"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), idParts[2])...)

	// For MX records, priority and data may be included
	if len(idParts) > 3 {
		// Try to parse as priority first
		if priority, err := strconv.ParseInt(idParts[3], 10, 64); err == nil {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("priority"), priority)...)
		} else {
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("data"), idParts[3])...)
		}
	}

	if len(idParts) > 4 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("data"), idParts[4])...)
	}
}

// buildRecordOptions creates a map of options based on record type for API calls
func (r *DNSRecordResource) buildRecordOptions(ctx context.Context, data *DNSRecordResourceModel, opType string) map[string]string {
	options := make(map[string]string)

	// Different operation types need different parameter names
	recordType := data.Type.ValueString()

	switch recordType {
	case "A", "AAAA":
		paramName := "ipAddress"
		if opType == "new" {
			paramName = "newIpAddress"
		}
		options[paramName] = data.Data.ValueString()

	case "CNAME":
		paramName := "cname"
		if opType == "new" {
			paramName = "newCname"
		}
		options[paramName] = data.Data.ValueString()

	case "MX":
		exchangeParam := "exchange"
		preferenceParam := "preference"

		if opType == "new" {
			exchangeParam = "newExchange"
			preferenceParam = "newPreference"
		}

		options[exchangeParam] = data.Data.ValueString()

		if !data.Priority.IsNull() && !data.Priority.IsUnknown() {
			options[preferenceParam] = strconv.FormatInt(data.Priority.ValueInt64(), 10)
		}

	case "TXT":
		textParam := "text"
		if opType == "new" {
			textParam = "newText"
		}

		// Handle TXT record special formatting
		txtValue := data.Data.ValueString()

		// Remove quotes if already present in the string, Technitium API will add them if needed
		txtValue = strings.Trim(txtValue, "\"")

		options[textParam] = txtValue

	case "PTR":
		ptrParam := "ptrName"
		if opType == "new" {
			ptrParam = "newPtrName"
		}
		options[ptrParam] = data.Data.ValueString()

	case "NS":
		nsParam := "nameServer"
		if opType == "new" {
			nsParam = "newNameServer"
		}
		options[nsParam] = data.Data.ValueString()

	case "SRV":
		targetParam := "target"
		priorityParam := "priority"
		weightParam := "weight"
		portParam := "port"

		if opType == "new" {
			targetParam = "newTarget"
			priorityParam = "newPriority"
			weightParam = "newWeight"
			portParam = "newPort"
		}

		options[targetParam] = data.Data.ValueString()

		if !data.Priority.IsNull() && !data.Priority.IsUnknown() {
			options[priorityParam] = strconv.FormatInt(data.Priority.ValueInt64(), 10)
		}

		if !data.Weight.IsNull() && !data.Weight.IsUnknown() {
			options[weightParam] = strconv.FormatInt(data.Weight.ValueInt64(), 10)
		}

		if !data.Port.IsNull() && !data.Port.IsUnknown() {
			options[portParam] = strconv.FormatInt(data.Port.ValueInt64(), 10)
		}
	}

	// Add comments for create and update operations
	if (opType == "create" || opType == "new") && !data.Comments.IsNull() && !data.Comments.IsUnknown() {
		options["comments"] = data.Comments.ValueString()
	}

	return options
}

// validateRecord performs validation based on record type
func (r *DNSRecordResource) validateRecord(data *DNSRecordResourceModel, options map[string]string) error {
	recordType := data.Type.ValueString()

	switch recordType {
	case "A":
		// Validate IPv4 address format - basic validation only
		if !strings.Contains(data.Data.ValueString(), ".") {
			return fmt.Errorf("invalid IPv4 address format for A record: %s", data.Data.ValueString())
		}

	case "AAAA":
		// Validate IPv6 address format - basic validation only
		if !strings.Contains(data.Data.ValueString(), ":") {
			return fmt.Errorf("invalid IPv6 address format for AAAA record: %s", data.Data.ValueString())
		}

	case "MX":
		// Ensure priority is set for MX records
		if data.Priority.IsNull() || data.Priority.IsUnknown() {
			return fmt.Errorf("priority is required for MX records")
		}

	case "SRV":
		// Ensure all required fields are set for SRV records
		if data.Priority.IsNull() || data.Priority.IsUnknown() {
			return fmt.Errorf("priority is required for SRV records")
		}

		if data.Weight.IsNull() || data.Weight.IsUnknown() {
			return fmt.Errorf("weight is required for SRV records")
		}

		if data.Port.IsNull() || data.Port.IsUnknown() {
			return fmt.Errorf("port is required for SRV records")
		}
	}

	return nil
}
