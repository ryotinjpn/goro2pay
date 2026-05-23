variable "env" {
  description = "Environment name (dev/stg/prd)"
  type        = string
}

variable "region" {
  description = "AWS region (ap-northeast-1 想定)"
  type        = string
}

variable "tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default     = {}
}
