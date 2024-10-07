package provider

import (
	"github.com/sclevine/agouti"
	"log"
	"os"
	"strconv"
	"time"
)

func create_vm(Machine_Data Machine_Data) {

	// ブラウザはChromeを指定して起動
	driver := agouti.ChromeDriver(
		agouti.ChromeOptions(
			"args", []string{
				"--headless",
				"--disavle-gpu",
			}),
	)
	log.Printf("Open Google Chorome...")

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
	log.Printf("Access to FUN VM WebAPI...")
	time.Sleep(1 * time.Second)

	elem_user := page.FindByName("username")
	elem_pass := page.FindByName("password")
	elem_user.Fill(Machine_Data.username)
	elem_pass.Fill(Machine_Data.password)
	// Submit
	if err := page.FindByClass("credentials_input_submit").Click(); err != nil {
		log.Fatalf("Failed to login:%v", err)
		return
	}
	log.Printf("Success to login FUN VM WebAPI!!")

	time.Sleep(1 * time.Second)
	for i := 1; i < 5; i++ {

		text, _ := page.FindByXPath("/html/body/div/div/main/div/form/div[1]/div/select/option[" + strconv.Itoa(i) + "]").Text()
		if text == Machine_Data.environment {

			log.Printf("発見:%s=%s\n", Machine_Data.environment, text)
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
	time.Sleep(1 * time.Second)
	if err := page.FindByName("createBtn").Click(); err != nil {
		log.Fatalf("Failed to create;%v", err)
		return
	}
	log.Printf("Now Creating...")

	if err := page.FindByXPath("/html/body/form/div/div[5]/div/div/div[3]/button[1]").Click(); err != nil {
		log.Fatalf("dismiss to create:%v", err)
		return
	}

	log.Printf("Success to create new machine!!")
	log.Printf("Save new machine_name...")
	time.Sleep(2 * time.Second)

	for i := 20; i > 0; i-- {
		if name, err := page.FindByID("INSTANCE_NAME_" + strconv.Itoa(i)).Text(); err == nil {
			// 作成したマシンの名前を保存する
			//machine_name = text
			name, err = page.FindByID("INSTANCE_NAME_" + strconv.Itoa(i)).Text()
			if err == nil {
				log.Printf("machine_name = " + name)

				f, err := os.Create(".machine_name.txt")
				if err != nil {
					log.Fatal(err)
				}
				defer f.Close()

				d := []byte(name)

				_, err = f.Write(d)
				if err != nil {
					log.Fatal(err)
				}
				log.Printf("save new machine_name! machine_name is %v", name)

				return

			} else {
				log.Fatalf("Failed to save machine_name;%v", err)
				return
			}

		}
	}
	page.CloseWindow()
}

/*
func main() {
	username := "b1021204"
	password := "EPa6ouQ2"
	var machine_name string
	//login(username, password)
	machine_name = create_vm(username, password, machine_name)

}*/
