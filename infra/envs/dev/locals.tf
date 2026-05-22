# dev env の固定値 locals
# 本 MVP は本番運用しない (requirements.md Q14) ため、env ごとの値は
# tfvars で外部注入せず env ごとに locals.tf で固定する。
# 本番化時は infra/envs/prd/locals.tf で別値を定義する。

locals {
  env           = "dev"
  region        = "ap-northeast-1"
  github_owner  = "ryotinjpn"
  github_repo   = "goro2pay"
  github_branch = "develop"
}
