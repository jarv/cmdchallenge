variable "GCP_CRED_JSON_FNAME" {
  type    = string
  default = "../private/google/cmdchallenge.json"
}

variable "CA_PEM_FNAME" {
  type    = string
  default = "../private/ca/ca.pem"
}

variable "SSH_PUBLIC_KEY" {
  type    = string
  default = "../private/ssh/cmd_rsa.pub"
}

variable "SSH_PRIVATE_KEY" {
  type    = string
  default = "../private/ssh/cmd_rsa"
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
      version = "~> 2.1"
    }

    archive = {
      version = "~> 1.3"
    }

    aws = {
      version = "~> 2.59"
    }

    google = {
      version = "~> 3.39"
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
  is_prod             = terraform.workspace == "prod" ? "yes" : "no"
  timestamp           = timestamp()
  timestamp_sanitized = replace(local.timestamp, "/[- TZ:]/", "")
  name                = "${terraform.workspace}-cmdchallenge"
  short_sha           = data.external.short-sha.result.short_sha
  index_clean         = data.external.index-clean.result.index_clean
}

data "assert_test" "workspace" {
  test  = terraform.workspace != "default"
  throw = "'default' workspace is not valid in this project"
}

data "assert_test" "index_clean" {
  test  = terraform.workspace == "testing" || local.index_clean == "yes"
  throw = "Local git index is not clean, commit changes before running Terraform!"
}

provider "aws" {
  region                  = "us-east-1"
  shared_credentials_file = pathexpand("~/.aws/credentials")
  profile                 = "cmdchallenge-cicd"
}

provider "google" {
  credentials = file(var.GCP_CRED_JSON_FNAME)
  project     = "cmdchallenge-1"
  region      = "us-east1"
}

data "aws_caller_identity" "current" {
}

module "ec2" {
  source         = "./modules/ec2"
  ssh_public_key = var.SSH_PUBLIC_KEY
  short_sha      = local.short_sha
}

output "public_ip" {
  value = module.ec2.public_ip
}

output "public_dns" {
  value = module.ec2.public_dns
}

output "prometheus" {
  value = "http://${module.ec2.public_dns}:9090"
}
