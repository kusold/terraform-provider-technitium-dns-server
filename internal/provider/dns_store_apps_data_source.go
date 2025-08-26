package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &DNSStoreAppsDataSource{}

func NewDNSStoreAppsDataSource() datasource.DataSource {
	return &DNSStoreAppsDataSource{}
}

// DNSStoreAppsDataSource defines the data source implementation.
type DNSStoreAppsDataSource struct {
	client *client.Client
}

// DNSStoreAppsDataSourceModel describes the data source data model.
type DNSStoreAppsDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	StoreApps types.List   `tfsdk:"store_apps"`
}

// DNSStoreAppDataItem represents an individual store app for the data source
type DNSStoreAppDataItem struct {
	Name             types.String `tfsdk:"name"`
	Version          types.String `tfsdk:"version"`
	Description      types.String `tfsdk:"description"`
	URL              types.String `tfsdk:"url"`
	Size             types.String `tfsdk:"size"`
	Installed        types.Bool   `tfsdk:"installed"`
	InstalledVersion types.String `tfsdk:"installed_version"`
	UpdateAvailable  types.Bool   `tfsdk:"update_available"`
}

func (d *DNSStoreAppsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_store_apps"
}

func (d *DNSStoreAppsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Data source to retrieve all available DNS applications from the Technitium DNS App Store",
		MarkdownDescription: "Data source to retrieve all available DNS applications from the Technitium DNS App Store",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the data source.",
				Computed:            true,
			},
			"store_apps": schema.ListNestedAttribute{
				MarkdownDescription: "List of DNS applications available in the store.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the DNS application.",
							Computed:            true,
						},
						"version": schema.StringAttribute{
							MarkdownDescription: "Version of the DNS application.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description of the DNS application.",
							Computed:            true,
						},
						"url": schema.StringAttribute{
							MarkdownDescription: "Download URL for the DNS application.",
							Computed:            true,
						},
						"size": schema.StringAttribute{
							MarkdownDescription: "Size of the application package.",
							Computed:            true,
						},
						"installed": schema.BoolAttribute{
							MarkdownDescription: "Whether the application is currently installed.",
							Computed:            true,
						},
						"installed_version": schema.StringAttribute{
							MarkdownDescription: "Version of the currently installed application (if installed).",
							Computed:            true,
						},
						"update_available": schema.BoolAttribute{
							MarkdownDescription: "Whether an update is available for the installed application.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *DNSStoreAppsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DNSStoreAppsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DNSStoreAppsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading DNS store apps")

	// Get store apps from the API
	storeApps, err := d.client.ListStoreApps(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read DNS store apps: %s", err.Error()))
		return
	}

	tflog.Debug(ctx, "Retrieved DNS store apps", map[string]interface{}{
		"store_app_count": len(storeApps),
	})

	// Convert store apps to Terraform format
	storeAppElements := make([]attr.Value, 0, len(storeApps))
	for _, storeApp := range storeApps {
		installedVersion := types.StringNull()
		if storeApp.InstalledVersion != "" {
			installedVersion = types.StringValue(storeApp.InstalledVersion)
		}

		storeAppObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"name":              types.StringType,
				"version":           types.StringType,
				"description":       types.StringType,
				"url":               types.StringType,
				"size":              types.StringType,
				"installed":         types.BoolType,
				"installed_version": types.StringType,
				"update_available":  types.BoolType,
			},
			map[string]attr.Value{
				"name":              types.StringValue(storeApp.Name),
				"version":           types.StringValue(storeApp.Version),
				"description":       types.StringValue(storeApp.Description),
				"url":               types.StringValue(storeApp.URL),
				"size":              types.StringValue(storeApp.Size),
				"installed":         types.BoolValue(storeApp.Installed),
				"installed_version": installedVersion,
				"update_available":  types.BoolValue(storeApp.UpdateAvailable),
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		storeAppElements = append(storeAppElements, storeAppObj)
	}

	// Create the store apps list
	storeAppsList, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":              types.StringType,
				"version":           types.StringType,
				"description":       types.StringType,
				"url":               types.StringType,
				"size":              types.StringType,
				"installed":         types.BoolType,
				"installed_version": types.StringType,
				"update_available":  types.BoolType,
			},
		},
		storeAppElements,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the results
	data.ID = types.StringValue("dns_store_apps")
	data.StoreApps = storeAppsList

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
