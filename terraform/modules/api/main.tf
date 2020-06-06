variable "region" {}
variable "account_id" {}
variable "lambda_arn" {}
variable "is_prod" {}
variable "name" {}

resource "aws_api_gateway_rest_api" "default" {
  name = var.name
}

resource "aws_api_gateway_method" "default" {
  rest_api_id   = aws_api_gateway_rest_api.default.id
  resource_id   = aws_api_gateway_rest_api.default.root_resource_id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "default" {
  depends_on              = [aws_api_gateway_method.default]
  rest_api_id             = aws_api_gateway_rest_api.default.id
  resource_id             = aws_api_gateway_rest_api.default.root_resource_id
  http_method             = aws_api_gateway_method.default.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = "arn:aws:apigateway:${var.region}:lambda:path/2015-03-31/functions/${var.lambda_arn}/invocations"
}

resource "aws_lambda_permission" "default" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = var.lambda_arn
  principal     = "apigateway.amazonaws.com"
  source_arn    = "arn:aws:execute-api:${var.region}:${var.account_id}:${aws_api_gateway_rest_api.default.id}/*/${aws_api_gateway_method.default.http_method}/"
}

resource "aws_api_gateway_deployment" "default" {
  depends_on  = [aws_api_gateway_integration.default]
  rest_api_id = aws_api_gateway_rest_api.default.id
  stage_name  = "r"

  provisioner "local-exec" {
    command = "echo ${aws_api_gateway_deployment.default.invoke_url} > ${path.root}/${terraform.workspace}-runcmd-api-gateway-endpoint"
  }

  provisioner "local-exec" {
    when    = destroy
    command = "rm -f ${path.root}/${terraform.workspace}-runcmd-api-gateway-endpoint"
  }
}

output "invoke_url" {
  value = aws_api_gateway_deployment.default.invoke_url
}

