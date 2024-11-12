package provider

import (
	"fmt"
	"github.com/sclevine/agouti"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func start_vm(Machine_Data *Machine_Data) {

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

	// Machine_data.machine_nameと合致するものをスクレイピングで探す
	//　作成できるマシンの最大個数を入れる関数を用意する
	max_machine := 5
	for i := 0; i <= max_machine; i++ {
		log.Printf("serch for machin_name = %v\n", Machine_Data.machine_name)
		instance_name := page.FindByID("INSTANCE_NAME_" + strconv.Itoa(i))

		// web上からterraformに指定されたmachine_nameと合致するものを探す
		if text, err := instance_name.Text(); err == nil {
			if text == Machine_Data.machine_name {
				log.Printf("found machin_name = %v!!!", Machine_Data.machine_name)
				log.Printf("start %v...\n", Machine_Data.machine_name)

				// インスタンスネームが指定されているときに、スクレイピングする

				for j := 0; j <= 4; j++ {
					log.Printf("serch for instance_type = %v \n", Machine_Data.instance_type)
					if Machine_Data.instance_type != "" {
						instance, _ := page.FindByXPath("/html/body/form/div/div[4]/div[2]/div[" + strconv.Itoa(i+1) + "]/table/tbody/tr[1]/td[2]/div/select/option[" + strconv.Itoa(j) + "]").Text()
						//	a, _ := page.FindByXPath("/html/body/form/div/div[4]/div[2]/div[3]/table/tbody/tr[1]/td[2]/div/select/option[2]").Text()
						log.Printf("now, scraping...: %v", instance)
						if instance == Machine_Data.instance_type {
							if err := page.FindByXPath("/html/body/form/div/div[4]/div[2]/div[" + strconv.Itoa(i+1) + "]/table/tbody/tr[1]/td[2]/div/select/option[" + strconv.Itoa(j) + "]").Click(); err != nil {
								log.Printf("Can't choice instance_type: %v\n", Machine_Data.instance_type)
								log.Fatalf("Pleace choeck instance_type\n")
								return
							}
							break
						}
					}

				}

				// 見つけたマシン名のスタートボタンをおす
				time.Sleep(1 * time.Second)
				if err := page.FindByName("startBtn_" + strconv.Itoa(i)).Click(); err != nil {
					log.Fatalf("Failed to start;%v\n", err)
					return
				}

				// 確認画面を進める（ボタンを押す）
				if err := page.FindByXPath("/html/body/form/div/div[5]/div/div/div[3]/button[1]").Click(); err != nil {
					log.Fatalf("Failed to start:%v\n", err)
					return
				}
				log.Printf("start %v!!\n", Machine_Data.machine_name)

			}

			// マシン名が見つからなかった場合、エラーにする
			if i == max_machine {
				log.Printf("Can't look up machinename: %vn", Machine_Data.machine_name)
				log.Fatal("Pleace cheack your machine_name\n")
			}
		}
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
		log.Printf("You use univ wifi like fun-wifi or free-wifi\n")
	} else {
		log.Printf("You don't use univ wifi\n")
	}

	max_machine = 5
	for i := 0; i <= max_machine; i++ {
		log.Printf("serch for machin_name = %v\n", Machine_Data.machine_name)
		instance_name := page.FindByID("INSTANCE_NAME_" + strconv.Itoa(i))

		// web上からterraformに指定されたmachine_nameと合致するものを探す
		if text, err := instance_name.Text(); err == nil {
			if text == Machine_Data.machine_name {

				//machine_passをスクレイピングする
				log.Printf("found machin_name = %v!!!\n", Machine_Data.machine_name)
				log.Printf("scraping %v...", Machine_Data.machine_name)
				Machine_Data.machine_pass, _ = page.FindByID("copiable-password-" + strconv.Itoa(i)).Text()
				log.Printf("%v", Machine_Data.machine_pass)
				if univ_ip {
					Machine_Data.ip, _ = page.FindByID("copiable-ip_address-" + strconv.Itoa(i)).Text()
				} else {
					Machine_Data.ip, _ = page.FindByID("copiable-public_ip_address-" + strconv.Itoa(i)).Text()
					log.Println(Machine_Data.ip + "\n\n")
				}
				log.Printf("%v", Machine_Data.ip)
				break
			}
		}
		if max_machine == i {
			log.Fatalf("Can't get machine_name")
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
