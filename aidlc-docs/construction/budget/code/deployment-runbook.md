# Deployment Runbook — Unit B (`budget`)

## 前提条件

- Unit A (auth) + Unit C (order) が既に dev 環境にデプロイ済み
- AWS Profile: `dev-kyoto-sso-administrator`
- Region: `ap-northeast-1`
- `gp-tfstate-dev` S3 backend にアクセス権あり

## デプロイ手順

### 1. Scheduler Lambda の bootstrap バイナリ生成 (必須)

terraform apply 前に **必ず** 実行する:

```bash
cd /Users/ryota_matsushita/work/goro2pay
make -C apps/api/cmd/scheduler build
ls -lh apps/api/cmd/scheduler/bootstrap   # 14MB 程度
```

`apps/api/cmd/scheduler/bootstrap` (Linux arm64) が生成される。これが無いと terraform apply は失敗する (Q-I3=A: CI/CD なしの zip 方式)。

### 2. Terraform apply (envs/dev)

```bash
cd /Users/ryota_matsushita/work/goro2pay/infra/envs/dev
terraform init
terraform plan
terraform apply
```

期待される変更:
- DynamoDB テーブル 4 本作成 (`gp-dev-wallet`, `gp-dev-budget-settings`, `gp-dev-idempotency-keys`, `gp-dev-budget-reset-log`)
- Scheduler Lambda + EventBridge Scheduler 作成
- IAM Role / Policy 3 種作成
- API Gateway routes 2 本追加
- API Lambda environment 更新 (`DDB_TABLE_*` 4 個追加)、IAM Role に Unit B Policy attach

### 3. API Lambda Container Image の再デプロイ

API Lambda は既存 CodePipeline で `docker build` → ECR push → Lambda update の流れ。`develop` ブランチに push すると自動的に再ビルド・デプロイされる。

ローカル動作確認のみであれば、ECR への push は CodePipeline に任せて問題ない (環境変数は terraform apply で既に Lambda に注入済み)。

### 4. Frontend (Amplify Hosting)

`web/` 配下の変更は `develop` ブランチ push で Amplify Hosting が自動ビルド + デプロイ。

## 動作確認手順

### Backend API 直叩き (curl)

予算設定:
```bash
curl -X POST "${API_BASE}/api/wallet/budget" \
  -H "Authorization: Bearer ${ID_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"monthlyBudget": 30000}'
# → 200 {"monthlyBudget":30000,"appliedFrom":"..."}
```

残高取得:
```bash
curl "${API_BASE}/api/wallet" \
  -H "Authorization: Bearer ${ID_TOKEN}"
# → 200 {"balance":30000,"monthlyBudget":30000,"updatedAt":"..."}
```

注文 (Wallet 引き落とし発火):
```bash
curl -X POST "${API_BASE}/api/orders" \
  -H "Authorization: Bearer ${ID_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"category":"food","idempotencyKey":"USER_SUB:01HW3XKQ7EZ8XN1MX1C7Q4Z9TY"}'
# → 200 残高減算 → GET /api/wallet で確認
```

範囲外予算:
```bash
curl -X POST "${API_BASE}/api/wallet/budget" -d '{"monthlyBudget": 999999}' ...
# → 400 VALIDATION_FAILED
```

### Frontend ブラウザ動作確認

1. Amplify Hosting URL でログイン
2. `/budget` にアクセス → BudgetSetupScreen が表示される
3. クイックボタンを 1 つ選択 → 「ダメ予算を設定する」クリック → `/` にリダイレクト
4. MainScreen 上部に `BalanceDisplay` が表示される (¥XX,XXX, リセット日カウントダウン)
5. 「ご飯めんどくさい」ボタンを押す → 残高が減算される
6. 残高 < 注文額になるまで連打 → 402 受信 → InsufficientBalanceModal が出る
7. 「もっとダメになる」ボタン → `/budget` に遷移して再設定可能

### Scheduler Lambda 手動 invoke (動作検証)

```bash
aws lambda invoke \
  --function-name gp-dev-scheduler-fn \
  --profile dev-kyoto-sso-administrator \
  --region ap-northeast-1 \
  --payload '{}' \
  /tmp/scheduler-result.json

aws logs tail /aws/lambda/gp-dev-scheduler-fn --since 5m \
  --profile dev-kyoto-sso-administrator
# → "monthly_reset summary processedUsers=N failedUsers=0"
```

## ロールバック手順

```bash
cd /Users/ryota_matsushita/work/goro2pay/infra/envs/dev
git revert HEAD  # Unit B のコミットを revert
terraform apply  # → module.budget の destroy + 既存 module 改の revert
```

DynamoDB テーブルが削除されるとデータ消失するため、本番環境では PITR / バックアップ運用が必要 (Q-I6=A により dev では PITR off)。

## 既知の注意事項

- **bootstrap バイナリの再ビルド**: `apps/api/cmd/scheduler/main.go` 変更時は `make build` を再実行しないと Lambda は古いコードで動き続ける
- **Unit C の env 名と Unit B の env 名の差異**: Unit C は `ORDER_HISTORY_TABLE_NAME`、Unit B は `DDB_TABLE_*` を使う (凍結 IF §10 通り Unit B 採用)。両方が API Lambda に注入される
- **InsufficientBalanceModal と /budget-empty の併存**: Unit C 既存実装の `mapOrderError` 経由 `/budget-empty` ナビゲートはそのまま保持し、追加で modal も出る。今後 Unit E (metrics) 実装時に統合可能性あり

## 関連ドキュメント

- [Plan](../../plans/budget-code-generation-plan.md)
- [Backend Summary](./backend-summary.md)
- [Frontend Summary](./frontend-summary.md)
- [Infrastructure Summary](./infrastructure-summary.md)
- [Infrastructure Design](../infrastructure-design/infrastructure-design.md)
- [凍結 IF (unit-interfaces.md)](../../interfaces/unit-interfaces.md)
