terraform {
    required_providers {
        scraping = {
            source = "registry.terraform.io/hashicorp/scraping"
        }
    }
}

provider "scraping" {
    
    username = "b1021204"
    password = "EPa6ouQ2"    

}

resource "Server_Manage_FUN_resource" "example"{
    instance_type = "a"
    username = "b10210204"
}
