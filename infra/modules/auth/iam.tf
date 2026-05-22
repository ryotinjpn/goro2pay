# IAM Roles (Q-I9 / NFR Design 各 LC):
# - pre_signup_lambda: CloudWatch Logs 書込のみ
# - api_lambda:        当面 CloudWatch Logs 書込のみ (他 Unit B/C/D/E が DynamoDB / Bedrock 権限を後続で attach)
# - amplify_ssr:       Amplify SSR Compute role (logs 書込のみ)
# - codepipeline_api:  S3 artifact / CodeBuild start / CodeStar Connection 使用
# - codebuild_api:     CloudWatch Logs / ECR push / Lambda UpdateFunctionCode / S3 artifact 読書

# --- Pre Sign-up Lambda Role ---

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

# --- API Lambda Role ---

resource "aws_iam_role" "api_lambda" {
  name = "${local.prefix}-api-role"

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

resource "aws_iam_role_policy" "api_lambda_logs" {
  name = "logs"
  role = aws_iam_role.api_lambda.id

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
        aws_cloudwatch_log_group.api.arn,
        "${aws_cloudwatch_log_group.api.arn}:*",
      ]
    }]
  })
}

# --- Amplify SSR Role ---

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

resource "aws_iam_role_policy_attachment" "amplify_ssr_managed" {
  role       = aws_iam_role.amplify_ssr.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSAmplifyServerSideRendering"
}

# --- CodePipeline Role ---

resource "aws_iam_role" "codepipeline_api" {
  name = "${local.prefix}-codepipeline-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "codepipeline.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "codepipeline_api" {
  name = "policy"
  role = aws_iam_role.codepipeline_api.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:GetObjectVersion",
          "s3:PutObject",
          "s3:GetBucketVersioning",
        ]
        Resource = [
          aws_s3_bucket.codepipeline_artifacts.arn,
          "${aws_s3_bucket.codepipeline_artifacts.arn}/*",
        ]
      },
      {
        Effect   = "Allow"
        Action   = ["codestar-connections:UseConnection"]
        Resource = aws_codestarconnections_connection.github.arn
      },
      {
        Effect = "Allow"
        Action = [
          "codebuild:StartBuild",
          "codebuild:BatchGetBuilds",
        ]
        Resource = aws_codebuild_project.api.arn
      },
    ]
  })
}

# --- CodeBuild Role ---

resource "aws_iam_role" "codebuild_api" {
  name = "${local.prefix}-codebuild-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "codebuild.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy" "codebuild_api" {
  name = "policy"
  role = aws_iam_role.codebuild_api.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      # CloudWatch Logs
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
        ]
        Resource = [
          aws_cloudwatch_log_group.codebuild_api.arn,
          "${aws_cloudwatch_log_group.codebuild_api.arn}:*",
        ]
      },
      # ECR push (anyone can do GetAuthorizationToken)
      {
        Effect   = "Allow"
        Action   = ["ecr:GetAuthorizationToken"]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:PutImage",
          "ecr:InitiateLayerUpload",
          "ecr:UploadLayerPart",
          "ecr:CompleteLayerUpload",
          "ecr:BatchGetImage",
          "ecr:GetDownloadUrlForLayer",
        ]
        Resource = aws_ecr_repository.api.arn
      },
      # Lambda UpdateFunctionCode (Q-I15、対象を絞る)
      {
        Effect   = "Allow"
        Action   = ["lambda:UpdateFunctionCode"]
        Resource = aws_lambda_function.api.arn
      },
      # S3 artifact bucket
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:GetObjectVersion",
        ]
        Resource = "${aws_s3_bucket.codepipeline_artifacts.arn}/*"
      },
    ]
  })
}
