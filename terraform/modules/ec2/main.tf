variable "CA_PEM_FNAME" {}
variable "ws_name" {}
variable "is_prod" {}

data "aws_ami" "coreos" {
  most_recent = true

  filter {
    name   = "name"
    values = ["CoreOS-stable-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["595879546273"] # CoreOS
}

resource "null_resource" "pre_keypair" {
  provisioner "local-exec" {
    command = "${path.root}/../bin/create-ssh-keys"
  }
}

resource "aws_key_pair" "cmdchallenge" {
  key_name   = "cmdchallenge-${var.ws_name}"
  public_key = file("${path.root}/../private/ssh/cmd_rsa.pub")
  depends_on = [null_resource.pre_keypair]
}

resource "aws_instance" "runcmd" {
  # ami           = "${data.aws_ami.coreos.id}"
  lifecycle {
    create_before_destroy = true
  }

  ami             = "ami-ad593cbb"
  instance_type   = "t2.micro"
  security_groups = [aws_security_group.runcmd.name]
  key_name        = aws_key_pair.cmdchallenge.key_name

  tags = {
    Name        = "DockerRunCmd"
    Environment = var.ws_name
  }

  connection {
    host        = coalesce(self.public_ip, self.private_ip)
    type        = "ssh"
    user        = "core"
    private_key = file("${path.root}/../private/ssh/cmd_rsa")
  }

  provisioner "remote-exec" {
    inline = [
      "mkdir -p runcmd/private",
    ]
  }

  provisioner "file" {
    source      = "${path.root}/../cmdchallenge/ro_volume"
    destination = "runcmd"
  }

  provisioner "file" {
    source      = "${path.root}/../docker_cfg_files"
    destination = "runcmd"
  }

  provisioner "local-exec" {
    command = "${path.root}/../bin/create-ca-keys"
  }

  provisioner "local-exec" {
    command = "${path.root}/../bin/create-client-keys"
  }

  provisioner "local-exec" {
    command = "${path.root}/../bin/create-server-keys ${aws_instance.runcmd.public_dns}"
  }

  provisioner "file" {
    source      = var.CA_PEM_FNAME
    destination = "runcmd/private/ca.pem"
  }

  provisioner "file" {
    source      = "${path.root}/../private/server/${aws_instance.runcmd.public_dns}/server-cert.pem"
    destination = "runcmd/private/server-cert.pem"
  }

  provisioner "file" {
    source      = "${path.root}/../private/server/${aws_instance.runcmd.public_dns}/server-key.pem"
    destination = "runcmd/private/server-key.pem"
  }

  provisioner "remote-exec" {
    script = "${path.module}/bootstrap.bash"
  }

  provisioner "remote-exec" {
    inline = [
      "uname -a",
    ]
  }
}

resource "aws_security_group" "runcmd" {
  name        = "RunCmdSecurityGroup-${var.ws_name}"
  description = "Security group that allows ssh and connections for Docker"

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 2376
    to_port     = 2376
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

output "coreos_ami_id" {
  value = data.aws_ami.coreos.id
}

output "public_ip" {
  value = aws_instance.runcmd.public_ip
}

output "public_dns" {
  value = aws_instance.runcmd.public_dns
}

