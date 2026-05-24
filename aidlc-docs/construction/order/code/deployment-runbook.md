# Unit C — Deployment Runbook

**Generated**: 2026-05-24
**Stage**: Construction / Code Generation
**Unit**: C (`order` / 代行手配コア)

本書は Unit C のリソースを dev 環境にデプロイするための手順書。Unit A の
`apps/api/` / `web/` / `infra/` 既存構成への追加分のみを扱う。

---

## 1. 前提

| 項目 | 値 |
|---|---|
| Region | `ap-northeast-1` |
| Terraform Backend | S3 (`gp-tfstate-dev`) + native lock (`use_lockfile = true`) |
| AWS Account | (Unit A デプロイ時に bootstrap 済み) |
| Bedrock Model Access | Claude Haiku 4.5 が ap-northeast-1 で承認済み (要事前申請) |
| Unit A | デプロイ済み (PR #69-#73 マージ後の develop 環境) |

---

## 2. 初回 apply 手順

### 2.1 Bedrock モデルアクセス申請 (初回のみ)

AWS Console → Bedrock → Model access → "Anthropic Claude Haiku 4.5" を有効化。
APAC Inference Profile (`jp.anthropic.claude-haiku-4-5-20251001-v1:0`) を使うため、
ap-northeast-1 に加えて他の APAC リージョンの承認も必要な場合あり。

### 2.2 alarm_email の設定

`infra/envs/dev/locals.tf` の `local.alarm_email` を実メールアドレスに変更する
(デフォルトは `alerts@example.com` placeholder)。

### 2.3 terraform apply

```sh
cd infra/envs/dev
terraform init                    # 新規 module 3 種を初期化
terraform plan                    # 変更内容確認 (新規リソース追加のみ)
terraform apply                   # 確認後実行
```

### 2.4 SNS 購読確認メール承認

apply 完了後 5 分以内に AWS から `alarm_email` 宛に確認メールが届く。
本文中の `Confirm subscription` リンクをクリックして購読を有効化する。
未承認だと 80% 警告 / 100% 警報が届かない。

### 2.5 動作確認

```sh
terraform output order_history_table_name  # → gp-dev-order-history
terraform output sns_topic_arn             # → arn:aws:sns:ap-northeast-1:xxx:gp-dev-alarms
```

---

## 3. Backend (apps/api/) コードのデプロイ

Unit A が構築した CodePipeline がそのまま Unit C のコードもデプロイする。
追加作業なし。

```sh
git push origin develop
# → CodeStar Connection 経由で CodePipeline 起動
# → CodeBuild が docker build → ECR push → lambda update-function-code
# → 数分で API Lambda が新 image で稼働
```

CodePipeline の状態は AWS Console / `gh actions` 等で確認可能。

---

## 4. Frontend (web/) コードのデプロイ

Unit A が構築した Amplify Hosting がそのまま Unit C のコードもデプロイする。
追加作業なし。

```sh
git push origin develop
# → Amplify が GitHub から build → SSR 配信
# → 数分で MainScreen に GoroButton が表示される
```

---

## 5. ハッカソン実演前チェックリスト

| チェック項目 | 確認方法 |
|---|---|
| API Lambda が Bedrock を呼べる | CloudWatch Logs Insights で `event = "place_order_complete"` を検索、`bedrockLatencyMs` が記録されているか |
| OrderHistory に Insert 成功 | DynamoDB Console で `gp-dev-order-history` テーブルにレコード存在 |
| 3 アラームが INSUFFICIENT_DATA / OK | CloudWatch Alarms Console で全 alarm が ALARM 以外 |
| SNS 購読が Confirmed | SNS Console の Subscription で `Status = Confirmed` |
| Bedrock コストアラートが Active | Budgets Console で `gp-dev-bedrock-budget` が `Active` |
| Frontend GoroButton が動作 | https://{amplify_domain}/ → ログイン → GoroButton 押下 → OrderCompletionScreen 表示 |

**注意**: Unit B WalletService が未実装の場合、`POST /api/orders` は常に 402 を返す
(`apps/api/wallet_stub.go` の `noopWalletService`)。Frontend は BudgetEmptyScreen
へ遷移する動作確認のみ可能。Unit B 完了後 main.go の DI 配線で `walletStub` を
本実装に差し替え + 再デプロイで通常フロー (200 → OrderCompletionScreen) に切替。

---

## 6. ロールバック

### 6.1 アプリケーションコード (Backend / Frontend)

- Backend: ECR の旧 image tag に lambda update-function-code
- Frontend: Amplify Console で旧 build にロールバック

### 6.2 Terraform リソース

```sh
git revert <commit>
cd infra/envs/dev
terraform apply
```

DynamoDB OrderHistory の destroy は data 損失のため慎重に (本 MVP は破棄
許容、本番化時は `prevent_destroy = true` を検討)。

---

## 7. 後続ステージ (Build & Test) への引き継ぎ

- E2E テスト (Playwright): 「ログイン → ご飯めんどくさい押下 → OrderCompletionScreen → MainScreen 自動遷移」フロー検証
- CI ワークフロー: `go test ./...` / `npm test` / `terraform test` の自動化
- 負荷テスト: 数十 req/s で NFRC-C01 p95 3.0s 達成確認
- Bedrock 実呼出し動作確認: dev 環境のみ実 SDK で疎通テスト
