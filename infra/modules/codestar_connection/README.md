# codestar_connection module

GitHub への CodeStar Connection を 1 つだけ作成する共有モジュール。

## 用途

- CodePipeline (lambda_api module) の Source Provider
- Amplify Hosting (amplify module) の repo 接続情報

両モジュールが同じ ARN を参照するため、env (envs/dev/, envs/prd/) 配下で
1 度だけインスタンス化し、`module.codestar_connection.connection_arn` を
両モジュールに変数で渡す。

## 初回セットアップ

`terraform apply` 直後の Connection は **Pending** 状態で、AWS Console
で手動承認 (Update pending connection → GitHub App をインストール) する
必要がある。詳細は `aidlc-docs/construction/auth/infrastructure-design/`
配下の deployment-runbook.md §4 を参照。
