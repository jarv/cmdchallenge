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

provider "archive" {
  version = "~> 1.3"
}

provider "null" {
  version = "~> 2.1"
}

provider "external" {
  version = "~> 1.2"
}

terraform {
  backend "s3" {
    bucket  = "terraform-cmdchallenge"
    region  = "us-east-1"
    profile = "cmdchallenge-cicd"
    key     = "cicd"
  }
}

output "invoke-url" {
  value = module.api.invoke_url
}

output "test-hello-world" {
  value = "curl '${module.api.invoke_url}/?cmd=echo+hello+world&challenge_slug=hello_world'"
}

data "external" "short-sha" {
  program = ["sh", "short-sha.sh"]
}

output "instance-fqdn" {
  value = module.gce.public_dns
}

locals {
  is_prod             = terraform.workspace == "prod" ? "yes" : "no"
  timestamp           = timestamp()
  timestamp_sanitized = replace(local.timestamp, "/[- TZ:]/", "")
  name                = "${terraform.workspace}-cmdchallenge"
  short_sha           = data.external.short-sha.result.short_sha
}

# Hack to assert if the terraform workspace
# is set to default
# https://github.com/hashicorp/terraform/issues/15469#issuecomment-507689324
resource "null_resource" "assert_workspace" {
  triggers = terraform.workspace != "default" ? {} : file("Default workspace not allowed")
  lifecycle {
    ignore_changes = [
      triggers
    ]
  }
}

provider "aws" {
  region                  = "us-east-1"
  shared_credentials_file = pathexpand("~/.aws/credentials")
  profile                 = "cmdchallenge-cicd"
  version                 = "~> 2.59"
}

provider "google" {
  credentials = file(var.GCP_CRED_JSON_FNAME)
  project     = "cmdchallenge-1"
  region      = "us-east1"
  version     = "~> 3.39"
}

data "aws_caller_identity" "current" {
}

resource "null_resource" "generate_client_keys" {
  triggers = {
    build_number = timestamp()
  }
  provisioner "local-exec" {
    command = "${path.root}/../bin/create-client-keys"
  }
}

resource "null_resource" "copy_files_for_lambda" {
  triggers = {
    build_number = timestamp()
  }
  provisioner "local-exec" {
    command = "${path.root}/../bin/copy-files-for-lambda"
  }
  depends_on = [null_resource.generate_client_keys]
}

data "archive_file" "lambda_runcmd_zip" {
  type        = "zip"
  source_dir  = "../lambda_src/runcmd"
  output_path = "lambda-runcmd.zip"
  depends_on  = [null_resource.copy_files_for_lambda, null_resource.generate_client_keys]
}

data "archive_file" "lambda_runcmd_cron_zip" {
  type        = "zip"
  source_dir  = "../lambda_src/runcmd_cron"
  output_path = "lambda-runcmd-cron.zip"
  depends_on  = [null_resource.copy_files_for_lambda, null_resource.generate_client_keys]
}

module "dynamo" {
  source  = "./modules/dynamo"
  is_prod = local.is_prod
  name    = "${local.name}-db"
}

module "api" {
  source     = "./modules/api"
  lambda_arn = module.lambda.arn
  name       = "${local.name}-api"
}

module "lambda" {
  source                 = "./modules/lambda"
  submissions_table_name = module.dynamo.submissions_table_name
  commands_table_name    = module.dynamo.commands_table_name

  ec2_public_dns = module.gce.public_dns.0
  code_base64    = data.archive_file.lambda_runcmd_zip.output_base64sha256
  code_fname     = data.archive_file.lambda_runcmd_zip.output_path
  is_prod        = local.is_prod
  name           = "${local.name}-lambda"
  short_sha      = local.short_sha
}

module "lambda-cron" {
  source                 = "./modules/lambda-cron"
  num_shards             = 10
  submissions_table_name = module.dynamo.submissions_table_name
  commands_table_name    = module.dynamo.commands_table_name
  code_base64            = data.archive_file.lambda_runcmd_cron_zip.output_base64sha256
  code_fname             = data.archive_file.lambda_runcmd_cron_zip.output_path
  bucket_name            = local.is_prod == "yes" ? "cmdchallenge.com" : "testing.cmdchallenge.com"
  name                   = "${local.name}-lambda-cron"
}

module "gce" {
  num_instances     = 1
  source            = "./modules/gce"
  name              = local.is_prod == "yes" ? local.name : format("%v-%v", local.name, local.timestamp_sanitized)
  machine_type      = local.is_prod == "yes" ? "e2-micro" : "f1-micro"
  ca_pem_fname      = var.CA_PEM_FNAME
  use_static_ip     = local.is_prod == "yes" ? true : false
  preemptible       = local.is_prod == "yes" ? false : true
  automatic_restart = local.is_prod == "yes" ? true : false
  ssh_private_key   = var.SSH_PRIVATE_KEY
  ssh_public_key    = var.SSH_PUBLIC_KEY
}
