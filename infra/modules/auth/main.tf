# Auth Unit Terraform module の main 定義。
# データソース等の共通要素を集約。リソースは責務別の *.tf ファイルに分割している。

terraform {
  required_version = ">= 1.10.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.70"
    }
  }
}

data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

locals {
  account_id = data.aws_caller_identity.current.account_id
  region     = data.aws_region.current.name
  # リソース命名 prefix (Q-I4=B: gp-{env}-{resource})
  prefix = "gp-${var.env}"
}
