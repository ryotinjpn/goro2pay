output "bedrock_policy_arn" {
  value       = aws_iam_policy.bedrock_inference.arn
  description = "Bedrock 呼出 IAM policy ARN (lambda_api/iam.tf で attach する。Unit D も同 ARN を再利用予定)"
}
