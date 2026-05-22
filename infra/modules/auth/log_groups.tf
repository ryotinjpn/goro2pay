# CloudWatch Log Groups (Q-I12=A: retention 7 日)

resource "aws_cloudwatch_log_group" "pre_signup" {
  name              = "/aws/lambda/${local.prefix}-presignup-fn"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_group" "api" {
  name              = "/aws/lambda/${local.prefix}-api-fn"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_group" "codebuild_api" {
  name              = "/aws/codebuild/${local.prefix}-api-build"
  retention_in_days = 7
}
