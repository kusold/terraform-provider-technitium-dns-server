package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/kusold/terraform-provider-technitium-dns-server/internal/client"
)

// Ensure TechnitiumProvider satisfies various provider interfaces.
var _ provider.Provider = &TechnitiumProvider{}
var _ provider.ProviderWithFunctions = &TechnitiumProvider{}

// TechnitiumProvider defines the provider implementation.
type TechnitiumProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// TechnitiumProviderModel describes the provider data model.
type TechnitiumProviderModel struct {
	Host               types.String `tfsdk:"host"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	Token              types.String `tfsdk:"token"`
	TimeoutSeconds     types.Int64  `tfsdk:"timeout_seconds"`
	RetryAttempts      types.Int64  `tfsdk:"retry_attempts"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
}

func (p *TechnitiumProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "technitium"
	resp.Version = p.version
}

func (p *TechnitiumProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Technitium provider is used to manage Technitium DNS Server instances via the REST API.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Technitium DNS Server host URL (e.g., http://localhost:5380)",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for authentication. Either username/password or token must be provided.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for authentication. Either username/password or token must be provided.",
				Optional:            true,
				Sensitive:           true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "API token for authentication. Either username/password or token must be provided.",
				Optional:            true,
				Sensitive:           true,
			},
			"timeout_seconds": schema.Int64Attribute{
				MarkdownDescription: "Request timeout in seconds. Defaults to 30.",
				Optional:            true,
			},
			"retry_attempts": schema.Int64Attribute{
				MarkdownDescription: "Number of retry attempts for failed requests. Defaults to 3.",
				Optional:            true,
			},
			"insecure_skip_verify": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification. Defaults to false.",
				Optional:            true,
			},
		},
	}
}

func (p *TechnitiumProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data TechnitiumProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Validate configuration
	if data.Host.IsNull() || data.Host.IsUnknown() {
		resp.Diagnostics.AddError(
			"Missing Host Configuration",
			"The host configuration is required to connect to the Technitium DNS server.",
		)
		return
	}

	// Check authentication
	hasUsernamePassword := !data.Username.IsNull() && !data.Username.IsUnknown() &&
		!data.Password.IsNull() && !data.Password.IsUnknown()
	hasToken := !data.Token.IsNull() && !data.Token.IsUnknown()

	if !hasUsernamePassword && !hasToken {
		resp.Diagnostics.AddError(
			"Missing Authentication Configuration",
			"Either username/password or token must be provided for authentication.",
		)
		return
	}

	// Set defaults for optional values
	timeoutSeconds := int64(30)
	if !data.TimeoutSeconds.IsNull() && !data.TimeoutSeconds.IsUnknown() {
		timeoutSeconds = data.TimeoutSeconds.ValueInt64()
	}

	retryAttempts := int64(3)
	if !data.RetryAttempts.IsNull() && !data.RetryAttempts.IsUnknown() {
		retryAttempts = data.RetryAttempts.ValueInt64()
	}

	insecureSkipVerify := false
	if !data.InsecureSkipVerify.IsNull() && !data.InsecureSkipVerify.IsUnknown() {
		insecureSkipVerify = data.InsecureSkipVerify.ValueBool()
	}

	// Create client configuration
	config := client.Config{
		Host:               data.Host.ValueString(),
		TimeoutSeconds:     timeoutSeconds,
		RetryAttempts:      retryAttempts,
		InsecureSkipVerify: insecureSkipVerify,
	}

	if hasToken {
		config.Token = data.Token.ValueString()
	} else {
		config.Username = data.Username.ValueString()
		config.Password = data.Password.ValueString()
	}

	// Create the client
	apiClient, err := client.NewClient(config)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Technitium Client",
			"An unexpected error occurred when creating the Technitium API client. "+
				"Error: "+err.Error(),
		)
		return
	}

	// Test the connection by authenticating
	if err := apiClient.Authenticate(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Authenticate with Technitium Server",
			"Failed to authenticate with the Technitium DNS server. "+
				"Please verify your credentials and server URL. "+
				"Error: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Successfully configured Technitium DNS provider", map[string]interface{}{
		"host":        data.Host.ValueString(),
		"auth_method": map[bool]string{true: "token", false: "username/password"}[hasToken],
	})

	// Make client available to data sources and resources
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

func (p *TechnitiumProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewZoneResource,
		NewDNSRecordResource,
	}
}

func (p *TechnitiumProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewZoneDataSource,
		NewDNSRecordsDataSource,
	}
}

func (p *TechnitiumProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		// TODO: Add functions if needed
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TechnitiumProvider{
			version: version,
		}
	}
}
