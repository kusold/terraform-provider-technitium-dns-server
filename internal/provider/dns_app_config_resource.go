package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DNSAppConfigResource{}
var _ resource.ResourceWithImportState = &DNSAppConfigResource{}

func NewDNSAppConfigResource() resource.Resource {
	return &DNSAppConfigResource{}
}

// DNSAppConfigResource defines the resource implementation.
type DNSAppConfigResource struct {
	client *client.Client
}

// DNSAppConfigResourceModel describes the resource data model.
type DNSAppConfigResourceModel struct {
	ID     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Config types.String `tfsdk:"config"`
}

func (r *DNSAppConfigResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_app_config"
}

func (r *DNSAppConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "DNS App Config resource for managing Technitium DNS Server application configuration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier (app name)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the DNS application to configure",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"config": schema.StringAttribute{
				MarkdownDescription: "JSON configuration for the DNS application",
				Required:            true,
			},
		},
	}
}

func (r *DNSAppConfigResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSAppConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSAppConfigResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	config := data.Config.ValueString()

	tflog.Debug(ctx, "Creating DNS app config", map[string]interface{}{
		"name": name,
	})

	// Verify the app exists
	apps, err := r.client.ListApps(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list apps: %s", err.Error()))
		return
	}

	found := false
	for _, app := range apps {
		if app.Name == name {
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError("App Not Found", fmt.Sprintf("DNS app '%s' not found. Ensure the app is installed before configuring it.", name))
		return
	}

	// Set the app configuration
	if err := r.client.SetAppConfig(ctx, name, config); err != nil {
		resp.Diagnostics.AddError("Config Creation Failed", fmt.Sprintf("Unable to set app config: %s", err.Error()))
		return
	}

	// Update the state
	data.ID = types.StringValue(name)

	tflog.Debug(ctx, "Successfully created DNS app config", map[string]interface{}{
		"name": name,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSAppConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSAppConfigResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	tflog.Debug(ctx, "Reading DNS app config", map[string]interface{}{
		"name": name,
	})

	// Verify the app still exists
	apps, err := r.client.ListApps(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list apps: %s", err.Error()))
		return
	}

	found := false
	for _, app := range apps {
		if app.Name == name {
			found = true
			break
		}
	}

	if !found {
		// App not found - it was deleted outside of Terraform
		tflog.Debug(ctx, "DNS app not found, removing config from state", map[string]interface{}{
			"name": name,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Get app configuration
	config, err := r.client.GetAppConfig(ctx, name)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get app config: %s", err.Error()))
		return
	}

	if config == nil {
		// No config found - this might indicate the config was removed outside of Terraform
		tflog.Debug(ctx, "No config found for app, removing from state", map[string]interface{}{
			"name": name,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the state
	data.Config = types.StringValue(*config)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSAppConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSAppConfigResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()
	config := data.Config.ValueString()

	tflog.Debug(ctx, "Updating DNS app config", map[string]interface{}{
		"name": name,
	})

	// Set the app configuration
	if err := r.client.SetAppConfig(ctx, name, config); err != nil {
		resp.Diagnostics.AddError("Config Update Failed", fmt.Sprintf("Unable to update app config: %s", err.Error()))
		return
	}

	tflog.Debug(ctx, "Successfully updated DNS app config", map[string]interface{}{
		"name": name,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSAppConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSAppConfigResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	tflog.Debug(ctx, "Deleting DNS app config", map[string]interface{}{
		"name": name,
	})

	// Set empty configuration to "clear" the config
	if err := r.client.SetAppConfig(ctx, name, ""); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to clear app config: %s", err.Error()))
		return
	}

	tflog.Debug(ctx, "Successfully deleted DNS app config", map[string]interface{}{
		"name": name,
	})
}

func (r *DNSAppConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using the app name as the ID
	appName := req.ID

	// Validate the app exists
	apps, err := r.client.ListApps(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list apps during import: %s", err.Error()))
		return
	}

	found := false
	for _, app := range apps {
		if app.Name == appName {
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError("App Not Found", fmt.Sprintf("DNS app '%s' not found on server", appName))
		return
	}

	// Validate the app has configuration
	config, err := r.client.GetAppConfig(ctx, appName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get app config during import: %s", err.Error()))
		return
	}

	if config == nil || *config == "" {
		resp.Diagnostics.AddError("No Config Found", fmt.Sprintf("DNS app '%s' has no configuration to import", appName))
		return
	}

	// Set the app name and ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), appName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), appName)...)
}
