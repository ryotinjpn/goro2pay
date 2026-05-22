# `infra/modules/auth/` — Auth Unit Terraform module

Cognito User Pool / Pre Sign-up Lambda / API Gateway HTTP API / API Lambda /
Amplify Hosting / CodePipeline + CodeBuild / ECR / IAM Roles を 1 module に
まとめた、Auth Unit + 横串インフラの Terraform module。

## 利用例 (`infra/envs/dev/main.tf`)

```hcl
module "auth" {
  source = "../../modules/auth"

  env              = "dev"
  region           = "ap-northeast-1"
  github_owner     = "ryotinjpn"
  github_repo_name = "goro2pay"
  github_branch    = "develop"
}
```

## デプロイ手順

1. **S3 tfstate bucket を作成** (1 度だけ):
   ```bash
   bash infra/scripts/bootstrap-backend.sh
   ```
2. **Terraform apply (1 回目)**:
   ```bash
   cd infra/envs/dev
   terraform init
   terraform apply
   ```
   この時点で:
   - Cognito / Pre Sign-up Lambda / API Gateway / Authorizer / IAM / ECR / Amplify / CodePipeline は作成される
   - **API Lambda は ECR :bootstrap タグの初期 image を参照しようとする** → image がまだない
3. **CodeStar Connection を承認** (手動、1 度だけ):
   - AWS Console → Developer Tools → Settings → Connections
   - `gp-dev-github-conn` を選択 → "Update pending connection" → GitHub App をインストール
   - 状態が `Available` になることを確認
4. **ECR 初回 image を push** (1 度だけ):
   ```bash
   bash infra/scripts/bootstrap-ecr-initial.sh
   ```
5. **Terraform apply (2 回目)** で API Lambda を再作成:
   ```bash
   cd infra/envs/dev
   terraform apply
   ```
6. **Frontend は GitHub develop branch への push で auto deploy** (Amplify Hosting)
7. **API Lambda は GitHub develop branch への push で auto CD** (CodePipeline → CodeBuild)

## 削除

```bash
cd infra/envs/dev
terraform destroy
```

`deletion_protection = INACTIVE` (Q-I13) のため User Pool も削除可能。
ECR Repository は `force_delete = true` (dev のみ) で image 残っていても削除可能。

## ファイル構成

| ファイル | リソース |
|---|---|
| `main.tf` | データソース・共通定義 |
| `cognito.tf` | aws_cognito_user_pool / app_client / lambda_permission |
| `pre_signup_lambda.tf` | aws_lambda_function (Node.js auto-confirm) + archive_file |
| `api_gateway.tf` | aws_apigatewayv2_api / authorizer / stage |
| `api_lambda.tf` | aws_lambda_function (Go + Gin + LWA) |
| `ecr.tf` | aws_ecr_repository + lifecycle_policy |
| `logout_route.tf` | aws_apigatewayv2_route + integration |
| `amplify.tf` | aws_amplify_app / branch / aws_codestarconnections_connection |
| `codepipeline.tf` | aws_codepipeline / aws_codebuild_project / S3 artifacts |
| `iam.tf` | IAM Role × 5 (Pre Sign-up / API Lambda / Amplify SSR / CodePipeline / CodeBuild) |
| `log_groups.tf` | CloudWatch Log Group × 3 |
| `variables.tf` | 入力変数 |
| `outputs.tf` | 13 種 |
| `tests/*.tftest.hcl` | mock_provider テスト 6 種 |
