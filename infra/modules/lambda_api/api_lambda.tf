# API Lambda (Go + Gin + LWA, container image)

resource "aws_lambda_function" "api" {
  function_name = "${local.prefix}-api-fn"
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.api.repository_url}:bootstrap"
  role          = aws_iam_role.api_lambda.arn

  memory_size   = 512
  timeout       = 30
  architectures = ["arm64"]

  environment {
    variables = {
      AWS_LWA_PORT                 = "8080"
      LOG_LEVEL                    = "info"
      COGNITO_USER_POOL_ID         = var.cognito_user_pool_id
      COGNITO_APP_CLIENT_ID        = var.cognito_user_pool_client_id
      AWS_LWA_READINESS_CHECK_PATH = "/health"
    }
  }

  # CodeBuild が aws lambda update-function-code で image_uri を上書き
  lifecycle {
    ignore_changes = [image_uri]
  }
}

# API Gateway → API Lambda invoke 許可
resource "aws_lambda_permission" "apigw_invoke_api" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${var.api_gateway_execution_arn}/*/*"
}

resource "aws_cloudwatch_log_group" "api" {
  name              = "/aws/lambda/${local.prefix}-api-fn"
  retention_in_days = 7
}
