package provider

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/sclevine/agouti"
	"strconv"
	"time"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VMResource{}
var _ resource.ResourceWithImportState = &VMResource{}
var _ function.Function = &ip{}

func NewVMResource() resource.Resource {
	return &VMResource{}
}

// ExampleResource defines the resource implementation.
type VMResource struct {
	client *http.Client
}

type Machine_Data struct {
	enviroment   string
	username     string
	password     string
	machine_name string
	machine_stop bool
}

// ExampleResourceModel describes the resource data model.
type VMResourceModel struct {
	Environment  types.String `tfsdk:"environment"`
	Username     types.String `tfsdk:"username"`
	Password     types.String `tfsdk:"password"`
	Machine_name types.String `tfsdk:"machine_name"`
	Machine_stop types.Bool   `tfsdk:"machine_stop"`
}

// 　IPアドレスをスクレイピングする関数
type ip struct{}

func NewIp() function.Function {
	return &ip{}
}

func (f *ip) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "ip"
}

func (f *ip) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Search for ip address",
		Description: "Given a machine_name, return ip address",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "machine_name",
				Description: "machine's name",
			},
		},
		Return: function.StringReturn{},
	}
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

			"machine_stop": schema.BoolAttribute{
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
	var Machine_Data Machine_Data

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Machine_Dataにtfファイルの内容を渡す
	Machine_Data.username = data.Username.ValueString()
	Machine_Data.password = data.Password.ValueString()
	Machine_Data.machine_name = data.Machine_name.ValueString()
	Machine_Data.machine_stop = data.Machine_stop.ValueBool()

	ctx = tflog.SetField(ctx, "username", Machine_Data.username)
	ctx = tflog.SetField(ctx, "password", Machine_Data.password)

	if Machine_Data.machine_name == "" {
		// マシン名が指定されていない時、新規でVMを立ち上げる
		log.Printf("machine_name is null." +
			"We will create new machine. If you want to stand-up machine which already created, you should put name in machine_name")
	} else {
		ctx = tflog.SetField(ctx, "machine_name", Machine_Data.machine_name)
	}
	if Machine_Data.machine_name == "" {
		// machine名が入力されてなければ作成
		create_vm(Machine_Data)
	} else {

		if Machine_Data.machine_stop {
			stop_vm(Machine_Data)

		} else {
			log.Printf("already start VM.")
			start_vm(Machine_Data)
		}

	}

	log.Printf("Compleate!!!")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *VMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

	var data VMResourceModel
	var Machine_Data Machine_Data

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	Machine_Data.username = data.Username.ValueString()
	Machine_Data.password = data.Password.ValueString()
	Machine_Data.machine_name = data.Machine_name.ValueString()
	Machine_Data.machine_stop = data.Machine_stop.ValueBool()

	ctx = tflog.SetField(ctx, "username", Machine_Data.username)
	ctx = tflog.SetField(ctx, "password", Machine_Data.password)

	if Machine_Data.machine_name == "" {
		log.Printf("machine_name is null." +
			"We will create new machine. If you want to stand-up machine which already created, you should put name in machine_name")
	} else {
		ctx = tflog.SetField(ctx, "machine_name", Machine_Data.machine_name)
	}
	// machine名が入力されていれば起動、なければ作成
	if Machine_Data.machine_name == "" {
		create_vm(Machine_Data)
		//log.Printf("Save machine_name")
	} else {

		if Machine_Data.machine_stop {
			stop_vm(Machine_Data)
		} else {
			log.Printf("スタートしてるよーーーー")
			start_vm(Machine_Data)
		}

	}

	log.Printf("Compleate!!!")
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
	//machine_stop := data.Machine_stop.ValueBool()
	var choice string

	fmt.Printf("%s kill or stop? Pleace input your choice.", machine_name)
	fmt.Scan(&choice)

	if machine_name == "" {
		f, err := os.Open(".machine_name.txt")
		if err != nil {
			fmt.Println("Can't get machine_name. You should confirm the file which named \".machine_name.txt\" .")
		}
		defer f.Close()

		buf := make([]byte, 1024)
		n, err := f.Read(buf)
		if err != nil {
			fmt.Printf("error! You should confirm the file which named \".machine_name.txt\" .")
		}

		machine_name = string(buf[:n])
		log.Printf("%s", machine_name)
		//delete_vm(username, password, machine_name)

	}
	log.Printf("うおおおおおおおおおお%s\n\n\n\n\n", machine_name)

	delete_vm(username, password, machine_name)
}

func (r *VMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	// Get refreshed order value from HashiCups
}

func (r *VMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

}

func (f *ip) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var Machine_Data Machine_Data
	var ip string

	// Read Terraform argument data into the variables
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &Machine_Data))

	//　スクレイピングでipアドレスを抽出する
	driver := agouti.ChromeDriver(agouti.Browser("chrome"))
	/*
		   デバック中のためコメントアウト
			   driver := agouti.ChromeDriver(
				   agouti.ChromeOptions(
					   "args", []string{
						   "--headless",
						   "--disavle-gpu",
					   }),
			   )*/
	log.Printf("Open Google Chorome...")

	if err := driver.Start(); err != nil {
		log.Fatalf("Failed to start driver:%v", err)
	}
	defer driver.Stop()
	log.Printf("Access to FUN VM WebAPI...")
	page, err := driver.NewPage()
	if err != nil {
		log.Fatalf("Failed to open page:%v", err)
		time.Sleep(1 * time.Second)
	} // go to login page
	if err := page.Navigate("https://manage.p.fun.ac.jp/server_manage"); err != nil {
		log.Fatalf("Failed to navigate:%v", err)
	}
	log.Printf("Success to FUN VM WebAPI")
	time.Sleep(1 * time.Second)

	elem_user := page.FindByName("username")
	log.Printf("Input username = %v", Machine_Data.username)

	elem_pass := page.FindByName("password")
	log.Printf("Input password")

	elem_user.Fill(Machine_Data.username)
	elem_pass.Fill(Machine_Data.password)
	log.Printf("login...")
	// Submit
	if err := page.FindByClass("credentials_input_submit").Click(); err != nil {
		log.Fatalf("Failed to login:%v", err)
		return
	}
	log.Printf("Success to login FUN VM WebAPI!!")

	time.Sleep(1 * time.Second)
	if err := page.FindByXPath("/html/body/div/div/main/div/form/div[2]/div/span").Click(); err != nil {
		log.Fatalf("Failed to choice:%v", err)
		return
	}

	for i := 0; i < 20; i++ {
		log.Printf("serch for machin_name = %v", Machine_Data.machine_name)
		instance_name := page.FindByID("INSTANCE_NAME_" + strconv.Itoa(i))

		// web上からterraformに指定されたmachine_nameと合致するものを探す
		if text, err := instance_name.Text(); err == nil {
			if text == Machine_Data.machine_name {
				log.Printf("found machin_name = %v!!!", Machine_Data.machine_name)
				log.Printf("start %v...", Machine_Data.machine_name)
				log.Printf("%v", page.FindByID("copiable-ip_address-"+strconv.Itoa(i)))

			}
		}

	}
	page.CloseWindow()
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, ip))
	return
}
