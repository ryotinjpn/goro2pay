# Pre Sign-up Lambda IAM Role (Q-I9=A: CloudWatch Logs インラインのみ)

resource "aws_iam_role" "pre_signup_lambda" {
  name = "${local.prefix}-presignup-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "pre_signup_lambda_logs" {
  name = "logs"
  role = aws_iam_role.pre_signup_lambda.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
      ]
      Resource = [
        aws_cloudwatch_log_group.pre_signup.arn,
        "${aws_cloudwatch_log_group.pre_signup.arn}:*",
      ]
    }]
  })
}
