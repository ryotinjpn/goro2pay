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
    variables = merge(
      {
        AWS_LWA_PORT                 = "8080"
        LOG_LEVEL                    = "info"
        COGNITO_USER_POOL_ID         = var.cognito_user_pool_id
        COGNITO_APP_CLIENT_ID        = var.cognito_user_pool_client_id
        AWS_LWA_READINESS_CHECK_PATH = "/health"
        # Unit C: Bedrock Inference Profile (NFRC-C20 / Q-I10 = A)
        BEDROCK_INFERENCE_PROFILE_ID = "jp.anthropic.claude-haiku-4-5-20251001-v1:0"
      },
      # Unit C: OrderHistory テーブル名は env 経由で注入。空文字なら省略
      # (lambda 側の os.Getenv が "" を返すと repo 層が err を出すため、
      #  Unit C 未配線環境では設定を入れない)
      var.order_history_table_name != "" ? {
        ORDER_HISTORY_TABLE_NAME = var.order_history_table_name
      } : {},
    )
  }

  # CodeBuild が aws lambda update-function-code で image_uri を上書き
  lifecycle {
    ignore_changes = [image_uri]
  }
}

# 注意: API Gateway → API Lambda の invoke 許可 (aws_lambda_permission) は
# envs/dev/main.tf 側で作成する。本 module で持つと
# api_gateway.execution_arn → lambda_api、lambda_api.function_name → api_gateway
# の双方向依存が発生して module グラフが循環するため、permission resource は
# 「両 module の output を組み合わせる envs 側」に置く。

resource "aws_cloudwatch_log_group" "api" {
  name              = "/aws/lambda/${local.prefix}-api-fn"
  retention_in_days = 7
}
