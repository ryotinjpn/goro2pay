# Unit B Scheduler Lambda 用 CloudWatch Log Group。
# Unit A と同方針 (retention 7 日)。
resource "aws_cloudwatch_log_group" "scheduler" {
  name              = local.scheduler_log_group_name
  retention_in_days = 7

  tags = { Unit = "budget" }
}
