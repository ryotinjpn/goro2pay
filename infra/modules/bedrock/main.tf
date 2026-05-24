# Unit C / D 共有: Bedrock Claude Haiku 4.5 呼出用 IAM Policy
# Q-I3 = A (最小権限) / Q-I10 = A (Inference Profile 経由)
# Foundation Model + Inference Profile の両 ARN を許可する必要あり

resource "aws_iam_policy" "bedrock_inference" {
  name        = local.policy_name
  description = "Bedrock Claude Haiku 4.5 inference (InvokeModel + Converse, model + inference profile ARN limited)"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "BedrockInferenceClaudeHaiku"
        Effect = "Allow"
        Action = [
          "bedrock:InvokeModel",
          "bedrock:Converse",
        ]
        Resource = local.model_arns
      },
    ]
  })

  tags = { Unit = "bedrock" }
}
