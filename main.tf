terraform {
    required_providers {
        scraping = {
            source = "registry.terraform.io/hashicorp/scraping"
        }
    }
}

provider "scraping" {
    
/*    host = "https://sso.fun.ac.jp/my.policy"
    username = "b1021204"
    password = "EPa6ouQ2"
    
*/
}

resource "scraping_resource" "example"{
    instance_type = "a"
}
