variable "env" {
  description = "Environment name (dev/stg/prd)"
  type        = string
}

variable "region" {
  description = "AWS region (ap-northeast-1 想定)"
  type        = string
}

# I-I5 修正: 旧 `var.tags` は envs から渡されないため削除済み。
# Project / Env / ManagedBy は provider.default_tags で全リソースに伝播するため、
# 各 module は `Unit = "..."` のみ個別付与する (Unit A 既存 module 統一)。
