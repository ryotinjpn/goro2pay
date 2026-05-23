output "sns_topic_arn" {
  value       = aws_sns_topic.alarms.arn
  description = "SNS Topic ARN for alarm notifications (Unit D/E でも再利用可能)"
}
