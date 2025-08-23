package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Token    types.String `tfsdk:"token"`
}

func (p *TechnitiumProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "technitium"
	resp.Version = p.version
}

func (p *TechnitiumProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Technitium DNS Server host URL",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for authentication",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for authentication",
				Optional:            true,
				Sensitive:           true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "API token for authentication",
				Optional:            true,
				Sensitive:           true,
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

	// Configuration values are now available:
	// data.Host, data.Username, data.Password, data.Token

	// TODO: Create client and set in resp.DataSourceData and resp.ResourceData
}

func (p *TechnitiumProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// TODO: Add resources
	}
}

func (p *TechnitiumProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// TODO: Add data sources
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