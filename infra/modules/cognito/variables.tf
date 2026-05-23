variable "env" {
  type        = string
  description = "環境識別子 (例: dev, prd)"
}

variable "region" {
  type        = string
  description = "AWS Region (cognito リソース内では現状未参照だが、他 module と変数受口を統一するために宣言)"
  default     = "ap-northeast-1"
}
