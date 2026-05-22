variable "env" {
  type        = string
  description = "環境識別子 (例: dev, prd)"
  validation {
    condition     = contains(["dev", "prd"], var.env)
    error_message = "env must be one of: dev, prd"
  }
}

variable "region" {
  type        = string
  description = "AWS Region"
  default     = "ap-northeast-1"
}

variable "github_owner" {
  type        = string
  description = "GitHub オーナー (例: ryotinjpn)。CodeStar Connection / Amplify が利用。"
}

variable "github_repo_name" {
  type        = string
  description = "GitHub リポジトリ名"
  default     = "goro2pay"
}

variable "github_branch" {
  type        = string
  description = "Frontend / API CD のソースブランチ"
  default     = "develop"
}
