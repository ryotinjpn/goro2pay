# Unit E (`metrics`) — Deployment Architecture

**Document Version**: 1.0
**Created**: 2026-05-25
**Stage**: Construction / Infrastructure Design
**Unit**: E — `metrics`（ダメ化メトリクス）

---

## 1. デプロイメント構成図

```
[Browser / PWA]
      |
      | HTTPS
      v
[CloudFront + Amplify Hosting]
  web/src/app/page.tsx (MainScreen + MetricsPanel)
  web/src/app/budget-empty/page.tsx (BudgetEmptyScreen + RaiseModal)
      |
      | fetch /api/metrics
      | fetch /api/budget/raise/recommendation
      | POST  /api/budget/raise
      v
[API Gateway HTTP API (ap-northeast-1)]
  GET  /api/metrics                    → JWT Authorizer → Lambda
  GET  /api/budget/raise/recommendation → JWT Authorizer → Lambda
  POST /api/budget/raise               → JWT Authorizer → Lambda
      |
      | (Cognito JWT 検証後)
      v
[API Lambda (arm64, 256MB)]
  apps/api/internal/metrics/service.go   (MetricsService)
  apps/api/internal/metrics/handler.go   (MetricsHandler)
  apps/api/internal/budget_raise/service.go (BudgetRaiseService)
  apps/api/internal/handlers/budget_raise_handler.go
      |
      |-- errgroup (並列) ──→ [DynamoDB] GoroPay_BudgetSettings
      |-- errgroup (並列) ──→ [DynamoDB] GoroPay_Wallet
      |-- 直列 ────────────→ [DynamoDB] GoroPay_OrderHistory (CountThisMonth)
      |
      |-- BudgetRaiseService.Accept ──→ [DynamoDB] GoroPay_BudgetSettings (PutItem/UpdateItem)
```

---

## 2. AWS サービスマッピング

| コンポーネント | AWS サービス | 設定 |
|---|---|---|
| Frontend | Amplify Hosting | 既存（変更なし） |
| API ルーティング | API Gateway HTTP API | 既存 + 3 route 追記 |
| 認証 | Cognito JWT Authorizer | 既存（変更なし） |
| ビジネスロジック | Lambda (arm64, 256MB) | 既存 + コード追加 |
| メトリクス読取 | DynamoDB (GoroPay_Wallet, GoroPay_BudgetSettings, GoroPay_OrderHistory) | 既存テーブル（読取） |
| 予算更新 | DynamoDB (GoroPay_BudgetSettings) | 既存テーブル（書込） |
| ログ | CloudWatch Logs | 既存 log group（変更なし） |

---

## 3. データフロー（GET /api/metrics）

```
1. Browser → GET /api/metrics (Authorization: Bearer <JWT>)
2. API Gateway → Cognito JWT Authorizer 検証
3. API Gateway → Lambda invoke (event.requestContext.authorizer.jwt.claims.sub = userID)
4. MetricsHandler → auth.UserIDFromContext(c)
5. MetricsService.GetMetrics(ctx, userID):
   a. errgroup.Go: BudgetSettingsReader.Get → GoroPay_BudgetSettings GetItem
   b. errgroup.Go: WalletReader.Get        → GoroPay_Wallet GetItem
   c. (並列完了後) ErrNoBudgetSet チェック
   d. OrderHistoryReader.CountThisMonth    → GoroPay_OrderHistory Query
   e. computeMetrics(bs, wallet, count) → Metrics
6. MetricsHandler → 200 JSON レスポンス
```

---

## 4. データフロー（POST /api/budget/raise）

```
1. Browser → POST /api/budget/raise {"newMonthlyBudget": 45000}
2. API Gateway → Cognito JWT 検証
3. BudgetRaiseHandler → auth.UserIDFromContext(c)
4. BudgetRaiseService.Accept(ctx, userID, 45000):
   a. バリデーション (1 <= 45000 <= 100_000)
   b. computeNextMonthStart() → 2026-06-01T00:00:00+09:00 (JST)
   c. BudgetSettingsWriter.Set(ctx, userID, 45000, effectiveFrom)
      → GoroPay_BudgetSettings PutItem/UpdateItem
5. BudgetRaiseHandler → 200 {"newMonthlyBudget":45000,"appliedFrom":"2026-06-01T00:00:00+09:00"}
```

---

## 5. Terraform 変更サマリ

| module | 変更 | 理由 |
|---|---|---|
| `modules/api_gateway` | `routes.tf` に 3 route 追記 + `tests/` に 4 assert 追記 | Unit E エンドポイント追加 |
| `modules/lambda_api` | 変更なし | env/IAM は Unit B/C 実装済みでカバー |
| `modules/budget` | 変更なし | BudgetSettingsReadWrite 権限は既存 policy 内 |
| `modules/order_history` | 変更なし | Query 権限は Unit C 実装済みでカバー |
| `modules/amplify` | 変更なし | Frontend は既存 Hosting で配信 |

---

## 6. 非機能要件マッピング

| NFRE | インフラ対応 |
|---|---|
| NFRE-E01 P95 500ms | Lambda arm64 + errgroup 並列読取（コード実装） |
| NFRE-E04 DynamoDB 障害 → 500 | Lambda error handling（コード実装） |
| NFRE-E06 新規テーブルなし | 本設計の通り（Terraform 変更最小） |
| NFRE-E10 JWT 認証 | API Gateway JWT Authorizer（routes.tf 追記） |
