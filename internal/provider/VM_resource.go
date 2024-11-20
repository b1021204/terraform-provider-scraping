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
	"net"
	"strconv"
	"strings"
	"time"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VMResource{}
var _ resource.ResourceWithImportState = &VMResource{}
var _ function.Function = &ip{}
var _ function.Function = &machine_pass{}
var _ function.Function = &key{}

func NewVMResource() resource.Resource {
	return &VMResource{}
}

// ExampleResource defines the resource implementation.
type VMResource struct {
	client *http.Client
}

// 各関数内で使われるデータの構造体
type Machine_Data struct {
	environment   string
	username      string
	password      string
	machine_name  string
	machine_stop  bool
	instance_type string
	ip            string
	machine_pass  string
}

// ExampleResourceModel describes the resource data model.
type VMResourceModel struct {
	Environment   types.String `tfsdk:"environment"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	Machine_name  types.String `tfsdk:"machine_name"`
	Machine_stop  types.Bool   `tfsdk:"machine_stop"`
	Instance_type types.String `tfsdk:"instance_type"`
	Ip            types.String `tfsdk:"ip"`
	Machine_pass  types.String `tfsdk:"machine_pass"`
}

// 　IPアドレスをスクレイピングする関数
type ip struct{}

// VMのパスワードをスクレイピングする関数
type machine_pass struct{}

// VM の鍵をダウンロードし、アドレスを返す関数
type key struct{}

func NewIp() function.Function {
	return &ip{}
}

func NewMachinePass() function.Function {
	return &machine_pass{}
}

func NewKey() function.Function {
	return &key{}
}

// ipアドレススクレイピング用のメタデータ
func (f *ip) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "ip"
}

// ipアドレススクレイピング用の定義
func (f *ip) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Search for ip address",
		Description: "Given a machine_name, return ip address",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "username",
				Description: "username",
			},
			function.StringParameter{
				Name:        "password",
				Description: "pass",
			},
			function.StringParameter{
				Name:        "environment",
				Description: "env of VM",
			},
			function.StringParameter{
				Name:        "machine_name",
				Description: "machine's name",
			},
		},
		Return: function.StringReturn{},
	}
}

// マシンパスワードスクレイピング用のメタデータ
func (f *machine_pass) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "machien_pass"
}

// マシンパスワードスクレイピング用の定義
func (f *machine_pass) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Search for VM's password",
		Description: "Given a machine_name and fun userdata, return machine password",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "username",
				Description: "username",
			},
			function.StringParameter{
				Name:        "password",
				Description: "pass",
			},
			function.StringParameter{
				Name:        "environment",
				Description: "env of VM",
			},
			function.StringParameter{
				Name:        "machine_name",
				Description: "machine's name",
			},
		},
		Return: function.StringReturn{},
	}
}

// 鍵ダウンロード関数用のメタデータ
func (f *key) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "key"
}

// 鍵ダウンロード関数用の定義
func (f *key) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary:     "Download ssh key and return key address",
		Description: "Pleace give usenamae, pass, env and address you want to download.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:        "username",
				Description: "username",
			},
			function.StringParameter{
				Name:        "password",
				Description: "pass",
			},
			function.StringParameter{
				Name:        "environment",
				Description: "env of VM",
			},
			function.StringParameter{
				Name:        "address",
				Description: "address which you want to download.",
			},
		},
		Return: function.StringReturn{},
	}
}

// resource用のメタデータ
func (r *VMResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resource"
}

// resource用のスキーマ
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
				Computed: true,
			},
			"machine_stop": schema.BoolAttribute{
				Optional: true,
			},
			"instance_type": schema.StringAttribute{
				Optional: true,
			},
			"ip": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"machine_pass": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

// resource用のconigure
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

// マシン作成時の動作
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
	Machine_Data.environment = data.Environment.ValueString()
	Machine_Data.instance_type = data.Instance_type.ValueString()
	Machine_Data.ip = data.Ip.ValueString()
	Machine_Data.machine_pass = data.Machine_pass.ValueString()

	ctx = tflog.SetField(ctx, "username", Machine_Data.username)
	ctx = tflog.SetField(ctx, "password", Machine_Data.password)

	if Machine_Data.machine_name == "" {
		// マシン名が指定されていない時、新規でVMを立ち上げる
		log.Printf("machine_name is null." +
			"We will create new machine. If you want to stand-up machine which already created, you should put name in machine_name\n")
	} else {
		ctx = tflog.SetField(ctx, "machine_name", Machine_Data.machine_name)
	}
	if Machine_Data.machine_name == "" {
		// machine名が入力されてなければ作成
		create_vm(&Machine_Data)
		data.Machine_name = types.StringValue(Machine_Data.machine_name)
	} else {

		if Machine_Data.machine_stop {
			stop_vm(&Machine_Data)

		} else {
			log.Printf("already start VM.\n")
			start_vm(&Machine_Data)
		}

	}
	data.Ip = types.StringValue(Machine_Data.ip)
	data.Machine_pass = types.StringValue(Machine_Data.machine_pass)
	log.Printf("Compleate!!!\n")
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
	Machine_Data.environment = data.Environment.ValueString()
	Machine_Data.machine_name = data.Machine_name.ValueString()
	Machine_Data.machine_stop = data.Machine_stop.ValueBool()
	Machine_Data.instance_type = data.Instance_type.ValueString()
	Machine_Data.ip = data.Ip.ValueString()
	Machine_Data.machine_pass = data.Machine_pass.ValueString()

	ctx = tflog.SetField(ctx, "username", Machine_Data.username)
	ctx = tflog.SetField(ctx, "password", Machine_Data.password)

	if Machine_Data.machine_name == "" {
		log.Printf("machine_name is null." +
			"We will create new machine. If you want to stand-up machine which already created, you should put name in machine_name\n")
	} else {
		ctx = tflog.SetField(ctx, "machine_name\n", Machine_Data.machine_name)
	}
	// machine名が入力されていれば起動、なければ作成
	if Machine_Data.machine_name == "" {
		log.Printf("Now, create new vm machine...\n")
		create_vm(&Machine_Data)
		data.Machine_name = types.StringValue(Machine_Data.machine_name)
	} else {

		if Machine_Data.machine_stop {
			log.Printf("Now, %v is stoping...\n", Machine_Data.machine_name)
			stop_vm(&Machine_Data)
		} else {
			log.Printf("Now, %v is starting...\n", Machine_Data.machine_name)
			start_vm(&Machine_Data)

		}

	}

	data.Ip = types.StringValue(Machine_Data.ip)
	data.Machine_pass = types.StringValue(Machine_Data.machine_pass)
	log.Printf("Compleate!!!\n")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	var data VMResourceModel
	var Machine_Data Machine_Data

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	Machine_Data.username = data.Username.ValueString()
	Machine_Data.password = data.Password.ValueString()
	Machine_Data.environment = data.Environment.ValueString()
	Machine_Data.machine_name = data.Machine_name.ValueString()
	Machine_Data.machine_stop = data.Machine_stop.ValueBool()
	/*
	   machinedattaが空欄になることはないよう設計のためコメントアウト
	   	if Machine_Data.machine_name == "" {
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

	   		Machine_Data.machine_name = string(buf[:n])
	   	}
	*/
	log.Printf("%s", Machine_Data.machine_name)
	log.Printf("Deleating: %s...\n", Machine_Data.machine_name)

	delete_vm(Machine_Data)
}

func (r *VMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Get refreshed order value from HashiCups
}

func (r *VMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {

}

func (f *ip) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	//var Machine_Data Machine_Data
	var ip string
	var username string
	var password string
	var environment string
	var machine_name string

	// Read Terraform argument data into the variables
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &username, &password, &environment, &machine_name))
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
	log.Printf("Open Google Chorome...\n")

	if err := driver.Start(); err != nil {
		log.Fatalf("Failed to start driver:%v\n", err)
	}

	defer driver.Stop()
	page, err := driver.NewPage()
	if err != nil {
		log.Fatalf("Failed to open Chorome page:%v\n", err)
	}
	log.Printf("Success to open Google Chorome.\n")

	// access to FUN login page..
	log.Printf("Access to FUN VM WebAPI...\n")
	if err := page.Navigate("https://manage.p.fun.ac.jp/server_manage"); err != nil {
		log.Fatalf("Failed to access to FUN VM WebAPI:%v\n", err)
	}

	time.Sleep(1 * time.Second)

	// 入力ボックスにユーザ名・パスを打ち込む
	elem_user := page.FindByName("username")
	elem_pass := page.FindByName("password")
	elem_user.Fill(username)
	elem_pass.Fill(password)
	log.Printf("fill username: %v\n", username)
	log.Printf("fill password\n")

	// Submit
	if err := page.FindByClass("credentials_input_submit").Click(); err != nil {
		log.Fatalf("Failed to login:%v\n", err)
		return
	}
	log.Printf("Success to login FUN VM WebAPI!!\n")

	time.Sleep(1 * time.Second)

	// 環境画面の項目数を入れる関数。暫定５個に設定しておく
	max_environment := 5
	for i := 1; i <= max_environment; i++ {

		log.Printf("Serch for environment: %v\n...", environment)
		text, _ := page.FindByXPath("/html/body/div/div/main/div/form/div[1]/div/select/option[" + strconv.Itoa(i) + "]").Text()
		if text == environment {

			log.Printf("get environment: %v\n", text)
			if err := page.FindByXPath("/html/body/div/div/main/div/form/div[1]/div/select/option[" + strconv.Itoa(i) + "]").Click(); err != nil {
				log.Fatalf("Failed to click environment: %v\n", err)
			}
			break
		}
		//　max_environment個分のの項目をチェックしてなかった場合エラーにする
		if i == max_environment {
			log.Fatalf("Can't look up environment: %v\n", environment)
		}
	}

	// 次のページへ行く
	if err := page.FindByXPath("/html/body/div/div/main/div/form/div[2]/div/span").Click(); err != nil {
		log.Fatalf("faild to click next page bottuon")
	}

	time.Sleep(1 * time.Second)

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// ipアドレスを精査して,学内アドレスかグローバルIPアドレスかを判別

	var univ_ip bool
	for _, addr := range addrs {
		ip_text := addr.String()
		if strings.Index(ip_text, "10.") == 0 {
			univ_ip = true
			break
		}
		univ_ip = false
	}
	if univ_ip {
		log.Println("You use univ wifi like fun-wifi or free-wifi")
	} else {
		log.Println("You don't use univ wifi")
	}

	max_machine := 5
	for i := 0; i <= max_machine; i++ {
		log.Printf("serch for machin_name = %v", machine_name)
		instance_name := page.FindByID("INSTANCE_NAME_" + strconv.Itoa(i))

		// web上からterraformに指定されたmachine_nameと合致するものを探す
		if text, err := instance_name.Text(); err == nil {
			if text == machine_name {
				log.Printf("found machin_name = %v!!!", machine_name)
				log.Printf("start %v...", machine_name)
				if univ_ip {
					ip, _ = page.FindByID("copiable-ip_address-" + strconv.Itoa(i)).Text()
				} else {
					ip, _ = page.FindByID("copiable-public_ip_address-" + strconv.Itoa(i)).Text()
					log.Println(ip + "\n\n")
				}
				log.Printf("%v", ip)
				break
			}
		}
		if max_machine == i {
			log.Fatalf("Can't get machine_name")
		}
	}
	page.CloseWindow()
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, ip))
	return
}

// マシンパススクレイピング用のrun

func (f *machine_pass) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	//var Machine_Data Machine_Data
	var machine_pass string
	var username string
	var password string
	var environment string
	var machine_name string

	// Read Terraform argument data into the variables
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &username, &password, &environment, &machine_name))
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
	log.Printf("Open Google Chorome...\n")

	if err := driver.Start(); err != nil {
		log.Fatalf("Failed to start driver:%v\n", err)
	}

	defer driver.Stop()
	page, err := driver.NewPage()
	if err != nil {
		log.Fatalf("Failed to open Chorome page:%v\n", err)
	}
	log.Printf("Success to open Google Chorome.\n")

	// access to FUN login page..
	log.Printf("Access to FUN VM WebAPI...\n")
	if err := page.Navigate("https://manage.p.fun.ac.jp/server_manage"); err != nil {
		log.Fatalf("Failed to access to FUN VM WebAPI:%v\n", err)
	}

	time.Sleep(1 * time.Second)

	// 入力ボックスにユーザ名・パスを打ち込む
	elem_user := page.FindByName("username")
	elem_pass := page.FindByName("password")
	elem_user.Fill(username)
	elem_pass.Fill(password)
	log.Printf("fill username: %v\n", username)
	log.Printf("fill password\n")

	// Submit
	if err := page.FindByClass("credentials_input_submit").Click(); err != nil {
		log.Fatalf("Failed to login:%v\n", err)
		return
	}
	log.Printf("Success to login FUN VM WebAPI!!\n")

	time.Sleep(1 * time.Second)

	// 環境画面の項目数を入れる関数。暫定５個に設定しておく
	max_environment := 5
	for i := 1; i <= max_environment; i++ {

		log.Printf("Serch for environment: %v\n...", environment)
		text, _ := page.FindByXPath("/html/body/div/div/main/div/form/div[1]/div/select/option[" + strconv.Itoa(i) + "]").Text()
		if text == environment {

			log.Printf("get environment: %v\n", text)
			if err := page.FindByXPath("/html/body/div/div/main/div/form/div[1]/div/select/option[" + strconv.Itoa(i) + "]").Click(); err != nil {
				log.Fatalf("Failed to click environment: %v\n", err)
			}
			break
		}
		//　max_environment個分のの項目をチェックしてなかった場合エラーにする
		if i == max_environment {
			log.Fatalf("Can't look up environment: %v\n", environment)
		}
	}

	// 次のページへ行く
	if err := page.FindByXPath("/html/body/div/div/main/div/form/div[2]/div/span").Click(); err != nil {
		log.Fatalf("faild to click next page bottuon")
	}

	time.Sleep(1 * time.Second)

	max_machine := 5
	for i := 0; i <= max_machine; i++ {
		log.Printf("serch for machin_name = %v\n", machine_name)
		instance_name := page.FindByID("INSTANCE_NAME_" + strconv.Itoa(i))

		// web上からterraformに指定されたmachine_nameと合致するものを探す
		if text, err := instance_name.Text(); err == nil {
			if text == machine_name {
				log.Printf("found machin_name = %v!!!\n", machine_name)
				log.Printf("start %v...", machine_name)
				machine_pass, _ = page.FindByID("copiable-password-" + strconv.Itoa(i)).Text()
				log.Printf("%v", machine_pass)
			}
		}
		if max_machine == i {
			log.Fatalf("Can't get machine_name")
		}

	}
	page.CloseWindow()
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, machine_pass))
	return
}

// 鍵ダウンロード用のrun

func (f *key) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	//var Machine_Data Machine_Data
	var username string
	var password string
	var environment string
	var address string

	// Read Terraform argument data into the variables
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &username, &password, &environment, &address))
	// キーが存在しているか確認
	filename := address + "/" + "funawskey" + username + ".pem"
	log.Printf("\n%v\n\n", filename)
	if _, err := os.Stat(filename); err == nil {
		log.Printf("Already, you have %v!! \n", filename)
		address := filename
		resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, address))
		return
	} else {
		log.Printf("Now, Download key.\n")

		driver := agouti.ChromeDriver(
			// ここでChromeOptions
			agouti.ChromeOptions("prefs", map[string]interface{}{
				"download.default_directory":         address,
				"download.prompt_for_download":       false,
				"download.directory_upgrade":         true,
				"plugins.plugins_disabled":           "Chrome PDF Viewer",
				"plugins.always_open_pdf_externally": true,
			}),
			agouti.ChromeOptions("args", []string{
				"--headless",
				"--disavle-gpu",
			}),
			/*
				agouti.ChromeOptions("args", []string{
					"--disable-extensions",
					"--disable-print-preview",
					"--ignore-certificate-errors",
				}),*/
			agouti.Debug,
		)

		log.Printf("Open Google Chorome...\n")

		if err := driver.Start(); err != nil {
			log.Fatalf("Failed to start driver:%v\n", err)
		}

		defer driver.Stop()
		page, err := driver.NewPage()
		if err != nil {
			log.Fatalf("Failed to open Chorome page:%v\n", err)
		}
		log.Printf("Success to open Google Chorome.\n")

		// access to FUN login page..
		log.Printf("Access to FUN VM WebAPI...\n")
		if err := page.Navigate("https://manage.p.fun.ac.jp/server_manage"); err != nil {
			log.Fatalf("Failed to access to FUN VM WebAPI:%v\n", err)
		}

		time.Sleep(1 * time.Second)

		// 入力ボックスにユーザ名・パスを打ち込む
		elem_user := page.FindByName("username")
		elem_pass := page.FindByName("password")
		elem_user.Fill(username)
		elem_pass.Fill(password)
		log.Printf("fill username: %v\n", username)
		log.Printf("fill password\n")

		// Submit
		if err := page.FindByClass("credentials_input_submit").Click(); err != nil {
			log.Fatalf("Failed to login:%v\n", err)
			return
		}
		log.Printf("Success to login FUN VM WebAPI!!\n")

		time.Sleep(1 * time.Second)

		// 環境画面の項目数を入れる関数。暫定５個に設定しておく
		max_environment := 5
		for i := 1; i <= max_environment; i++ {

			log.Printf("Serch for environment: %v\n...", environment)
			text, _ := page.FindByXPath("/html/body/div/div/main/div/form/div[1]/div/select/option[" + strconv.Itoa(i) + "]").Text()
			if text == environment {

				log.Printf("get environment: %v\n", text)
				if err := page.FindByXPath("/html/body/div/div/main/div/form/div[1]/div/select/option[" + strconv.Itoa(i) + "]").Click(); err != nil {
					log.Fatalf("Failed to click environment: %v\n", err)
				}
				break
			}
			//　max_environment個分のの項目をチェックしてなかった場合エラーにする
			if i == max_environment {
				log.Fatalf("Can't look up environment: %v\n", environment)
			}
		}

		// 次のページへ行く
		if err := page.FindByXPath("/html/body/div/div/main/div/form/div[2]/div/span").Click(); err != nil {
			log.Fatalf("faild to click next page bottuon")
		}

		time.Sleep(1 * time.Second)

		//ダウンロードボタンクリック
		if err := page.FindByXPath("/html/body/form/div/div[4]/div[1]/div[3]/div/div[2]/div[2]/div/a").Click(); err != nil {
			log.Fatalf("Failed to Download key:%v\n", err)
			return
		}

		key_name, _ := page.FindByXPath("/html/body/form/div/div[4]/div[1]/div[3]/div/div[2]/div[2]/div/span").Text()
		log.Printf("key name is %v and  address is %v/%v.pem\n", key_name, address, key_name)
		address = address + "/" + key_name + ".pem"
		page.CloseWindow()
	}
	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, address))
	return
}
