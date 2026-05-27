# Lambda API module: API Lambda 本体 + ECR + CodePipeline + CodeBuild + S3 artifacts
# unit-of-work.md §4.1 の `infra/modules/lambda_api/` (Unit 横串、API Lambda + ECR) 定義に準拠。
# 他 Unit (B/C/D/E) は本 module の output (api_lambda_role_arn 等) に
# 権限を attach する形で機能を追加する。

# ----------------------------------------------------------------------------
# API Lambda (Go + Gin + LWA, container image)
# ----------------------------------------------------------------------------

resource "aws_lambda_function" "api" {
  function_name = "${local.prefix}-api-fn"
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.api.repository_url}:bootstrap"
  role          = aws_iam_role.api_lambda.arn

  memory_size   = 512
  timeout       = 30
  architectures = ["arm64"]

  environment {
    variables = merge(
      {
        AWS_LWA_PORT                 = "8080"
        LOG_LEVEL                    = "info"
        COGNITO_USER_POOL_ID         = var.cognito_user_pool_id
        COGNITO_APP_CLIENT_ID        = var.cognito_user_pool_client_id
        AWS_LWA_READINESS_CHECK_PATH = "/health"
        # Unit C: Bedrock Inference Profile (NFRC-C20 / Q-I10 = A)
        BEDROCK_INFERENCE_PROFILE_ID = "jp.anthropic.claude-haiku-4-5-20251001-v1:0"
      },
      # Unit C: OrderHistory テーブル名は env 経由で注入。空文字なら省略
      # (lambda 側の os.Getenv が "" を返すと repo 層が err を出すため、
      #  Unit C 未配線環境では設定を入れない)
      var.order_history_table_name != "" ? {
        ORDER_HISTORY_TABLE_NAME = var.order_history_table_name
      } : {},
      # Unit B: 4 テーブル名を env 経由で注入 (凍結 IF §10 の DDB_TABLE_* 命名)
      var.wallet_table_name != "" ? {
        DDB_TABLE_WALLET = var.wallet_table_name
      } : {},
      var.budget_settings_table_name != "" ? {
        DDB_TABLE_BUDGET_SETTINGS = var.budget_settings_table_name
      } : {},
      var.idempotency_keys_table_name != "" ? {
        DDB_TABLE_IDEMPOTENCY = var.idempotency_keys_table_name
      } : {},
      var.budget_reset_log_table_name != "" ? {
        DDB_TABLE_BUDGET_RESET_LOG = var.budget_reset_log_table_name
      } : {},
      # Unit D: Suggestion テーブル名を env 経由で注入 (凍結 IF §10 DDB_TABLE_SUGGESTION)
      var.suggestion_table_name != "" ? {
        DDB_TABLE_SUGGESTION = var.suggestion_table_name
      } : {},
    )
  }

  # CodeBuild が aws lambda update-function-code で image_uri を上書き
  lifecycle {
    ignore_changes = [image_uri]
  }
}

# 注意: API Gateway → API Lambda の invoke 許可 (aws_lambda_permission) は
# envs/dev/main.tf 側で作成する。本 module で持つと
# api_gateway.execution_arn → lambda_api、lambda_api.function_name → api_gateway
# の双方向依存が発生して module グラフが循環するため、permission resource は
# 「両 module の output を組み合わせる envs 側」に置く。

resource "aws_cloudwatch_log_group" "api" {
  name              = "/aws/lambda/${local.prefix}-api-fn"
  retention_in_days = 7
}

# ----------------------------------------------------------------------------
# ECR
# ----------------------------------------------------------------------------

resource "aws_ecr_repository" "api" {
  name                 = "${local.prefix}-api-image"
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  encryption_configuration {
    encryption_type = "AES256"
  }

  force_delete = true # dev のみ
}

