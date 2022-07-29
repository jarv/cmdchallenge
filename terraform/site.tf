variable "SSH_PUBLIC_KEY" {
  type    = string
  default = "../private/ssh/cmd_rsa.pub"
}

terraform {
  required_providers {
    assert = {
      source  = "bwoznicki/assert"
      version = "0.0.1"
    }

    external = {
      version = "~> 1.2"
    }
    null = {
      version = "~> 3.0"
    }
    archive = {
      version = "~> 1.3"
    }
  }

  backend "s3" {
    bucket  = "terraform-cmdchallenge"
    region  = "us-east-1"
    profile = "cmdchallenge-cicd"
    key     = "cicd"
  }
}

data "external" "short-sha" {
  program = ["sh", "short-sha.sh"]
}

data "external" "index-clean" {
  program = ["sh", "index-clean.sh"]
}

locals {
  is_prod             = terraform.workspace == "prod" ? true : false
  timestamp           = timestamp()
  timestamp_sanitized = replace(local.timestamp, "/[- TZ:]/", "")
  name                = "${terraform.workspace}-cmdchallenge"
}

data "assert_test" "workspace" {
  test  = terraform.workspace != "default"
  throw = "'default' workspace is not valid in this project"
}

provider "aws" {
  region                  = "us-east-1"
  shared_credentials_file = pathexpand("~/.aws/credentials")
  profile                 = "cmdchallenge-cicd"
}

data "aws_caller_identity" "current" {
}

module "ec2" {
  source         = "./modules/ec2"
  ssh_public_key = var.SSH_PUBLIC_KEY
}

module "cloudflare" {
  source = "./modules/cloudflare"
  zone   = local.is_prod ? "cmdchallenge.com" : "funformentals.com"
  value  = module.ec2.public_ip
  names  = ["@", "oops", "12days"]
}

output "public_ip" {
  value = module.ec2.public_ip
}

output "public_dns" {
  value = module.cloudflare.public_dns
}

output "prometheus" {
  value = "http://${module.cloudflare.public_dns}:9090"
}
