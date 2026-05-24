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
  description = "Amplify build_spec の YAML ファイル絶対パス (envs 側で path.root 起点で渡すこと)"
}

variable "codestar_connection_arn" {
  type        = string
  description = "GitHub CodeStar Connection ARN (envs で modules/codestar_connection から渡す)"
  default     = ""
  # NOTE: default = "" は Console での手動承認後にのみ Amplify が repo を解決する
  # 現行運用 (oauth_token / access_token と同様) との互換のため。Console 接続を
  # 完全に terraform 側に取り込む場合は、aws_amplify_app.web に
  # `connection_arn = var.codestar_connection_arn` を設定して default を外す。
}

variable "github_oauth_token" {
  type        = string
  description = "GitHub Personal Access Token (classic, scopes: admin:repo_hook + repo)。CreateApp 時に Amplify が webhook 登録するため必要。"
  sensitive   = true
  default     = ""
}
