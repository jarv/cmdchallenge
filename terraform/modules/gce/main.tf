variable "num_instances" {}
variable "name" {}
variable "ca_pem_fname" {}
variable "machine_type" {}
variable "use_static_ip" {}
variable "preemptible" {}
variable "automatic_restart" {}
variable "ssh_public_key" {}
variable "ssh_private_key" {}

resource "google_compute_address" "external" {
  count = var.use_static_ip ? var.num_instances : 0
  name = format(
    "%v-%02d",
    var.name,
    count.index + 1
  )
}

locals {
  external_ips = concat(google_compute_address.external.*.address, [""])
}

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
    ports    = ["2376", "22", "9090"]
  }

  source_ranges = ["0.0.0.0/0"]

  target_tags = ["default"]
}

resource "google_compute_instance" "default" {
  allow_stopping_for_update = true
  count                     = var.num_instances
  name = format(
    "%v-%02d",
    var.name,
    count.index + 1
  )

  machine_type = var.machine_type

  metadata_startup_script = file("${path.module}/bootstrap.bash")

  metadata = {
    ssh-keys = "jarv:${file(var.ssh_public_key)}"
    fqdn = format(
      "%v-%02d.gcp.cmdchallenge.com",
      var.name,
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
    # access_config {
    #   // ephemeral
    # }
    dynamic "access_config" {
      for_each = var.use_static_ip ? [] : [0]
      content {
        // ephemeral
        // nat_ip = var.use_static_ip ? element(local.external_ips, count.index) : ""
      }
    }

    dynamic "access_config" {
      for_each = var.use_static_ip ? [0] : []
      content {
        nat_ip = element(local.external_ips, count.index)
      }
    }
  }

  scheduling {
    preemptible       = var.preemptible
    automatic_restart = var.automatic_restart
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
    private_key = file(var.ssh_private_key)
    script_path = "/tmp/bootstrap.sh"
  }

  provisioner "local-exec" {
    command = "${path.root}/../bin/create-server-keys ${self.metadata.fqdn}"
  }

  provisioner "file" {
    source      = "${path.root}/../cmdchallenge/ro_volume"
    destination = "/var/tmp/"
  }

  provisioner "file" {
    source      = "${path.root}/../docker_cfg_files"
    destination = "/var/tmp/"
  }

  provisioner "file" {
    source      = "${path.root}/../private/server/${self.metadata.fqdn}"
    destination = "/var/tmp/"
  }

  provisioner "file" {
    source      = var.ca_pem_fname
    destination = "/var/tmp/ca.pem"
  }

  provisioner "remote-exec" {
    script = "${path.root}/modules/gce/bootstrap.bash"
    connection {
      host        = self.network_interface.0.access_config.0.nat_ip
      type        = "ssh"
      user        = "jarv"
      timeout     = "10m"
      private_key = file(var.ssh_private_key)
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
