locals {
  env_prefix       = "gp-${var.env}"
  sns_topic_name   = "${local.env_prefix}-alarms"
  metric_namespace = "GoroPay/Order"

  # Budgets は固定の time_period_start が必要 (本 MVP は 2026-05 開始)。
  # NFRC-C20 月次予算 = $5 / 月 (Q-I7 = A 80% / 100%)。
  budget_start = "2026-05-01_00:00"

  tags = var.tags
}
