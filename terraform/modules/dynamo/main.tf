variable "is_prod" {}
variable "name" {}

resource "aws_dynamodb_table" "runcmd_commands" {
  name           = var.is_prod == "yes" ? "commands" : "${var.name}-commands"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "id"

  attribute {
    name = "correct_length"
    type = "N"
  }

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "challenge_slug"
    type = "S"
  }

  global_secondary_index {
    name            = "challenge_slug-correct_length-index"
    hash_key        = "challenge_slug"
    range_key       = "correct_length"
    write_capacity  = 1
    read_capacity   = var.is_prod == "yes" ? 20 : 1
    projection_type = "ALL"
  }

  tags = {
    Name        = var.is_prod == "yes" ? "commands" : "${var.name}-commands"
    Environment = terraform.workspace
  }
}

resource "aws_dynamodb_table" "runcmd_submissions" {
  name           = var.is_prod == "yes" ? "submissions" : "${var.name}-submissions"
  read_capacity  = 1
  write_capacity = 1
  hash_key       = "source_ip"
  range_key      = "create_time"

  attribute {
    name = "source_ip"
    type = "S"
  }

  attribute {
    name = "create_time"
    type = "N"
  }

  tags = {
    Name        = var.is_prod == "yes" ? "submissions" : "${var.name}-submissions"
    Environment = terraform.workspace
  }
}

output "submissions_table_name" {
  value = aws_dynamodb_table.runcmd_submissions.name
}

output "commands_table_name" {
  value = aws_dynamodb_table.runcmd_commands.name
}

