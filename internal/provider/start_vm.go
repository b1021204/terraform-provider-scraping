package main

import (
	//"fmt"
	"github.com/sclevine/agouti"
	"log"
	"strconv"
	"time"
)

func start_vm(username string, password string, machine_name string) {
	// ブラウザはChromeを指定して起動
	driver := agouti.ChromeDriver(agouti.Browser("chrome"))
	if err := driver.Start(); err != nil {
		log.Fatalf("Failed to start driver:%v", err)
	}
	defer driver.Stop()
	page, err := driver.NewPage()
	if err != nil {
		log.Fatalf("Failed to open page:%v", err)
	} // go to login page
	if err := page.Navigate("https://manage.p.fun.ac.jp/server_manage"); err != nil {
		log.Fatalf("Failed to navigate:%v", err)
	}
	time.Sleep(1 * time.Second)

	elem_user := page.FindByName("username")
	elem_pass := page.FindByName("password")
	elem_user.Fill(username)
	elem_pass.Fill(password)
	// Submit
	if err := page.FindByClass("credentials_input_submit").Click(); err != nil {
		log.Fatalf("Failed to login:%v", err)
		return
	}
	time.Sleep(2 * time.Second)
	if err := page.FindByXPath("/html/body/div/div/main/div/form/div[2]/div/span").Click(); err != nil {
		log.Fatalf("Failed to choice:%v", err)
		return
	}
	time.Sleep(1 * time.Second)

	for i := 0; i < 20; i++ {
		instance_name := page.FindByID("INSTANCE_NAME_" + strconv.Itoa(i))

		// web上からterraformに指定されたmachine_nameと合致するものを探す
		if text, err := instance_name.Text(); err == nil {
			if text == machine_name {
				if err := page.FindByName("startBtn_" + strconv.Itoa(i)).Click(); err != nil {
					log.Fatalf("Failed to start;%v", err)
					return
				}
				if err := page.FindByXPath("/html/body/form/div/div[5]/div/div/div[3]/button[1]").Click(); err != nil {
					log.Fatalf("Failed to start:%v", err)
					return
				}
				return
			}
		}

	}
	page.CloseWindow()

}

func main() {
	username := "b1021204"
	password := "EPa6ouQ2"
	machine_name := "EC2-geotail-153025"
	//login(username, password)
	start_vm(username, password, machine_name)

}
