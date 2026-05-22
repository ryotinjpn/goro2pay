# API Gateway module: HTTP API + JWT Authorizer + Stage + Logout/Health route
# unit-of-work.md §4.1 の `infra/modules/api_gateway/` (Unit 横串) 定義に準拠。
# Authorizer 設定値は cognito module の output を受け取る。
# Lambda integration は lambda_api module の output (invoke_arn) を受け取る。

terraform {
  required_version = ">= 1.10.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.46"
    }
  }
}

locals {
  prefix = "gp-${var.env}"
}
