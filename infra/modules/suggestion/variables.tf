variable "env" {
  description = "Environment name (dev/stg/prd)"
  type        = string
}

# Project / Env / ManagedBy は provider.default_tags で全リソースに伝播するため、
# 各 module は `Unit = "..."` のみ個別付与する (Unit A/C 既存 module 統一)。
