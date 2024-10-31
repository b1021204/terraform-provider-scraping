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
    machine_name = "EC2-geotail-155163"
    machine_stop = false
    instance_type = "t4g.large"


/*
  connection {
    type     = "ssh"
    user     = "ubuntu"
    password = provider::scraping::ip("b1021204", "SAKURAskip108", "Linux(Ubuntu22.04LTS)(2024前期)(10/31廃止)", "EC2-geotail-146055")
    private_key = file(provider::scraping::key("b1021204", "SAKURAskip108", "Linux(Ubuntu22.04LTS)(2024前期)(10/31廃止)", "/Users/nsysk_0101/univ/b4/terraform-provider-scraping"))
    host     = provider::scraping::ip("b1021204", "SAKURAskip108", "Linux(Ubuntu22.04LTS)(2024前期)(10/31廃止)", "EC2-geotail-146055")
  }

     provisioner "remote-exec" {
        inline = [
      "echo The servers IP address is ??? >> a.txt"
        ]
  }
  
} 


output "ip" {
    value = provider::scraping::ip("b1021204", "SAKURAskip108", "Linux(Ubuntu22.04LTS)(2024前期)(10/31廃止)", "EC2-geotail-146000")
}

*/


}