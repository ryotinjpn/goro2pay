# Cognito module: User Pool + App Client + Pre Sign-up Lambda Trigger
# unit-of-work.md §4.1 の `infra/modules/cognito/` 定義に準拠。
# Auth Unit が所有する論理リソース群。

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
