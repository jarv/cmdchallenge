variable "origin" {
}
variable "aliases" {
  type = list
}

resource "aws_cloudfront_distribution" "default" {
  origin {
    domain_name = var.origin
  }

  aliases = var.aliases
}
