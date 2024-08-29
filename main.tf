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
    environment = "Linux(Ubuntu22.04LTS)(2024前期)"
    username = "b1021204"
    password = "SAKURAskip108" 
    machine_name = "EC2-geotail-147041"
machine_stop = false
} 
