variable "github_owner" {
  type        = string
  description = "GitHub オーナー (例: ryotinjpn)。CodeStar Connection / Amplify が利用。terraform.tfvars で指定する。"
}

variable "github_branch" {
  type        = string
  description = "Frontend / API CD のソースブランチ"
  default     = "develop"
}
