package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &DNSAppResource{}
var _ resource.ResourceWithImportState = &DNSAppResource{}

func NewDNSAppResource() resource.Resource {
	return &DNSAppResource{}
}

// DNSAppResource defines the resource implementation.
type DNSAppResource struct {
	client *client.Client
}

// DNSAppResourceModel describes the resource data model.
type DNSAppResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	InstallMethod types.String `tfsdk:"install_method"`
	URL           types.String `tfsdk:"url"`
	FileContent   types.String `tfsdk:"file_content"`
	Config        types.String `tfsdk:"config"`

	// Computed attributes
	Version types.String `tfsdk:"version"`
	DNSApps types.List   `tfsdk:"dns_apps"`
}

// DNSAppInfo represents a single DNS app within an app package for Terraform
type DNSAppInfo struct {
	ClassPath                     types.String `tfsdk:"class_path"`
	Description                   types.String `tfsdk:"description"`
	IsAppRecordRequestHandler     types.Bool   `tfsdk:"is_app_record_request_handler"`
	RecordDataTemplate            types.String `tfsdk:"record_data_template"`
	IsRequestController           types.Bool   `tfsdk:"is_request_controller"`
	IsAuthoritativeRequestHandler types.Bool   `tfsdk:"is_authoritative_request_handler"`
	IsRequestBlockingHandler      types.Bool   `tfsdk:"is_request_blocking_handler"`
	IsQueryLogger                 types.Bool   `tfsdk:"is_query_logger"`
	IsPostProcessor               types.Bool   `tfsdk:"is_post_processor"`
}

func (r *DNSAppResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_app"
}

