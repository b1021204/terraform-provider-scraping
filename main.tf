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
    machine_name = "EC2-pollux-144224"
    machine_stop = false
    instance_type = "t4g.large"

  connection {
    type     = "ssh"
    user     = "ubuntu"
    password = "vQ7k1Kn6"
    private_key = file(provider::scraping::key("b1021204", "SAKURAskip108", "Linux(Ubuntu22.04LTS)(2024後期)", "/Users/nsysk_0101/univ/b4/terraform-provider-scraping"))
    host     =  scraping_resource.example.ip

  }

     provisioner "remote-exec" {
        inline = [
      "mkdir mid_seminar"
        ]
  }
}

  

