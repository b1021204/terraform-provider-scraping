package provider

import (
	"github.com/sclevine/agouti"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func create_vm(Machine_Data *Machine_Data) {

	//driver := agouti.ChromeDriver(agouti.Browser("chrome"))

	driver := agouti.ChromeDriver(
		agouti.ChromeOptions(
			"args", []string{
				"--headless",
				"--disavle-gpu",
			}),
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

	//指定されたインスタンスタイプを選択
	for j := 1; j <= 4; j++ {
		log.Printf("serch for instance_type = %v \n", Machine_Data.instance_type)
		if Machine_Data.instance_type != "" {
			instance, err := page.Find("#INSTANCE_TYPE_RUN > option:nth-child(" + strconv.Itoa(j) + ")").Text()
			if err != nil {
				log.Printf("instance_type err:%v\n", err)
			}
			//	a, _ := page.FindByXPath("/html/body/form/div/div[4]/div[2]/div[3]/table/tbody/tr[1]/td[2]/div/select/option[2]").Text()
			log.Printf("now, scraping...: %v\n", instance)
			if instance == Machine_Data.instance_type {
				log.Printf("Succece instance_type = %v\n", instance)
				if err := page.Find("#INSTANCE_TYPE_RUN > option:nth-child(" + strconv.Itoa(j) + ")").Click(); err != nil {
					log.Printf("Can't choice instance_type: %v\n", Machine_Data.instance_type)
					log.Fatalf("Pleace choeck instance_type\n")
					return
				}
				break
			}
		}

	}

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
			Machine_Data.machine_name = name
			log.Printf("you create %v\n", Machine_Data.machine_name)
			/*
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
			*/
			break

		} else {
			log.Printf("Can't find :%v\n", name)
		}

	}

	// ipアドレスを精査して,学内アドレスかグローバルIPアドレスかを判別
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

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
				log.Printf("machine pass = %v", Machine_Data.machine_pass)
				if univ_ip {
					Machine_Data.ip, _ = page.FindByID("copiable-ip_address-" + strconv.Itoa(i)).Text()

					log.Printf("machine's ip is %v\n", Machine_Data.ip)

				} else {
					Machine_Data.ip, _ = page.FindByID("copiable-public_ip_address-" + strconv.Itoa(i)).Text()

					log.Printf("machine's ip is %v\n", Machine_Data.ip)

				}
				//og.Printf("%v", Machine_Data.ip)
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
func main() {
	username := "b1021204"
	password := "EPa6ouQ2"
	var machine_name string
	//login(username, password)
	machine_name = create_vm(username, password, machine_name)

}*/
