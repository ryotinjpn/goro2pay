variable "env" {
  type        = string
  description = "環境識別子"
}

variable "region" {
  type        = string
  description = "AWS Region (Frontend Amplify branch env 用)"
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
  description = "Frontend デプロイのソースブランチ (例: develop)"
}

# Auth + API 連携 env vars
variable "cognito_user_pool_id" {
  type        = string
  description = "Cognito User Pool ID (Browser に NEXT_PUBLIC_USER_POOL_ID として注入)"
}

variable "cognito_user_pool_client_id" {
  type        = string
  description = "Cognito App Client ID (Browser に NEXT_PUBLIC_USER_POOL_CLIENT_ID として注入)"
}

variable "api_endpoint" {
  type        = string
  description = "API Gateway 公開 URL (server-only env API_ENDPOINT として注入、BFF パターン)"
}

variable "amplify_yml_path" {
  type        = string
  description = "Amplify build_spec の YAML ファイルパス (project root からの相対)"
  default     = "../../../web/amplify.yml"
}
