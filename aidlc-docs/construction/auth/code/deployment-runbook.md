# Auth Unit — Deployment Runbook

**Date**: 2026-05-22
**Stage**: Code Generation Step 19

ハッカソン dev 環境への初回デプロイ手順と運用 runbook。

## 前提

- AWS アカウント (Tokyo region 利用可、IAM 権限フル)
- Terraform 1.10+
- AWS CLI 2.x
- Docker (buildx 利用可、arm64 build 可能)
- GitHub: `ryotinjpn/goro2pay` リポジトリ owner 権限

## 初回デプロイ (1 度だけ)

### 1. S3 tfstate bucket を作成

```bash
cd /Users/ryota_matsushita/work/goro2pay
bash infra/scripts/bootstrap-backend.sh
```

成功すると `gp-tfstate-dev` bucket が `ap-northeast-1` に作成される。
versioning / encryption / public-access-block 設定済み。

### 2. terraform.tfvars を作成

```bash
cd infra/envs/dev
cp terraform.tfvars.example terraform.tfvars
# terraform.tfvars を編集して github_owner を実際の値に
```

### 3. Terraform apply (1 回目)

```bash
cd infra/envs/dev
terraform init
terraform apply
```

この時点で構築されるもの:
- Cognito User Pool / App Client / Pre Sign-up Lambda
- API Gateway HTTP API + JWT Authorizer + Stage
- ECR Repository (gp-dev-api-image、空)
- API Lambda (image_uri = ECR :bootstrap、image がまだないので invocation すると失敗)
- Amplify App + branch (auto deploy 設定済みだが Connection 未承認)
- CodePipeline + CodeBuild (Connection 未承認のため動かない)
- IAM Roles 5 種 / CloudWatch Log Groups 3 種

### 4. CodeStar Connection を承認 (手動、1 度だけ)

1. AWS Console を開く
2. Developer Tools → Settings → Connections
3. `gp-dev-github-conn` を選択
4. **"Update pending connection"** をクリック
5. GitHub にリダイレクトされる → ryotinjpn/goro2pay リポジトリへのアクセスを承認 (GitHub App をインストール)
6. AWS Console に戻り、Connection 状態が `Available` になることを確認

### 5. ECR 初回 image を push

```bash
cd /Users/ryota_matsushita/work/goro2pay
bash infra/scripts/bootstrap-ecr-initial.sh
```

成功すると `gp-dev-api-image:bootstrap` が ECR に push される。
これにより API Lambda が起動可能な状態になる。

### 6. Terraform apply (2 回目)

```bash
cd infra/envs/dev
terraform apply
```

API Lambda が再作成され、image_uri が解決される。lifecycle.ignore_changes により以降の terraform は image_uri 変更を無視する。

### 7. Frontend / API CD 動作確認

```bash
cd /Users/ryota_matsushita/work/goro2pay
git checkout develop
# 何か空コミットを push
git commit --allow-empty -m "Trigger initial deploy"
git push origin develop
```

- Amplify Console で `gp-dev-web` の build / deploy が走ることを確認 (~3-5 分)
- CodePipeline Console で `gp-dev-api-pipeline` が動作 (Source → Build) (~5-8 分)
- CodeBuild ログで docker build → ECR push → lambda update-function-code が成功することを確認

### 8. 動作確認

- Frontend: `terraform output amplify_default_domain` で取得した URL を開く
- Sign-up: `/signup` で登録 → CONFIRMED で User が作成されることを Cognito Console で確認
- Logout: ログイン後にログアウト → `POST /api/auth/logout` 204 → CloudWatch Logs `/aws/lambda/gp-dev-api-fn` で `"action":"logout"` 構造化ログ確認

## 通常運用

### Frontend / API コード変更

`develop` ブランチに push するだけで Amplify (Frontend) と CodePipeline (API) が自動デプロイ。

### Pre Sign-up Lambda コード変更

```bash
cd infra/envs/dev
terraform apply
```

archive_file が再 zip 化、source_code_hash 変化検出 → Lambda update。

### インフラ変更 (Terraform)

```bash
cd infra/envs/dev
terraform plan
terraform apply
```

## トラブルシュート

### CodePipeline が動かない

- CodeStar Connection の状態が `Available` か確認 (Pending なら手順 4 を実施)
- IAM Role に必要権限があるか確認

### Lambda invocation が 5xx を返す

- CloudWatch Logs `/aws/lambda/gp-dev-api-fn` を確認
- ECR image_uri が正しい tag を指しているか (CodeBuild ログを確認)
- ECR :bootstrap が push 済みか (`aws ecr describe-images --repository-name gp-dev-api-image`)

### Sign-up で 500

- `/aws/lambda/gp-dev-presignup-fn` のログを確認 (Pre Sign-up Trigger の例外)
- Cognito User Pool の Lambda config が正しく Pre Sign-up Lambda を指しているか

## 削除

```bash
cd infra/envs/dev
terraform destroy
```

`force_delete = true` (ECR) / `deletion_protection = INACTIVE` (Cognito) のため、すべて削除可能。S3 tfstate bucket は手動削除が必要 (Terraform 管理外)。
