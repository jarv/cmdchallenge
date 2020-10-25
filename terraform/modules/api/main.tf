variable "lambda_arn" {}
variable "name" {}

resource "aws_apigatewayv2_api" "default" {
  name          = var.name
  protocol_type = "HTTP"
  target        = var.lambda_arn
}

resource "aws_lambda_permission" "default" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = var.lambda_arn
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.default.execution_arn}/*/*"
}

output "invoke_url" {
  value = aws_apigatewayv2_api.default.api_endpoint
}
