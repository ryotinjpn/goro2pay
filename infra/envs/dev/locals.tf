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

  # Unit C 追加: CloudWatch Alarms 通知先メール (Q-I6 = A)。
  # 本番化時は env ごとの terraform.tfvars 等で個別アドレスを設定する想定。
  alarm_email = "alerts@example.com"
}