resource "aws_ecr_lifecycle_policy" "api" {
  repository = aws_ecr_repository.api.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 5 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 5
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

# ----------------------------------------------------------------------------
# IAM Roles: api_lambda / codepipeline / codebuild
# ----------------------------------------------------------------------------

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

# Unit C / D / E が必要に応じて IAM Policy ARN を渡す (Q-I12 改定版)。
# Policy リソース定義は呼出元 module (modules/order_history, modules/bedrock 等) に
# 置き、本 module は attach のみを担う。
#
# for_each のキーは ARN 文字列ではなくインデックス (list の位置) を使う。
# ARN は呼び出し側で `module.X.policy_arn` 形式で渡されるため apply 時依存となり、
# toset(var.additional_policy_arns) では plan 段階でキーが確定せず以下のエラーになる:
#   The "for_each" set includes values derived from resource attributes that
#   cannot be determined until apply
# インデックスキーであれば list の長さ (= 静的) でキー集合が確定する。
resource "aws_iam_role_policy_attachment" "additional" {
  for_each   = { for i, arn in var.additional_policy_arns : tostring(i) => arn }
  role       = aws_iam_role.api_lambda.name
  policy_arn = each.value
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
        Resource = var.codestar_connection_arn
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
      {
        Effect   = "Allow"
        Action   = ["lambda:UpdateFunctionCode"]
        Resource = aws_lambda_function.api.arn
      },
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

# ----------------------------------------------------------------------------
# CodePipeline + CodeBuild (Q-I15=D): API Lambda CD
# ----------------------------------------------------------------------------

resource "aws_s3_bucket" "codepipeline_artifacts" {
  bucket        = "${local.prefix}-codepipeline-artifacts"
  force_destroy = true
}

resource "aws_s3_bucket_versioning" "codepipeline_artifacts" {
  bucket = aws_s3_bucket.codepipeline_artifacts.id
  versioning_configuration {
    status = "Disabled"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "codepipeline_artifacts" {
  bucket = aws_s3_bucket.codepipeline_artifacts.id

  rule {
    id     = "expire-old-artifacts"
    status = "Enabled"
    expiration {
      days = 30
    }
  }
}

resource "aws_codebuild_project" "api" {
  name         = "${local.prefix}-api-build"
  service_role = aws_iam_role.codebuild_api.arn

  artifacts {
    type = "CODEPIPELINE"
  }

  # ARM ネイティブビルダ。Lambda が arm64 / Go SDK v2 が大量モジュールを抱える
  # ため、x86_64 + QEMU クロスビルドだと build に 10 分超かかっていた。
  # ARM ネイティブにすると buildx --platform エミュレーションが不要となり
  # ~5x 高速化。compute_type も MEDIUM (7 GB / 4 vCPU) に底上げする。
  environment {
    compute_type    = "BUILD_GENERAL1_MEDIUM"
    image           = "aws/codebuild/amazonlinux2-aarch64-standard:3.0"
    type            = "ARM_CONTAINER"
    privileged_mode = true

    environment_variable {
      name  = "AWS_DEFAULT_REGION"
      value = var.region
    }
    environment_variable {
      name  = "ECR_REPOSITORY_URI"
      value = aws_ecr_repository.api.repository_url
    }
    environment_variable {
      name  = "LAMBDA_FUNCTION_NAME"
      value = aws_lambda_function.api.function_name
    }
  }

  source {
    type      = "CODEPIPELINE"
    buildspec = "apps/api/buildspec.yml"
  }

  logs_config {
    cloudwatch_logs {
      group_name = aws_cloudwatch_log_group.codebuild_api.name
    }
  }
}

resource "aws_cloudwatch_log_group" "codebuild_api" {
  name              = "/aws/codebuild/${local.prefix}-api-build"
  retention_in_days = 7
}

resource "aws_codepipeline" "api" {
  name     = "${local.prefix}-api-pipeline"
  role_arn = aws_iam_role.codepipeline_api.arn

  artifact_store {
    location = aws_s3_bucket.codepipeline_artifacts.bucket
    type     = "S3"
  }

  stage {
    name = "Source"

    action {
      name             = "Source"
      category         = "Source"
      owner            = "AWS"
      provider         = "CodeStarSourceConnection"
      version          = "1"
      output_artifacts = ["source_output"]

      configuration = {
        ConnectionArn    = var.codestar_connection_arn
        FullRepositoryId = "${var.github_owner}/${var.github_repo_name}"
        BranchName       = var.github_branch
      }
    }
  }

  # NOTE: 本 MVP では Deploy 専用 stage を持たず、CodeBuild の post_build フェーズ
  # で `aws lambda update-function-code` を実行する (buildspec.yml 参照)。
  # 「BuildAndDeploy」名でステージとアクションを表記し、Deploy 工程も含まれる
  # ことを明示する。本番化時は別途 Deploy stage を切り出す予定 (envs/prd/README)。
  stage {
    name = "BuildAndDeploy"

    action {
      name             = "BuildAndDeploy"
      category         = "Build"
      owner            = "AWS"
      provider         = "CodeBuild"
      version          = "1"
      input_artifacts  = ["source_output"]
      output_artifacts = ["build_output"]

      configuration = {
        ProjectName = aws_codebuild_project.api.name
      }
    }
  }
}
