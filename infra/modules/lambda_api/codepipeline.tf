# CodePipeline + CodeBuild (Q-I15=D): API Lambda CD

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

  environment {
    compute_type    = "BUILD_GENERAL1_SMALL"
    image           = "aws/codebuild/standard:7.0"
    type            = "ARM_CONTAINER"
    privileged_mode = true

    environment_variable {
      name  = "AWS_DEFAULT_REGION"
      value = data.aws_region.current.name
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

data "aws_region" "current" {}

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

  stage {
    name = "Build"

    action {
      name             = "Build"
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
