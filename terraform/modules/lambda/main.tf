variable "submissions_table_name" {
}

variable "commands_table_name" {
}

variable "ec2_public_dns" {
}

variable "code_base64" {
}

variable "code_fname" {
}

variable "is_prod" {
}
variable "name" {
}

resource "aws_lambda_function" "default" {
  filename         = var.code_fname
  source_code_hash = var.code_base64
  function_name    = var.name
  role             = aws_iam_role.default.arn
  description      = "Lambda function for cmdchallenge - Managed by Terraform"
  handler          = "runcmd.handler"
  runtime          = "python3.7"
  timeout          = "20"

  environment {
    variables = {
      SUBMISSIONS_TABLE_NAME = var.submissions_table_name
      COMMANDS_TABLE_NAME    = var.commands_table_name
      DOCKER_EC2_DNS         = var.ec2_public_dns
    }
  }
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_policy" "default" {
  name   = var.name
  path   = "/"
  policy = data.aws_iam_policy_document.default.json
}

resource "aws_iam_role" "default" {
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

resource "aws_iam_role_policy_attachment" "default" {
  role       = aws_iam_role.default.name
  policy_arn = aws_iam_policy.default.arn
}

data "aws_iam_policy_document" "default" {
  statement {
    sid = "1"

    actions = [
      "logs:*",
    ]

    resources = [
      "arn:aws:logs:*:*:*",
    ]
  }

  statement {
    actions = [
      "dynamodb:*",
    ]

    resources = [
      "arn:aws:dynamodb:us-east-1:*:table/${var.submissions_table_name}",
    ]
  }

  statement {
    actions = [
      "dynamodb:*",
    ]

    resources = [
      "arn:aws:dynamodb:us-east-1:*:table/${var.commands_table_name}",
    ]
  }

  statement {
    actions = [
      "sns:*",
    ]

    resources = [
      "arn:aws:sns:*:*:*",
    ]
  }
}

output "arn" {
  value = aws_lambda_function.default.arn
}

