# Budgets の account_id 識別用 (cost_filter で必要になる場合があるため data 取得)
data "aws_caller_identity" "current" {}
