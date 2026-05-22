# AWS Provider + default_tags (Q-I11)

provider "aws" {
  region = "ap-northeast-1"

  # 注意: Unit タグは default_tags に入れない。
  # 後続 Unit B/C/D/E のリソースが同じ envs/dev に追加された際、
  # provider レベルの default が override されないと「Unit=auth」が
  # 全リソースに伝播してコスト集計が壊れるため、Unit タグは
  # 各 module 内で個別に `tags = merge({Unit = "..."}, ...)` で付ける。
  default_tags {
    tags = {
      Project   = "goro2pay"
      Env       = "dev"
      ManagedBy = "terraform"
    }
  }
}

terraform {
  required_version = ">= 1.10.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.46"
    }
  }
}
