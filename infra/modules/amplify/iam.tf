# Amplify SSR Compute IAM Role

resource "aws_iam_role" "amplify_ssr" {
  name = "${local.prefix}-amplify-ssr-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "amplify.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })
}

# Next.js SSR Compute がランタイム Logs を吐けるよう最小権限 inline policy を付与。
# 旧マネージドポリシー (AWSAmplifyServerSideRendering) は AWS から削除されており
# 参照すると "does not exist or is not attachable" で apply が失敗する。
resource "aws_iam_role_policy" "amplify_ssr_logs" {
  name = "${local.prefix}-amplify-ssr-logs"
  role = aws_iam_role.amplify_ssr.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
      ]
      Resource = "arn:aws:logs:*:*:log-group:/aws/amplify/*"
    }]
  })
}
