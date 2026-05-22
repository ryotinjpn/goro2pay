variable "env" {
  type        = string
  description = "環境識別子"
}

# API Gateway 連携 (lambda_permission に必要)
variable "api_gateway_execution_arn" {
  type        = string
  description = "API Gateway execution ARN (api_gateway module の output)"
}

# Cognito 連携 (API Lambda env として注入)
variable "cognito_user_pool_id" {
  type        = string
  description = "Cognito User Pool ID (Lambda env)"
}

variable "cognito_user_pool_client_id" {
  type        = string
  description = "Cognito App Client ID (Lambda env)"
}

# CD パイプライン用
variable "codestar_connection_arn" {
  type        = string
  description = "GitHub CodeStar Connection ARN (envs で作成、Amplify と共有)"
}

variable "github_owner" {
  type        = string
  description = "GitHub オーナー"
}

variable "github_repo_name" {
  type        = string
  description = "GitHub リポジトリ名"
}

variable "github_branch" {
  type        = string
  description = "API Lambda CD のソースブランチ"
}