func (r *DNSAppResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "DNS App resource for managing Technitium DNS Server applications.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Resource identifier (app name)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the DNS application",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9\s\-_.]+$`),
						"App name can only contain alphanumeric characters, spaces, hyphens, underscores, and periods",
					),
				},
			},
			"install_method": schema.StringAttribute{
				MarkdownDescription: "Installation method: 'url' to download from URL, 'file' to upload from file content",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("url", "file"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "URL to download the app from (required when install_method is 'url')",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^https://.*\.zip$`),
						"URL must start with https:// and end with .zip",
					),
				},
			},
			"file_content": schema.StringAttribute{
				MarkdownDescription: "Base64-encoded content of the app zip file (required when install_method is 'file')",
				Optional:            true,
				Sensitive:           true,
			},
			"config": schema.StringAttribute{
				MarkdownDescription: "JSON configuration for the DNS application",
				Optional:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Version of the installed app",
				Computed:            true,
			},
			"dns_apps": schema.ListNestedAttribute{
				MarkdownDescription: "List of DNS applications within this app package",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"class_path": schema.StringAttribute{
							MarkdownDescription: "Class path of the DNS application",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description of the DNS application",
							Computed:            true,
						},
						"is_app_record_request_handler": schema.BoolAttribute{
							MarkdownDescription: "Whether this app handles APP record requests",
							Computed:            true,
						},
						"record_data_template": schema.StringAttribute{
							MarkdownDescription: "Record data template for APP records",
							Computed:            true,
						},
						"is_request_controller": schema.BoolAttribute{
							MarkdownDescription: "Whether this app is a request controller",
							Computed:            true,
						},
						"is_authoritative_request_handler": schema.BoolAttribute{
							MarkdownDescription: "Whether this app handles authoritative requests",
							Computed:            true,
						},
						"is_request_blocking_handler": schema.BoolAttribute{
							MarkdownDescription: "Whether this app handles request blocking",
							Computed:            true,
						},
						"is_query_logger": schema.BoolAttribute{
							MarkdownDescription: "Whether this app is a query logger",
							Computed:            true,
						},
						"is_post_processor": schema.BoolAttribute{
							MarkdownDescription: "Whether this app is a post processor",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (r *DNSAppResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DNSAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DNSAppResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate install method configuration
	if err := r.validateInstallMethod(data); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}

	name := data.Name.ValueString()

	tflog.Debug(ctx, "Creating DNS app", map[string]interface{}{
		"name":           name,
		"install_method": data.InstallMethod.ValueString(),
	})

	// Install the app based on the method
	var app *client.App
	var err error

	switch data.InstallMethod.ValueString() {
	case "url":
		url := data.URL.ValueString()
		app, err = r.client.DownloadAndInstallApp(ctx, name, url)
	case "file":
		fileContent := data.FileContent.ValueString()
		fileData, decodeErr := decodeBase64(fileContent)
		if decodeErr != nil {
			resp.Diagnostics.AddError("Invalid File Content", fmt.Sprintf("Failed to decode base64 file content: %s", decodeErr.Error()))
			return
		}
		app, err = r.client.InstallApp(ctx, name, fileData)
	}

	if err != nil {
		resp.Diagnostics.AddError("App Installation Failed", fmt.Sprintf("Unable to install app: %s", err.Error()))
		return
	}

	// Set app configuration if provided
	if !data.Config.IsNull() && !data.Config.IsUnknown() {
		config := data.Config.ValueString()
		if err := r.client.SetAppConfig(ctx, name, config); err != nil {
			tflog.Warn(ctx, "Failed to set app config", map[string]interface{}{
				"error": err.Error(),
			})
			// Don't fail the resource creation for config errors
		}
	}

	// Update the state with the installed app data
	data.ID = types.StringValue(name)
	data.Version = types.StringValue(app.Version)

	// Convert DNS apps to Terraform format
	dnsApps, diags := r.convertDNSAppsToTerraform(ctx, app.DNSApps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.DNSApps = dnsApps

	tflog.Debug(ctx, "Successfully created DNS app", map[string]interface{}{
		"name":           name,
		"version":        app.Version,
		"dns_apps_count": len(app.DNSApps),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DNSAppResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	tflog.Debug(ctx, "Reading DNS app", map[string]interface{}{
		"name": name,
	})

	// Get list of installed apps
	apps, err := r.client.ListApps(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read apps: %s", err.Error()))
		return
	}

	// Find our app
	var app *client.App
	for _, a := range apps {
		if a.Name == name {
			app = &a
			break
		}
	}

	if app == nil {
		// App not found - it was deleted outside of Terraform
		tflog.Debug(ctx, "DNS app not found, removing from state", map[string]interface{}{
			"name": name,
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Get app configuration
	config, err := r.client.GetAppConfig(ctx, name)
	if err != nil {
		tflog.Warn(ctx, "Failed to get app config", map[string]interface{}{
			"error": err.Error(),
		})
		// Don't fail the read for config errors, just set to null
		data.Config = types.StringNull()
	} else if config != nil {
		data.Config = types.StringValue(*config)
	} else {
		data.Config = types.StringNull()
	}

	// Update computed attributes
	data.Version = types.StringValue(app.Version)

	// Convert DNS apps to Terraform format
	dnsApps, diags := r.convertDNSAppsToTerraform(ctx, app.DNSApps)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.DNSApps = dnsApps

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DNSAppResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	tflog.Debug(ctx, "Updating DNS app", map[string]interface{}{
		"name": name,
	})

	// Handle app updates based on install method
	if !data.URL.IsNull() && !data.URL.IsUnknown() && data.InstallMethod.ValueString() == "url" {
		url := data.URL.ValueString()
		app, err := r.client.DownloadAndUpdateApp(ctx, name, url)
		if err != nil {
			resp.Diagnostics.AddError("App Update Failed", fmt.Sprintf("Unable to update app: %s", err.Error()))
			return
		}

		// Update computed attributes
		data.Version = types.StringValue(app.Version)

		// Convert DNS apps to Terraform format
		dnsApps, diags := r.convertDNSAppsToTerraform(ctx, app.DNSApps)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.DNSApps = dnsApps
	} else if !data.FileContent.IsNull() && !data.FileContent.IsUnknown() && data.InstallMethod.ValueString() == "file" {
		fileContent := data.FileContent.ValueString()
		fileData, err := decodeBase64(fileContent)
		if err != nil {
			resp.Diagnostics.AddError("Invalid File Content", fmt.Sprintf("Failed to decode base64 file content: %s", err.Error()))
			return
		}

		app, err := r.client.UpdateApp(ctx, name, fileData)
		if err != nil {
			resp.Diagnostics.AddError("App Update Failed", fmt.Sprintf("Unable to update app: %s", err.Error()))
			return
		}

		// Update computed attributes
		data.Version = types.StringValue(app.Version)

		// Convert DNS apps to Terraform format
		dnsApps, diags := r.convertDNSAppsToTerraform(ctx, app.DNSApps)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.DNSApps = dnsApps
	}

	// Update app configuration if provided
	if !data.Config.IsNull() && !data.Config.IsUnknown() {
		config := data.Config.ValueString()
		if err := r.client.SetAppConfig(ctx, name, config); err != nil {
			resp.Diagnostics.AddError("Config Update Failed", fmt.Sprintf("Unable to update app config: %s", err.Error()))
			return
		}
	}

	tflog.Debug(ctx, "Successfully updated DNS app", map[string]interface{}{
		"name": name,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DNSAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data DNSAppResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	name := data.Name.ValueString()

	tflog.Debug(ctx, "Deleting DNS app", map[string]interface{}{
		"name": name,
	})

	if err := r.client.UninstallApp(ctx, name); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to uninstall app: %s", err.Error()))
		return
	}

	tflog.Debug(ctx, "Successfully deleted DNS app", map[string]interface{}{
		"name": name,
	})
}

func (r *DNSAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

	// Set the app name and ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), appName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), appName)...)

	// Set install_method to "url" as default for imported resources
	// Users will need to update the configuration with the actual install method
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("install_method"), "url")...)
}

// Helper functions

func (r *DNSAppResource) validateInstallMethod(data DNSAppResourceModel) error {
	installMethod := data.InstallMethod.ValueString()

	switch installMethod {
	case "url":
		if data.URL.IsNull() || data.URL.IsUnknown() {
			return fmt.Errorf("'url' is required when install_method is 'url'")
		}
		if !data.FileContent.IsNull() && !data.FileContent.IsUnknown() {
			return fmt.Errorf("'file_content' should not be set when install_method is 'url'")
		}
	case "file":
		if data.FileContent.IsNull() || data.FileContent.IsUnknown() {
			return fmt.Errorf("'file_content' is required when install_method is 'file'")
		}
		if !data.URL.IsNull() && !data.URL.IsUnknown() {
			return fmt.Errorf("'url' should not be set when install_method is 'file'")
		}
	default:
		return fmt.Errorf("invalid install_method: %s", installMethod)
	}

	return nil
}

func (r *DNSAppResource) convertDNSAppsToTerraform(ctx context.Context, dnsApps []client.DNSApp) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics

	if len(dnsApps) == 0 {
		return types.ListNull(types.ObjectType{
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
		}), diags
	}

	// Convert to Terraform objects
	var elements []attr.Value
	for _, dnsApp := range dnsApps {
		recordDataTemplate := types.StringNull()
		if dnsApp.RecordDataTemplate != nil {
			recordDataTemplate = types.StringValue(*dnsApp.RecordDataTemplate)
		}

		obj, objDiags := types.ObjectValue(
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
		diags.Append(objDiags...)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{}), diags
		}
		elements = append(elements, obj)
	}

	list, listDiags := types.ListValue(
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
		elements,
	)
	diags.Append(listDiags...)

	return list, diags
}

// Helper functions

func decodeBase64(encoded string) ([]byte, error) {
	// Remove any whitespace
	encoded = strings.ReplaceAll(encoded, " ", "")
	encoded = strings.ReplaceAll(encoded, "\n", "")
	encoded = strings.ReplaceAll(encoded, "\r", "")
	encoded = strings.ReplaceAll(encoded, "\t", "")

	return base64.StdEncoding.DecodeString(encoded)
}
