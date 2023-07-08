terraform {
  required_providers {
    aws = {
      version = "~> 4.0"
    }
  }
}

variable "ssh_public_key" {}

locals {
  is_prod            = terraform.workspace == "prod" ? true : false
  release_bucket     = "${terraform.workspace}-cmd-release"
  backup_bucket      = "${terraform.workspace}-cmd-backups"
  bootstrap_fname    = "bootstrap.sh"
  serve_artifact     = "s3://${local.release_bucket}/serve"
  bootstrap_artifact = "s3://${local.release_bucket}/${local.bootstrap_fname}"
  dist_artifact      = "s3://${local.release_bucket}/dist.tar.gz"
  backup_artifact    = "s3://${local.backup_bucket}/db.sqlite3.bak.gz"
  bootstrap_content = templatefile("${path.module}/bootstrap.tpl", {
    serve_artifact  = local.serve_artifact
    backup_artifact = local.backup_artifact
    cmd_img_suffix  = local.is_prod ? "" : "-testing"
    cmd_extra_opts  = local.is_prod ? "-setRateLimit" : ""
    dist_artifact   = local.dist_artifact
  })
  user_data_content = templatefile("${path.module}/userdata.tpl", {
    bootstrap_artifact = local.bootstrap_artifact
    bootstrap_fname    = local.bootstrap_fname
    bootstrap_sha      = sha1(local.bootstrap_content)
    backup_artifact    = local.backup_artifact
  })

}

resource "aws_s3_bucket" "release" {
  bucket = local.release_bucket
  # force_destroy = true
  tags = {
    Env = terraform.workspace
  }
}

resource "aws_s3_bucket_acl" "release" {
  bucket = aws_s3_bucket.release.id
  acl    = "private"
}

resource "aws_s3_object" "bootstrap" {
  bucket  = aws_s3_bucket.release.bucket
  key     = local.bootstrap_fname
  content = local.bootstrap_content
}

resource "aws_s3_bucket" "backups" {
  bucket = local.backup_bucket
  # force_destroy = true

  tags = {
    Env = terraform.workspace
  }
}

resource "aws_s3_bucket_acl" "backups" {
  bucket = aws_s3_bucket.backups.id
  acl    = "private"
}

resource "aws_key_pair" "default" {
  key_name   = "${terraform.workspace}-cmd"
  public_key = file(var.ssh_public_key)
}

resource "aws_security_group" "default" {
  name        = "${terraform.workspace}-cmd"
  description = "Security group that allows ssh"

  dynamic "ingress" {
    for_each = toset([22, 443, 80, 9090])
    content {
      from_port   = ingress.value
      to_port     = ingress.value
      protocol    = "tcp"
      cidr_blocks = ["0.0.0.0/0"]
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_iam_role" "default" {
  name = "${terraform.workspace}-cmd"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  tags = {
    Env = terraform.workspace
  }
}

resource "aws_iam_instance_profile" "default" {
  name = "${terraform.workspace}-cmd"
  role = aws_iam_role.default.name
}

resource "aws_iam_role_policy" "default" {
  name = "${terraform.workspace}-cmd"
  role = aws_iam_role.default.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "s3:*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_instance" "default" {
  user_data            = local.user_data_content
  iam_instance_profile = aws_iam_instance_profile.default.name
  lifecycle {
    create_before_destroy = true
  }
  ami = "ami-087c17d1fe0178315" # base AMI
  # ami             = "ami-0c6e5082cb2f30d7f" # cmd-20210918
  instance_type   = "t3.micro"
  security_groups = [aws_security_group.default.name]
  key_name        = aws_key_pair.default.key_name

  tags = {
    Env = terraform.workspace
  }
}

resource "aws_eip" "default" {
  instance = aws_instance.default.id
}

output "public_ip" {
  value = aws_eip.default.public_ip
}
