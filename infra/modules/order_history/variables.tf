variable "env" {
  description = "Environment name (dev/stg/prd)"
  type        = string
}

# I-I5 修正: 旧 `var.tags` は envs から渡されないため dead code だった。
# Project / Env / ManagedBy は provider.default_tags で全リソースに伝播するため、
# 各 module は `Unit = "..."` のみ merge() で個別付与する (Unit A 既存 module 統一)。
