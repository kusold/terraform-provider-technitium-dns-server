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
var _ datasource.DataSource = &DNSAppsDataSource{}

func NewDNSAppsDataSource() datasource.DataSource {
	return &DNSAppsDataSource{}
}

// DNSAppsDataSource defines the data source implementation.
type DNSAppsDataSource struct {
	client *client.Client
}

// DNSAppsDataSourceModel describes the data source data model.
type DNSAppsDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Apps types.List   `tfsdk:"apps"`
}

// DNSAppDataItem represents an individual DNS app for the data source
type DNSAppDataItem struct {
	Name    types.String `tfsdk:"name"`
	Version types.String `tfsdk:"version"`
	DNSApps types.List   `tfsdk:"dns_apps"`
}

func (d *DNSAppsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_apps"
}

func (d *DNSAppsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source to retrieve all installed DNS applications from a Technitium DNS Server",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the data source.",
				Computed:            true,
			},
			"apps": schema.ListNestedAttribute{
				MarkdownDescription: "List of installed DNS applications.",
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
						"dns_apps": schema.ListNestedAttribute{
							MarkdownDescription: "List of DNS application components within this app package.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"class_path": schema.StringAttribute{
										MarkdownDescription: "Class path of the DNS application.",
										Computed:            true,
									},
									"description": schema.StringAttribute{
										MarkdownDescription: "Description of the DNS application.",
										Computed:            true,
									},
									"is_app_record_request_handler": schema.BoolAttribute{
										MarkdownDescription: "Whether this app handles APP record requests.",
										Computed:            true,
									},
									"record_data_template": schema.StringAttribute{
										MarkdownDescription: "Record data template for APP records.",
										Computed:            true,
									},
									"is_request_controller": schema.BoolAttribute{
										MarkdownDescription: "Whether this app is a request controller.",
										Computed:            true,
									},
									"is_authoritative_request_handler": schema.BoolAttribute{
										MarkdownDescription: "Whether this app handles authoritative requests.",
										Computed:            true,
									},
									"is_request_blocking_handler": schema.BoolAttribute{
										MarkdownDescription: "Whether this app handles request blocking.",
										Computed:            true,
									},
									"is_query_logger": schema.BoolAttribute{
										MarkdownDescription: "Whether this app is a query logger.",
										Computed:            true,
									},
									"is_post_processor": schema.BoolAttribute{
										MarkdownDescription: "Whether this app is a post processor.",
										Computed:            true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *DNSAppsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DNSAppsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DNSAppsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading DNS apps")

	// Get installed apps from the API
	apps, err := d.client.ListApps(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read DNS apps: %s", err.Error()))
		return
	}

	tflog.Debug(ctx, "Retrieved DNS apps", map[string]interface{}{
		"app_count": len(apps),
	})

	// Convert apps to Terraform format
	appElements := make([]attr.Value, 0, len(apps))
	for _, app := range apps {
		// Convert DNS apps for this app
		dnsAppElements := make([]attr.Value, 0, len(app.DNSApps))
		for _, dnsApp := range app.DNSApps {
			recordDataTemplate := types.StringNull()
			if dnsApp.RecordDataTemplate != nil {
				recordDataTemplate = types.StringValue(*dnsApp.RecordDataTemplate)
			}

			dnsAppObj, diags := types.ObjectValue(
				map[string]attr.Type{
					"class_path":                       types.StringType,
					"description":                      types.StringType,
					"is_app_record_request_handler":    types.BoolType,
					"record_data_template":             types.StringType,
					"is_request_controller":            types.BoolType,
					"is_authoritative_request_handler": types.BoolType,
					"is_request_blocking_handler":      types.BoolType,
					"is_query_logger":                  types.BoolType,
					"is_post_processor":                types.BoolType,
				},
				map[string]attr.Value{
					"class_path":                       types.StringValue(dnsApp.ClassPath),
					"description":                      types.StringValue(dnsApp.Description),
					"is_app_record_request_handler":    types.BoolValue(dnsApp.IsAppRecordRequestHandler),
					"record_data_template":             recordDataTemplate,
					"is_request_controller":            types.BoolValue(dnsApp.IsRequestController),
					"is_authoritative_request_handler": types.BoolValue(dnsApp.IsAuthoritativeRequestHandler),
					"is_request_blocking_handler":      types.BoolValue(dnsApp.IsRequestBlockingHandler),
					"is_query_logger":                  types.BoolValue(dnsApp.IsQueryLogger),
					"is_post_processor":                types.BoolValue(dnsApp.IsPostProcessor),
				},
			)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			dnsAppElements = append(dnsAppElements, dnsAppObj)
		}

		// Create DNS apps list
		dnsAppsList, diags := types.ListValue(
			types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"class_path":                       types.StringType,
					"description":                      types.StringType,
					"is_app_record_request_handler":    types.BoolType,
					"record_data_template":             types.StringType,
					"is_request_controller":            types.BoolType,
					"is_authoritative_request_handler": types.BoolType,
					"is_request_blocking_handler":      types.BoolType,
					"is_query_logger":                  types.BoolType,
					"is_post_processor":                types.BoolType,
				},
			},
			dnsAppElements,
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// Create app object
		appObj, diags := types.ObjectValue(
			map[string]attr.Type{
				"name":     types.StringType,
				"version":  types.StringType,
				"dns_apps": dnsAppsList.Type(ctx),
			},
			map[string]attr.Value{
				"name":     types.StringValue(app.Name),
				"version":  types.StringValue(app.Version),
				"dns_apps": dnsAppsList,
			},
		)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		appElements = append(appElements, appObj)
	}

	// Create the apps list
	appsList, diags := types.ListValue(
		types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":    types.StringType,
				"version": types.StringType,
				"dns_apps": types.ListType{ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"class_path":                       types.StringType,
						"description":                      types.StringType,
						"is_app_record_request_handler":    types.BoolType,
						"record_data_template":             types.StringType,
						"is_request_controller":            types.BoolType,
						"is_authoritative_request_handler": types.BoolType,
						"is_request_blocking_handler":      types.BoolType,
						"is_query_logger":                  types.BoolType,
						"is_post_processor":                types.BoolType,
					},
				}},
			},
		},
		appElements,
	)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the results
	data.ID = types.StringValue("dns_apps")
	data.Apps = appsList

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
