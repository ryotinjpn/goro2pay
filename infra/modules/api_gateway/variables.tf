variable "env" {
  type        = string
  description = "環境識別子"
}

variable "region" {
  type        = string
  description = "AWS Region (Cognito JWKS URL の組立に必要)"
}

# cognito module から渡される値
variable "cognito_user_pool_id" {
  type        = string
  description = "Cognito User Pool ID (JWT issuer URL 組立)"
}

variable "cognito_user_pool_client_id" {
  type        = string
  description = "Cognito App Client ID (JWT audience)"
}

# lambda_api module から渡される値 (Unit A 範囲: Logout 用)
variable "api_lambda_invoke_arn" {
  type        = string
  description = "API Lambda invoke ARN (integration target)"
}
