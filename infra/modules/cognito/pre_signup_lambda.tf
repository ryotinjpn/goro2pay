# Pre Sign-up Lambda (LC-07): auto-confirm 用 Node.js
# Q-D7=A (Node.js) / Q-I3=A (archive_file)

data "archive_file" "pre_signup" {
  type        = "zip"
  source_file = "${path.module}/../../lambdas/pre-signup/index.js"
  output_path = "${path.module}/.terraform/tmp/pre-signup.zip"
}

resource "aws_lambda_function" "pre_signup" {
  function_name    = "${local.prefix}-presignup-fn"
  filename         = data.archive_file.pre_signup.output_path
  source_code_hash = data.archive_file.pre_signup.output_base64sha256

  runtime       = "nodejs20.x"
  handler       = "index.handler"
  architectures = ["arm64"]
  role          = aws_iam_role.pre_signup_lambda.arn

  memory_size = 128
  timeout     = 5

  environment {
    variables = {
      LOG_LEVEL = "info"
    }
  }
}

resource "aws_cloudwatch_log_group" "pre_signup" {
  name              = "/aws/lambda/${local.prefix}-presignup-fn"
  retention_in_days = 7
}
