package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &VMDataSource{}
)

// NewVMDataSource is a helper function to simplify the provider implementation.
func NewVMDataSource() datasource.DataSource {
	return &VMDataSource{}
}

// VMDataSource is the data source implementation.
type VMDataSource struct {
	client *http.Client
}

type VMDataSourceModel struct {
	instance_type types.String `tfsdk:"instance_type"`
}

// Metadata returns the data source type name.
func (d *VMDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_VM"
}

// Schema defines the schema for the data source.
func (d *VMDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "VM data source",

		Attributes: map[string]schema.Attribute{
			"instance_type": schema.StringAttribute{
				MarkdownDescription: "VM Machine Type",
				Computed:            true,
				Optional:            true,
			},
		},
	}
}

func (d *VMDataSource) Configre(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Congigure Type",
			fmt.Sprintf("Expecred *http.Client, got: %T. Please report this issue to the provider devrlopers", req.ProviderData),
		)

		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *VMDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VMDataSourceModel

	// Read Terraform configuration data inro the  model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.instance_type = types.StringValue("t4g.medium")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (d *VMDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	/*
	 Add a nil check when handling ProviderData because
	 Terraform sets that data after it calls the ConfigureProvider RPC.
	*/
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *hashicups.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	d.client = client
}
