package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &scrapingProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &scrapingProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *scrapingProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "scraping"
	resp.Version = p.version
}

type scrapingProvider struct {
	version string
}

type scrapingProviderModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// Schema defines the provider-level schema for configuration data.
func (p *scrapingProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

// Configure prepares a HashiCups API client for data sources and resources.
func (p *scrapingProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

	var data scrapingProviderModel
	username := "gwoo"
	password := "joejg"
	// Read configutation data into model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	/*
		Check configure data. which should take precedence over
		enviroment variable data, if found
	*/
	if data.Username.ValueString() != "" {
		username = data.Username.ValueString()
	}

	if data.Password.ValueString() != "" {
		password = data.Password.ValueString()
	}

	if username == "" {
		resp.Diagnostics.AddError(
			"Missing username Configuration",
			"While configuring the provider, the username was not found in "+
				"configuration block username attribute.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddError(
			"Missing password Configuration",
			"While configuring the provider, the password was not found in "+
				"configuration block password attribute.",
		)
		return
	}

}

// DataSources defines the data sources implemented in the provider.
func (p *scrapingProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewVMDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *scrapingProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVMResource,
	}
}
