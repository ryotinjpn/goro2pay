module "auth" {
  source = "../../modules/auth"

  env              = "dev"
  region           = "ap-northeast-1"
  github_owner     = local.github_owner
  github_repo_name = local.github_repo
  github_branch    = local.github_branch
}
