terraform {
    required_providers {
        scraping = {
            source = "registry.terraform.io/hashicorp/scraping"
        }
    }
}

provider "scraping" {
   

}
/*
resource "scraping_resource" "example"{
    instance_type = "a"
        
    username = "b1021204"
    password = "EPa6ouQ2" 
    machine_name = "EC2-geotail-147189"

}
*/
resource "scraping_resource" "example2"{
    environment = "Linux(Ubuntu22.04LTS)(2024前期)"
        
    username = "b1021204"
    password = "EPa6ouQ2" 
    machine_name = ""

}

output "name" {
    value = scraping_resource.example2.machine_name
}