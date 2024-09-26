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
    environment = "Linux(Ubuntu22.04LTS)(2024後期)"
    username = "b1021204"
    password = "SAKURAskip108" 
    machine_name = "EC2-pollux-153142"
    machine_stop = true

/*
  connection {
    type     = "ssh"
    user     = "ubuntu"
    password = "1OVXcWsd"
    private_key = file("/Users/nsysk01/univ/b4/terraform-provider-scraping/funawskeyb1021204.pem")
    host     = provider::scraping::ip("b1021204", "SAKURAskip108", "EC2-geotail-153037")
  }

     provisioner "remote-exec" {
        inline = [
      "echo The servers IP address is ??? >> a.txt"
        ]
  }
  */
} 
/*
output "ip" {
    value = provider::scraping::ip("b1021204", "SAKURAskip108", "EC2-geotail-153037")
}
*/


