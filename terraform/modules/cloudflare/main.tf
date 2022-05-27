terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 3.0"
    }
  }
}

variable "zone" {}
variable "value" {}
variable "names" {}

locals {
  is_prod = terraform.workspace == "prod" ? true : false
}

resource "cloudflare_record" "default" {
  for_each        = toset(var.names)
  zone_id         = cloudflare_zone.default.id
  name            = each.key
  value           = var.value
  type            = "A"
  ttl             = 1
  allow_overwrite = true
  proxied         = true
}

resource "cloudflare_record" "instance" {
  zone_id         = cloudflare_zone.default.id
  name            = "${terraform.workspace}.ec2"
  value           = var.value
  type            = "A"
  ttl             = 1
  allow_overwrite = true
  proxied         = false
}

resource "cloudflare_zone" "default" {
  zone = var.zone
}

resource "cloudflare_zone_settings_override" "default" {
  zone_id = cloudflare_zone.default.id
  settings {
    ssl = "full"
  }
}

# Google MX records

resource "cloudflare_record" "mx_aspmx" {
  count    = local.is_prod ? 1 : 0
  zone_id  = cloudflare_zone.default.id
  name     = var.zone
  value    = "aspmx.l.google.com"
  type     = "MX"
  priority = 1
}

resource "cloudflare_record" "mx_alt1" {
  count    = local.is_prod ? 1 : 0
  zone_id  = cloudflare_zone.default.id
  name     = var.zone
  value    = "alt1.aspmx.l.google.com"
  type     = "MX"
  priority = 5
}

resource "cloudflare_record" "mx_alt2" {
  count    = local.is_prod ? 1 : 0
  zone_id  = cloudflare_zone.default.id
  name     = var.zone
  value    = "alt2.aspmx.l.google.com"
  type     = "MX"
  priority = 5
}

resource "cloudflare_record" "mx_alt3" {
  count    = local.is_prod ? 1 : 0
  zone_id  = cloudflare_zone.default.id
  name     = var.zone
  value    = "alt3.aspmx.l.google.com"
  type     = "MX"
  priority = 10
}

resource "cloudflare_record" "mx_alt4" {
  count    = local.is_prod ? 1 : 0
  zone_id  = cloudflare_zone.default.id
  name     = var.zone
  value    = "alt4.aspmx.l.google.com"
  type     = "MX"
  priority = 10
}

output "public_dns" {
  value = cloudflare_record.instance.hostname
}
