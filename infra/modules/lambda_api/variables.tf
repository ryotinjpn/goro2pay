variable "env" {
  type        = string
  description = "環境識別子"
}

variable "region" {
  type        = string
  description = "AWS Region (CodeBuild の AWS_DEFAULT_REGION env に注入)"
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

# Unit C で追加: 後続 Unit が IAM policy ARN を attach するための受け口。
# Q-I12 改定版: Policy 定義は責務元 module、attach は本 module で for_each。
variable "additional_policy_arns" {
  type        = list(string)
  description = "API Lambda Role に追加 attach する IAM Policy ARN のリスト (Unit C/D/E が利用)"
  default     = []
}

# Unit C で追加: API Lambda の environment に注入する OrderHistory テーブル名。
variable "order_history_table_name" {
  type        = string
  description = "OrderHistory DynamoDB テーブル名 (空文字なら env 注入をスキップ)"
  default     = ""
}
