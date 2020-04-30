variable "timestamp" {}
variable "num_instances" {}
variable "is_prod" {}
variable "name" {}
variable "CA_PEM_FNAME" {}

resource "google_compute_firewall" "default" {
  count = var.num_instances
  name = format(
    "%v-%02d",
    var.name,
    count.index + 1
  )
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["2376", "22"]
  }

  source_ranges = ["0.0.0.0/0"]

  target_tags = ["default"]
}

resource "google_compute_instance" "default" {
  count = var.num_instances
  name = format(
    "%v-%02d",
    var.name,
    count.index + 1
  )

  machine_type = "f1-micro"

  metadata_startup_script = file("${path.module}/bootstrap-cloud.bash")

  metadata = {
    ssh-keys = "jarv:${file("${path.root}/../private/ssh/cmd_rsa.pub")}"
    fqdn = format(
      "%v-%v-%02d.gcp.cmdchallenge.com",
      var.name,
      var.timestamp,
      count.index + 1
    )
  }

  project = "cmdchallenge-1"
  zone    = "us-east1-b"

  lifecycle {
    create_before_destroy = false
  }

  network_interface {
    network = "default"

    access_config {
      // Ephemeral IP
    }
  }

  scheduling {
    preemptible       = var.is_prod == "yes" ? false : true
    automatic_restart = var.is_prod == "yes" ? true : false
  }

  boot_disk {
    auto_delete = true

    initialize_params {
      # image = "coreos-stable"
      # gcloud compute images list  --project cos-cloud  --no-standard-images
      image = "cos-cloud/cos-stable-81-12871-96-0"
    }
  }

  tags = [
    "default",
  ]

  connection {
    host        = self.network_interface.0.access_config.0.nat_ip
    type        = "ssh"
    user        = "jarv"
    timeout     = "10m"
    private_key = file("${path.root}/../private/ssh/cmd_rsa")
    script_path = "/tmp/bootstrap.sh"
  }

  provisioner "local-exec" {
    command = "${path.root}/../bin/create-server-keys ${self.metadata.fqdn}"
  }

  provisioner "file" {
    source      = "${path.root}/../cmdchallenge/ro_volume"
    destination = "/tmp/"
  }

  provisioner "file" {
    source      = "${path.root}/../docker_cfg_files"
    destination = "/tmp/"
  }

  provisioner "file" {
    source      = "${path.root}/../private/server/${self.metadata.fqdn}"
    destination = "/tmp/"
  }

  provisioner "file" {
    source      = var.CA_PEM_FNAME
    destination = "/tmp/ca.pem"
  }

  provisioner "remote-exec" {
    script = "${path.root}/modules/gce/bootstrap.bash"
    connection {
      host        = self.network_interface.0.access_config.0.nat_ip
      type        = "ssh"
      user        = "jarv"
      timeout     = "10m"
      private_key = file("${path.root}/../private/ssh/cmd_rsa")
      script_path = "/var/lib/cloud/bootstrap/bootstrap.sh"
    }
  }
}

resource "aws_route53_record" "default" {
  count   = var.num_instances
  zone_id = "Z3TFJ1MMW7EJ7R"
  name    = google_compute_instance.default[count.index].metadata.fqdn
  type    = "A"
  ttl     = "300"
  records = [google_compute_instance.default[count.index].network_interface.0.access_config.0.nat_ip]
}

output "public_dns" {
  value = aws_route53_record.default.*.fqdn
}
