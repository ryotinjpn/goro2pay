# Cognito module: User Pool + App Client + Pre Sign-up Lambda Trigger
# unit-of-work.md §4.1 の `infra/modules/cognito/` 定義に準拠。
# Auth Unit が所有する論理リソース群。

# ----------------------------------------------------------------------------
# Cognito User Pool (LC-15) + App Client (LC-16)
# ----------------------------------------------------------------------------

resource "aws_cognito_user_pool" "main" {
  name = "${local.prefix}-userpool"

  username_attributes      = ["email"]
  auto_verified_attributes = ["email"]

  # NFR Design A-NFR-SEC-02 / Q-A2=B (8 文字、英大小+数字混在)
  password_policy {
    minimum_length    = 8
    require_uppercase = true
    require_lowercase = true
    require_numbers   = true
    require_symbols   = false
  }

  mfa_configuration = "OFF"

  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }

  admin_create_user_config {
    allow_admin_create_user_only = false
  }

  # Pre Sign-up Lambda Trigger (Q-A1=B auto-confirm)
  lambda_config {
    pre_sign_up = aws_lambda_function.pre_signup.arn
  }

  email_configuration {
    email_sending_account = "COGNITO_DEFAULT"
  }

  # Q-I13=A: dev では terraform destroy 可能にする
  deletion_protection = "INACTIVE"

  tags = {
    Name = "${local.prefix}-userpool"
  }
}

resource "aws_cognito_user_pool_client" "web" {
  name         = "${local.prefix}-appclient-web"
  user_pool_id = aws_cognito_user_pool.main.id

  generate_secret = false # PWA Public Client

  # SRP は平文パスワードを Cognito へ送らないため、TLS 誤設定時の被害幅が
  # USER_PASSWORD_AUTH より小さい (defense in depth)。Amplify Auth v6 の
  # signIn default も SRP。USER_PASSWORD_AUTH は許可しない。
  explicit_auth_flows = [
    "ALLOW_USER_SRP_AUTH",
    "ALLOW_REFRESH_TOKEN_AUTH",
  ]

  # NFR Design A-NFR-SEC-03 / Q-N1=B (8h / 30d)
  id_token_validity      = 8
  access_token_validity  = 8
  refresh_token_validity = 30
  token_validity_units {
    id_token      = "hours"
    access_token  = "hours"
    refresh_token = "days"
  }

  prevent_user_existence_errors = "ENABLED"
  enable_token_revocation       = true
}

# Cognito → Pre Sign-up Lambda の invoke 許可
resource "aws_lambda_permission" "cognito_invoke_pre_signup" {
  statement_id  = "AllowCognitoInvokePreSignup"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.pre_signup.function_name
  principal     = "cognito-idp.amazonaws.com"
  source_arn    = aws_cognito_user_pool.main.arn
}

# ----------------------------------------------------------------------------
# Pre Sign-up Lambda (LC-07): auto-confirm 用 Node.js
# Q-D7=A (Node.js) / Q-I3=A (archive_file)
# ----------------------------------------------------------------------------

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
}

resource "aws_cloudwatch_log_group" "pre_signup" {
  name              = "/aws/lambda/${local.prefix}-presignup-fn"
  retention_in_days = 7
}

# ----------------------------------------------------------------------------
# Pre Sign-up Lambda IAM Role (Q-I9=A: CloudWatch Logs インラインのみ)
# ----------------------------------------------------------------------------

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
