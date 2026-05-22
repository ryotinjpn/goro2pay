# AWS Provider + default_tags (Q-I11)

provider "aws" {
  region = "ap-northeast-1"

  default_tags {
    tags = {
      Project   = "goro2pay"
      Env       = "dev"
      Unit      = "auth"
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
