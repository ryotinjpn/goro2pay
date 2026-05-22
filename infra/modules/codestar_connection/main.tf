# CodeStar Connection module
# CodePipeline (lambda_api) と Amplify Hosting (amplify) の双方が GitHub への
# 接続情報として参照する共有リソース。env ごとに 1 つだけ作成する。
#
# 初回 apply 後、AWS Console で手動承認が必要 (Pending → Available)。
# deployment-runbook.md §4 に承認手順を記載。

terraform {
  required_version = ">= 1.10.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.46"
    }
  }
}

locals {
  prefix = "gp-${var.env}"
}

resource "aws_codestarconnections_connection" "github" {
  name          = "${local.prefix}-github-conn"
  provider_type = "GitHub"
  tags = {
    Unit = "shared"
  }
}
