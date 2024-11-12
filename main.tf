//variable "stop" { default =true  }
  

terraform {
    required_providers {
        scraping = {
            source = "registry.terraform.io/hashicorp/scraping"
        }
    }
}

provider "scraping" {}

resource "scraping_resource" "example"{
    environment = "Linux(Ubuntu22.04LTS)(2024前期)(10/31廃止)"
    username = "b1021204"
    password = "SAKURAskip108" 
    machine_name = "EC2-geotail-153037"
    machine_stop = true
    instance_type = "t4g.large"

}

/*
  connection {
    type     = "ssh"
    user     = "ubuntu"
    //password = "KKU1Sq8K"
    private_key = file("/Users/nsysk_0101/univ/b4/terraform-provider-scraping/funawskeyb1021204.pem")
    host     =  scraping_resource.example.ip
  }

     provisioner "remote-exec" {
        inline = [
      "echo The servers IP address is ??? >> a.txt"
        ]
  }
}
  




output "ip" {
    value = scraping_resource.example.ip
}


    output "machine_pass"{
      value = scraping_resource.example.machine_pass
    }
    
*/
