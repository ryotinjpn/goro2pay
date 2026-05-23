# CloudWatch Alarms / SNS / Budgets 検証 (Q-I13 = C)
# mock_provider で AWS API 呼出なし。

mock_provider "aws" {}

variables {
  env                = "dev"
  alarm_email        = "alerts@example.com"
  api_log_group_name = "/aws/lambda/gp-dev-api-fn"
  tags               = { Project = "goro2pay", Env = "dev", ManagedBy = "terraform" }
}

run "sns_topic_basic" {
  command = plan

  assert {
    condition     = aws_sns_topic.alarms.name == "gp-dev-alarms"
    error_message = "SNS Topic name must be gp-dev-alarms (Q-I5 = C 横串命名)"
  }

  assert {
    condition     = aws_sns_topic_subscription.alarm_email.protocol == "email"
    error_message = "Subscription protocol must be email (Q-I6 = A)"
  }

  assert {
    condition     = aws_sns_topic_subscription.alarm_email.endpoint == "alerts@example.com"
    error_message = "Subscription endpoint must come from var.alarm_email"
  }
}

run "alarm_thresholds_default" {
  command = plan

  assert {
    condition     = aws_cloudwatch_metric_alarm.place_order_p95_breach.threshold == 3000
    error_message = "p95 alarm threshold must default to 3000 ms (NFRC-C13-1)"
  }

  assert {
    condition     = aws_cloudwatch_metric_alarm.bedrock_retry_burst.threshold == 5
    error_message = "Bedrock retry alarm threshold must default to 5 / 5 min (NFRC-C13-2)"
  }

  assert {
    condition     = aws_cloudwatch_metric_alarm.fallback_triggered_burst.threshold == 3
    error_message = "Fallback alarm threshold must default to 3 / 5 min (NFRC-C13-3)"
  }
}

run "alarm_actions_point_to_sns" {
  command = plan

  assert {
    condition     = length(aws_cloudwatch_metric_alarm.place_order_p95_breach.alarm_actions) == 1
    error_message = "p95 alarm must have exactly 1 alarm action"
  }

  assert {
    condition     = length(aws_cloudwatch_metric_alarm.bedrock_retry_burst.alarm_actions) == 1
    error_message = "retry alarm must have exactly 1 alarm action"
  }

  assert {
    condition     = length(aws_cloudwatch_metric_alarm.fallback_triggered_burst.alarm_actions) == 1
    error_message = "fallback alarm must have exactly 1 alarm action"
  }
}

run "budgets_limit_amount" {
  command = plan

  assert {
    condition     = aws_budgets_budget.bedrock.limit_amount == "5"
    error_message = "Bedrock budget limit must default to $5 (NFRC-C20 / Q-I7 = A)"
  }

  assert {
    condition     = aws_budgets_budget.bedrock.limit_unit == "USD"
    error_message = "Bedrock budget unit must be USD"
  }

  assert {
    condition     = aws_budgets_budget.bedrock.time_unit == "MONTHLY"
    error_message = "Bedrock budget time_unit must be MONTHLY"
  }
}

run "metric_filter_pattern" {
  command = plan

  assert {
    condition     = aws_cloudwatch_log_metric_filter.place_order_latency.pattern == "{ $.event = \"place_order_complete\" && $.latencyMs = * }"
    error_message = "place_order_latency pattern must match place_order_complete event"
  }

  assert {
    condition     = aws_cloudwatch_log_metric_filter.bedrock_retry.pattern == "{ $.event = \"bedrock_retry\" }"
    error_message = "bedrock_retry pattern must match bedrock_retry event"
  }

  assert {
    condition     = aws_cloudwatch_log_metric_filter.fallback_triggered.pattern == "{ $.event = \"fallback_triggered\" }"
    error_message = "fallback_triggered pattern must match fallback_triggered event"
  }
}
