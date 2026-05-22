# API Lambda (Go + Gin + LWA, container image)
# 初期 image_uri は ECR :bootstrap タグ。CD 後は CodeBuild が更新するため
# Terraform 側は lifecycle.ignore_changes = [image_uri] (Q-I15)。

resource "aws_lambda_function" "api" {
  function_name = "${local.prefix}-api-fn"
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.api.repository_url}:bootstrap"
  role          = aws_iam_role.api_lambda.arn

  memory_size = 512
  timeout     = 30
  architectures = ["arm64"]

  environment {
    variables = {
      AWS_LWA_PORT             = "8080"
      LOG_LEVEL                = "info"
      COGNITO_USER_POOL_ID     = aws_cognito_user_pool.main.id
      COGNITO_APP_CLIENT_ID    = aws_cognito_user_pool_client.web.id
      AWS_LWA_READINESS_CHECK_PATH = "/health"
    }
  }

  # CodeBuild が aws lambda update-function-code で image_uri を上書きするため
  # Terraform 管理対象から外す
  lifecycle {
    ignore_changes = [image_uri]
  }
}

# API Gateway が API Lambda を invoke する許可
resource "aws_lambda_permission" "apigw_invoke_api" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.main.execution_arn}/*/*"
}
