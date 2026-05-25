# Unit E `metrics` — Code Generation Plan

**Document Version**: 1.0
**Created**: 2026-05-25
**Unit**: E (`metrics` / ダメ化メトリクス)
**Construction Depth**: Standard

---

## 対象ストーリー

| ストーリー | 実装内容 |
|---|---|
| US-3-01 | MetricsPanel — ダメ化回数表示 |
| US-3-02 | MetricsPanel — 消化率 + 警告色 (ThresholdExceeded) |
| US-3-03 | BudgetEmptyScreen — remainingBalance===0 でリダイレクト |
| US-3-04 | RaiseModal — BudgetEmptyScreen マウント時自動表示 |
| US-X-03 | Accept フロー — 翌月増額 + 退化ループ完成 |

---

## 依存関係

- **Unit B**: `budget_settings.BudgetSettingsReader` / `BudgetSettingsWriter` / `wallet_repo.WalletReader` (全て実装済み)
- **Unit C**: `order_history.OrderHistoryReader` (CountThisMonth 等、本セッションで実装済み)
- **Unit A**: `auth.AttachUserID` / `auth.UserIDFromContext` (実装済み)
- **新規 Go 依存**: `golang.org/x/sync` (goroutine グループ管理、未追加 → Step 1 で追加)

---

## ステップ一覧

- [x] Step 1: go.mod に `golang.org/x/sync` を追加
- [x] Step 2: `apps/api/internal/metrics/types.go` — Metrics DTO + ErrNoBudgetSet
- [x] Step 3: `apps/api/internal/metrics/compute.go` — computeConsumptionRate / computeMetrics 純関数
- [x] Step 4: `apps/api/internal/metrics/service.go` — MetricsService (P-ME-PARALLEL-01)
- [x] Step 5: `apps/api/internal/metrics/service_test.go` — テーブルテスト + PBT (P-E-PBT-01)
- [x] Step 6: `apps/api/internal/budget_raise/types.go` — BudgetRaiseResult DTO + ErrInvalidBudget
- [x] Step 7: `apps/api/internal/budget_raise/compute.go` — computeRecommendedBudget / computeNextMonthStart 純関数
- [x] Step 8: `apps/api/internal/budget_raise/service.go` — BudgetRaiseService
- [x] Step 9: `apps/api/internal/budget_raise/service_test.go` — テーブルテスト + PBT (P-E-PBT-02)
- [x] Step 10: `apps/api/internal/handlers/metrics_handler.go` — MetricsHandler (GET /api/metrics)
- [x] Step 11: `apps/api/internal/handlers/budget_raise_handler.go` — BudgetRaiseHandler (GET+POST /api/budget/raise/...)
- [x] Step 12: `apps/api/internal/handlers/metrics_handler_test.go` + `budget_raise_handler_test.go`
- [x] Step 13: `apps/api/main.go` — DI 配線 + 3 route 追加
- [x] Step 14: `web/lib/api/metrics.ts` — API クライアント関数
- [x] Step 15: `web/hooks/useMetrics.ts` — TanStack Query hook (LC-ME-07)
- [x] Step 16: `web/hooks/useBudgetRaise.ts` — TanStack Query hook + mutation (LC-ME-08)
- [x] Step 17: `web/hooks/useOrder.ts` — 修正: invalidateQueries(['metrics']) 追加 (BR-FE03)
- [x] Step 18: `web/components/metrics/MetricsPanel.tsx` — DEG UX (animate-pulse / P-ME-FE-DEG-01)
- [x] Step 19: `web/components/metrics/RaiseModal.tsx` — 増額モーダル + ダメコピー
- [x] Step 20: `web/app/(authenticated)/budget-empty/page.tsx` — BudgetEmptyScreen (useState(true) で自動表示)
- [x] Step 21: `web/app/page.tsx` — 修正: MetricsPanel 組み込み + remainingBalance===0 リダイレクト
- [x] Step 22: `web/tests/useMetrics.test.tsx` + `useBudgetRaise.test.tsx`
- [x] Step 23: `web/tests/MetricsPanel.test.tsx` + `RaiseModal.test.tsx`
- [x] Step 24: `infra/modules/api_gateway/routes.tf` — 3 route 追記
- [x] Step 25: `infra/modules/api_gateway/tests/api_gateway_basic.tf テスト.hcl` — 4 assert 追記
- [x] Step 26: `aidlc-docs/construction/metrics/code/code-summary.md`

---

## ファイルパス一覧

### Backend (Go)
| ファイル | 新規/修正 |
|---|---|
| `apps/api/go.mod` | 修正（golang.org/x/sync 追加）|
| `apps/api/internal/metrics/types.go` | 新規 |
| `apps/api/internal/metrics/compute.go` | 新規 |
| `apps/api/internal/metrics/service.go` | 新規 |
| `apps/api/internal/metrics/service_test.go` | 新規 |
| `apps/api/internal/budget_raise/types.go` | 新規 |
| `apps/api/internal/budget_raise/compute.go` | 新規 |
| `apps/api/internal/budget_raise/service.go` | 新規 |
| `apps/api/internal/budget_raise/service_test.go` | 新規 |
| `apps/api/internal/handlers/metrics_handler.go` | 新規 |
| `apps/api/internal/handlers/budget_raise_handler.go` | 新規 |
| `apps/api/internal/handlers/metrics_handler_test.go` | 新規 |
| `apps/api/internal/handlers/budget_raise_handler_test.go` | 新規 |
| `apps/api/main.go` | 修正（DI + routes）|

### Frontend (TypeScript/Next.js)
| ファイル | 新規/修正 |
|---|---|
| `web/lib/api/metrics.ts` | 新規 |
| `web/hooks/useMetrics.ts` | 新規 |
| `web/hooks/useBudgetRaise.ts` | 新規 |
| `web/hooks/useOrder.ts` | 修正（invalidateQueries 追加）|
| `web/components/metrics/MetricsPanel.tsx` | 新規 |
| `web/components/metrics/RaiseModal.tsx` | 新規 |
| `web/app/(authenticated)/budget-empty/page.tsx` | 新規 |
| `web/app/page.tsx` | 修正（MetricsPanel + リダイレクト）|
| `web/tests/useMetrics.test.tsx` | 新規 |
| `web/tests/useBudgetRaise.test.tsx` | 新規 |
| `web/tests/MetricsPanel.test.tsx` | 新規 |
| `web/tests/RaiseModal.test.tsx` | 新規 |

### Infrastructure (Terraform)
| ファイル | 新規/修正 |
|---|---|
| `infra/modules/api_gateway/routes.tf` | 修正（3 route 追記）|
| `infra/modules/api_gateway/tests/api_gateway_basic.tf テスト.hcl` | 修正（4 assert 追記）|

### Documentation
| ファイル | 新規/修正 |
|---|---|
| `docs/construction/metrics/code/code-summary.md` | 新規 |

---

## 検証計画

| 検証 | コマンド |
|---|---|
| Backend Go test | `cd apps/api && go test ./internal/metrics/... ./internal/budget_raise/... ./internal/handlers/... -v` |
| Frontend Vitest | `cd web && npx vitest run tests/useMetrics.test.tsx tests/useBudgetRaise.test.tsx tests/MetricsPanel.test.tsx tests/RaiseModal.test.tsx` |
| TypeScript 型チェック | `cd web && npx tsc --noEmit` |
| Terraform validate | `cd infra/modules/api_gateway && terraform validate` |
| Terraform test | `cd infra/modules/api_gateway && terraform test` |
