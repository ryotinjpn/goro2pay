locals {
  # Unit A 命名規則 gp-{env}-{resource} (Q-I4) に整合
  prefix = "gp-${var.env}"

  # DynamoDB テーブル名
  wallet_table_name           = "${local.prefix}-wallet"
  budget_settings_table_name  = "${local.prefix}-budget-settings"
  idempotency_table_name      = "${local.prefix}-idempotency-keys"
  budget_reset_log_table_name = "${local.prefix}-budget-reset-log"

  # Scheduler Lambda 関連
  scheduler_function_name     = "${local.prefix}-scheduler-fn"
  scheduler_log_group_name    = "/aws/lambda/${local.prefix}-scheduler-fn"
  scheduler_role_name         = "${local.prefix}-scheduler-role"
  eventbridge_scheduler_role  = "${local.prefix}-eventbridge-scheduler-role"
  monthly_reset_schedule_name = "${local.prefix}-monthly-reset"

  # API Lambda attach 用 DynamoDB Policy
  dynamodb_access_policy_name = "${local.prefix}-budget-dynamodb-policy"

  # Scheduler Lambda の bootstrap バイナリパス (Code Review Minor 12)
  bootstrap_path = var.scheduler_bootstrap_path != "" ? var.scheduler_bootstrap_path : "${path.module}/../../../apps/api/cmd/scheduler/bootstrap"
}
