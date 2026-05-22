module "auth" {
  source = "../../modules/auth"

  env              = "dev"
  region           = "ap-northeast-1"
  github_owner     = var.github_owner
  github_repo_name = "goro2pay"
  github_branch    = var.github_branch
}
