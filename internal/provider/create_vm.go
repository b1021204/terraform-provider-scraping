package provider

import (
	"github.com/sclevine/agouti"
	"log"
	"os"
	"strconv"
	"time"
)

func create_vm(Machine_Data *Machine_Data) {

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
	elem_user.Fill(Machine_Data.username)
	elem_pass.Fill(Machine_Data.password)
	log.Printf("fill username: %v\n", Machine_Data.username)
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

		log.Printf("Serch for environment: %v\n...", Machine_Data.environment)
		text, _ := page.FindByXPath("/html/body/div/div/main/div/form/div[1]/div/select/option[" + strconv.Itoa(i) + "]").Text()
		if text == Machine_Data.environment {

			log.Printf("get environment: %v\n", text)
			if err := page.FindByXPath("/html/body/div/div/main/div/form/div[1]/div/select/option[" + strconv.Itoa(i) + "]").Click(); err != nil {
				log.Fatalf("Failed to click environment: %v\n", err)
			}
			break
		}
		//　max_environment個分のの項目をチェックしてなかった場合エラーにする
		if i == max_environment {
			log.Fatalf("Can't look up environment: %v\n", Machine_Data.environment)
		}
	}

	// 次のページへ行く
	if err := page.FindByXPath("/html/body/div/div/main/div/form/div[2]/div/span").Click(); err != nil {
		log.Fatalf("faild to click next page bottuon")
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

	time.Sleep(1 * time.Second)

	// 作成できるマシンの最大値
	max_machine := 7
	for i := max_machine; i > 0; i-- {
		if name, err := page.FindByID("INSTANCE_NAME_" + strconv.Itoa(i)).Text(); err == nil {
			// 作成したマシンの名前を保存する

			if err == nil {
				log.Printf("machine_name = %v", name)
				Machine_Data.machine_name = name

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

				log.Printf("save new machine_name! machine_name is %v\n", name)

				break

			} else {
				log.Fatalf("Failed to save machine_name;%v\n", err)
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
