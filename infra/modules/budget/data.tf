# Scheduler Lambda の bootstrap バイナリを zip 化 (Q-I2=A)。
# 前提: terraform apply 前に `make -C apps/api/cmd/scheduler build` を実行し
# `apps/api/cmd/scheduler/bootstrap` を生成しておく必要がある (Q-I3=A、CI/CD なし)。
#
# var.scheduler_bootstrap_path が空文字列のときはリポジトリ構造を仮定したデフォルト
# パスを使う (path.module から 3 階層上)。CI/CD で workspace 構造が異なる場合は
# 呼び出し側で絶対パスを渡せる (Code Review Minor 12)。
data "archive_file" "scheduler" {
  type        = "zip"
  source_file = local.bootstrap_path
  output_path = "${path.module}/.terraform/tmp/scheduler.zip"
}
