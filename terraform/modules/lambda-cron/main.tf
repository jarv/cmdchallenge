variable "submissions_table_name" {
}

variable "commands_table_name" {
}

variable "code_base64" {
}

variable "code_fname" {
}

variable "bucket_name" {
}

variable "num_shards" {
}

variable "name" {
}

resource "aws_cloudwatch_event_rule" "default" {
  count = var.num_shards
  name = format(
    "%v-every-one-day-%02d",
    var.name,
    count.index + 1
  )
  description         = "Fires every one day"
  schedule_expression = "rate(1 day)"
}

resource "aws_cloudwatch_event_target" "default" {
  count     = var.num_shards
  rule      = aws_cloudwatch_event_rule.default[count.index].name
  target_id = "cmdchallenge_lambda_cron"
  arn       = aws_lambda_function.default[count.index].arn
}

resource "aws_lambda_permission" "default" {
  count         = var.num_shards
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.default[count.index].function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.default[count.index].arn
}

resource "aws_lambda_function" "default" {
  count            = var.num_shards
  filename         = var.code_fname
  source_code_hash = var.code_base64
  function_name = format(
    "%v-%02d",
    var.name,
    count.index + 1
  )
  role        = aws_iam_role.default.arn
  description = "Lambda cron function for cmdchallenge - Managed by Terraform"
  handler     = "runcmd_cron.handler"
  runtime     = "python3.7"
  timeout     = "300"

  environment {
    variables = {
      SUBMISSIONS_TABLE_NAME = var.submissions_table_name
      COMMANDS_TABLE_NAME    = var.commands_table_name
      BUCKET_NAME            = var.bucket_name
      SHARD_INDEX            = count.index
      NUM_SHARDS             = var.num_shards
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
      "*",
    ]

    resources = [
      "arn:aws:s3:::cmdchallenge.com/*",
      "arn:aws:s3:::testing.cmdchallenge.com/*",
    ]
  }

  statement {
    actions = [
      "dynamodb:*",
    ]

    resources = [
      "arn:aws:dynamodb:us-east-1:*:table/${var.commands_table_name}",
      "arn:aws:dynamodb:us-east-1:*:table/${var.commands_table_name}/index/*",
      "arn:aws:dynamodb:us-east-1:*:table/${var.submissions_table_name}",
      "arn:aws:dynamodb:us-east-1:*:table/${var.submissions_table_name}/index/*",
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
  value = aws_lambda_function.default.*.arn
}
