package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	//"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VMResource{}
var _ resource.ResourceWithImportState = &VMResource{}

func NewVMResource() resource.Resource {
	return &VMResource{}
}

// ExampleResource defines the resource implementation.
type VMResource struct {
	client *http.Client
}

// ExampleResourceModel describes the resource data model.
type VMResourceModel struct {
	Environment  types.String `tfsdk:"environment"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	Machine_name types.String `tfsdk:"machine_name"`
}

func (r *VMResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

func (r *VMResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Example resource",

		Attributes: map[string]schema.Attribute{
			"environment": schema.StringAttribute{
				MarkdownDescription: "Example configurable attribute",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"machine_name": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *VMResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *VMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VMResourceModel
	username := "default"
	password := "default"
	machine_name := ""

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username = data.Username.ValueString()
	password = data.Password.ValueString()
	machine_name = data.Machine_name.ValueString()
	ctx = tflog.SetField(ctx, "username", username)
	ctx = tflog.SetField(ctx, "password", password)
	if machine_name == "" {
		log.Printf("machine_name is null." +
			"We will create new machine. If you want to stand-up machine which already created, you should put name in machine_name")
	} else {
		ctx = tflog.SetField(ctx, "machine_name", machine_name)
	}

	// machine名が入力されていれば起動、なければ作成
	if machine_name == "" {
		create_vm(username, password, machine_name)
		data.Machine_name = types.StringValue(machine_name)
		log.Printf("Save machine_name")
	} else {
		start_vm(username, password, machine_name)
	}

	log.Printf("Compleate!!!")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VMResourceModel
	username := "default"
	password := "default"
	machine_name := ""

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	username = data.Username.ValueString()
	password = data.Password.ValueString()
	machine_name = data.Machine_name.ValueString()

	// machine名が入力されていれば起動、なければ作成
	if machine_name == "" {
		create_vm(username, password, machine_name)
		data.Machine_name = types.StringValue(machine_name)
	} else {
		start_vm(username, password, machine_name)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VMResourceModel

	// Read Terraorm prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	// Convert from Terraform data model into strings
	var username, password, machine_name string
	username = data.Username.ValueString()
	password = data.Password.ValueString()
	machine_name = data.Machine_name.ValueString()

	delete_vm(username, password, machine_name)
}

func (r *VMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

}

func (r *VMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

}
