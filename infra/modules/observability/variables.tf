variable "env" {
  description = "Environment name (dev/stg/prd)"
  type        = string
}

variable "alarm_email" {
  description = "Email address to receive CloudWatch alarm notifications via SNS"
  type        = string
}

variable "api_log_group_name" {
  description = "CloudWatch Log Group name of API Lambda (metric filter source)"
  type        = string
}

# しきい値は variables.tf で外部化 (Q-I4 = C)。dev/stg/prd で異なる値を tfvars で
# 注入可能にすることで、誤報多発時の調整を Terraform コード変更なしで行える。
variable "p95_threshold_ms" {
  description = "PlaceOrder p95 latency alarm threshold (NFRC-C13-1)"
  type        = number
  default     = 3000
}

variable "retry_threshold_count" {
  description = "Bedrock retry burst alarm threshold per 5 min (NFRC-C13-2)"
  type        = number
  default     = 5
}

variable "fallback_threshold_count" {
  description = "Fallback triggered burst alarm threshold per 5 min (NFRC-C13-3)"
  type        = number
  default     = 3
}

variable "bedrock_budget_limit_usd" {
  description = "Monthly Bedrock budget limit in USD (NFRC-C20)"
  type        = number
  default     = 5
}

# I-I5 修正: 旧 `var.tags` は envs から渡されないため削除済み。
# Project / Env / ManagedBy は provider.default_tags で全リソースに伝播するため、
# 各 module は `Unit = "observability"` のみ個別付与する (Unit A 既存 module 統一)。
