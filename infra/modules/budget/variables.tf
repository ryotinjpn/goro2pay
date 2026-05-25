variable "env" {
  description = "Environment name (dev/stg/prd)"
  type        = string
}

variable "region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-1"
}

# Scheduler Lambda 用 bootstrap バイナリのパス。
# デフォルトはリポジトリ構造 (`infra/modules/budget` から 3 階層上) を前提とする。
# モノレポ移動・CI/CD で path.module 解決が異なる場合は呼び出し側で上書きする
# (Code Review Minor 12)。
variable "scheduler_bootstrap_path" {
  description = "Scheduler Lambda の bootstrap バイナリへのパス (terraform apply 前に make build で生成しておくこと、空文字列でリポジトリ構造デフォルト使用)"
  type        = string
  default     = ""
}
