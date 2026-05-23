# tflint 設定 (terraform-module-design 規約 §ディレクトリ構成)
#
# ローカル実行: cd infra && tflint --recursive
#
# CI への組込みは別 PR で実施予定 (現状 GitHub Actions の利用枠制限のため
# ローカル / 開発者環境での実行に留める)。

config {
  format = "compact"
  call_module_type = "local"
}

plugin "terraform" {
  enabled = true
  preset  = "recommended"
}

plugin "aws" {
  enabled = true
  version = "0.32.0"
  source  = "github.com/terraform-linters/tflint-ruleset-aws"
}

# 警告として残したいルールはここで level を rewrite できる。
# 本 MVP では default を採用。
