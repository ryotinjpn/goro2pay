# Amplify module: Amplify Hosting (Next.js App Router SSR) + Branch + SSR IAM Role
# unit-of-work.md §4.1 の `infra/modules/amplify/` (PWA 配信、Unit A/B/C/D/E 共通) 定義に準拠。
# CodeStar Connection は envs/ 側で作成して ARN を渡す (lambda_api と共有)。

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
