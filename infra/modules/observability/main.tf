# 横串 Observability: SNS Topic + CloudWatch metric filter ×3 + Alarms ×3 + Budgets
# NFRC-C13 (3 アラーム) / NFRC-C20 (Budgets) / Q-I4 〜 Q-I7

# ----------------------------------------------------------------------------
# SNS Topic + Email Subscription (Q-I5 = C / Q-I6 = A)
# ----------------------------------------------------------------------------
resource "aws_sns_topic" "alarms" {
  name = local.sns_topic_name
  tags = merge(local.tags, { Unit = "observability" })
}

resource "aws_sns_topic_subscription" "alarm_email" {
  topic_arn = aws_sns_topic.alarms.arn
  protocol  = "email"
  endpoint  = var.alarm_email
}

# ----------------------------------------------------------------------------
# CloudWatch Logs Metric Filter ×3 (NFRC-C13)
# ----------------------------------------------------------------------------

# NFRC-C13-1: PlaceOrder p95 違反検知用 latencyMs 抽出
resource "aws_cloudwatch_log_metric_filter" "place_order_latency" {
  name           = "${local.env_prefix}-place-order-latency"
  log_group_name = var.api_log_group_name
  pattern        = "{ $.event = \"place_order_complete\" && $.latencyMs = * }"

  metric_transformation {
    name      = "PlaceOrderLatencyMs"
    namespace = local.metric_namespace
    value     = "$.latencyMs"
    unit      = "Milliseconds"
  }
}

# NFRC-C13-2: Bedrock リトライ発動 (level=WARN, event=bedrock_retry)
resource "aws_cloudwatch_log_metric_filter" "bedrock_retry" {
  name           = "${local.env_prefix}-bedrock-retry"
  log_group_name = var.api_log_group_name
  pattern        = "{ $.event = \"bedrock_retry\" }"

  metric_transformation {
    name          = "BedrockRetryCount"
    namespace     = local.metric_namespace
    value         = "1"
    default_value = "0"
    unit          = "Count"
  }
}

# NFRC-C13-3: フォールバック発動 (level=WARN, event=fallback_triggered)
resource "aws_cloudwatch_log_metric_filter" "fallback_triggered" {
  name           = "${local.env_prefix}-fallback-triggered"
  log_group_name = var.api_log_group_name
  pattern        = "{ $.event = \"fallback_triggered\" }"

  metric_transformation {
    name          = "FallbackTriggeredCount"
    namespace     = local.metric_namespace
    value         = "1"
    default_value = "0"
    unit          = "Count"
  }
}

# ----------------------------------------------------------------------------
# CloudWatch Alarms ×3 (NFRC-C13)
# ----------------------------------------------------------------------------

resource "aws_cloudwatch_metric_alarm" "place_order_p95_breach" {
  alarm_name          = "${local.env_prefix}-place-order-p95-breach"
  alarm_description   = "PlaceOrder p95 latency exceeds ${var.p95_threshold_ms}ms (NFRC-C13-1)"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  threshold           = var.p95_threshold_ms
  treat_missing_data  = "notBreaching"

  metric_query {
    id          = "p95"
    return_data = true
    metric {
      metric_name = "PlaceOrderLatencyMs"
      namespace   = local.metric_namespace
      period      = 300 # 5 min
      stat        = "p95"
    }
  }

  alarm_actions = [aws_sns_topic.alarms.arn]
  ok_actions    = [aws_sns_topic.alarms.arn]
  tags          = merge(local.tags, { Unit = "observability", Severity = "High" })
}

resource "aws_cloudwatch_metric_alarm" "bedrock_retry_burst" {
  alarm_name          = "${local.env_prefix}-bedrock-retry-burst"
  alarm_description   = "Bedrock retry events ≥ ${var.retry_threshold_count} per 5 min (NFRC-C13-2)"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "BedrockRetryCount"
  namespace           = local.metric_namespace
  period              = 300
  statistic           = "Sum"
  threshold           = var.retry_threshold_count
  treat_missing_data  = "notBreaching"

  alarm_actions = [aws_sns_topic.alarms.arn]
  ok_actions    = [aws_sns_topic.alarms.arn]
  tags          = merge(local.tags, { Unit = "observability", Severity = "Medium" })
}

resource "aws_cloudwatch_metric_alarm" "fallback_triggered_burst" {
  alarm_name          = "${local.env_prefix}-fallback-triggered-burst"
  alarm_description   = "Fallback triggered events ≥ ${var.fallback_threshold_count} per 5 min (NFRC-C13-3)"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "FallbackTriggeredCount"
  namespace           = local.metric_namespace
  period              = 300
  statistic           = "Sum"
  threshold           = var.fallback_threshold_count
  treat_missing_data  = "notBreaching"

  alarm_actions = [aws_sns_topic.alarms.arn]
  ok_actions    = [aws_sns_topic.alarms.arn]
  tags          = merge(local.tags, { Unit = "observability", Severity = "High" })
}

# ----------------------------------------------------------------------------
# AWS Budgets: Bedrock 月次予算 (NFRC-C20 / Q-I7 = A)
# ----------------------------------------------------------------------------

resource "aws_budgets_budget" "bedrock" {
  name              = "${local.env_prefix}-bedrock-budget"
  budget_type       = "COST"
  limit_amount      = tostring(var.bedrock_budget_limit_usd)
  limit_unit        = "USD"
  time_unit         = "MONTHLY"
  time_period_start = local.budget_start

  cost_filter {
    name   = "Service"
    values = ["Amazon Bedrock"]
  }

  # Q-I7 = A: 80% 警告
  notification {
    comparison_operator       = "GREATER_THAN"
    threshold                 = 80
    threshold_type            = "PERCENTAGE"
    notification_type         = "ACTUAL"
    subscriber_sns_topic_arns = [aws_sns_topic.alarms.arn]
  }
  # Q-I7 = A: 100% 警報
  notification {
    comparison_operator       = "GREATER_THAN"
    threshold                 = 100
    threshold_type            = "PERCENTAGE"
    notification_type         = "ACTUAL"
    subscriber_sns_topic_arns = [aws_sns_topic.alarms.arn]
  }
}
