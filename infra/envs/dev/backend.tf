# Terraform S3 backend (Q-I6 改訂版: DynamoDB 不要、S3 ネイティブ Lock)
# 公式: https://developer.hashicorp.com/terraform/language/backend/s3

terraform {
  backend "s3" {
    bucket       = "gp-tfstate-dev"
    key          = "auth/terraform.tfstate"
    region       = "ap-northeast-1"
    encrypt      = true
    use_lockfile = true # Terraform 1.10+ の S3 ネイティブ Lock
  }
}
