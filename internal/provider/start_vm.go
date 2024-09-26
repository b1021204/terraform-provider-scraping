package provider

import (
	//"fmt"
	"github.com/sclevine/agouti"
	"log"
	"strconv"
	"time"
)

func start_vm(Machine_Data Machine_Data) {
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
	for i := 1; i < 5; i++ {

		text, _ := page.FindByXPath("/html/body/div/div/main/div/form/div[1]/div/select/option[" + strconv.Itoa(i) + "]").Text()
		if text == "Linux(Ubuntu22.04LTS)(2024後期)" {

			log.Printf("発見！！\n")
			if err := page.FindByXPath("/html/body/div/div/main/div/form/div[1]/div/select/option[" + strconv.Itoa(i) + "]").Click(); err != nil {
				log.Fatalf("Failed to choice:%v", err)
				return
			}

		}
	}
	if err := page.FindByXPath("/html/body/div/div/main/div/form/div[2]/div/span").Click(); err != nil {
		log.Fatalf("Failed to choice:%v", err)
		return
	}

	for i := 0; i < 20; i++ {
		log.Printf("serch for machin_name = %v", Machine_Data.machine_name)
		instance_name := page.FindByID("INSTANCE_NAME_" + strconv.Itoa(i))
		if i == 19 {
			log.Fatal("マシン名ねーじゃん！！！")
		}

		// web上からterraformに指定されたmachine_nameと合致するものを探す
		if text, err := instance_name.Text(); err == nil {
			if text == Machine_Data.machine_name {
				log.Printf("found machin_name = %v!!!", Machine_Data.machine_name)
				log.Printf("start %v...", Machine_Data.machine_name)

				if err := page.FindByName("startBtn_" + strconv.Itoa(i)).Click(); err != nil {
					log.Fatalf("Failed to start;%v", err)
					return
				}
				if err := page.FindByXPath("/html/body/form/div/div[5]/div/div/div[3]/button[1]").Click(); err != nil {
					log.Fatalf("Failed to start:%v", err)
					return
				}
				log.Printf("start %v!!", Machine_Data.machine_name)
				return

			}
		}

	}
	page.CloseWindow()

}

/*
デバック用コード
func main() {
	username := "b1021204"
	password := "SAKURAskip108"
	machine_name := "EC2-geotail-153025"
	start_vm(username, password, machine_name)
}
*/
