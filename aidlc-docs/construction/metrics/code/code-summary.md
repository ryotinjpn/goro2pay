# Unit E (metrics) Code Generation Summary

## Overview

Unit E implements ダメ化メトリクス表示・予算増額機能。
Backend: errgroup 並列読み取り + 消化率計算 + 増額推奨ロジック。
Frontend: MetricsPanel (DEG UX) + RaiseModal + BudgetEmptyScreen。

---

## Backend — `apps/api/`

### 新規ファイル

| ファイル | 説明 |
|---|---|
| `internal/metrics/types.go` | `Metrics` struct, `MetricsService` interface, `ErrNoBudgetSet` |
| `internal/metrics/compute.go` | `computeConsumptionRate` (clamp [0,1]), `computeMetrics`, `formatYen` |
| `internal/metrics/service.go` | P-ME-PARALLEL-01: errgroup 3 フェーズ並列読み取り |
| `internal/metrics/service_test.go` | table tests + gopter PBT P-E-PBT-01 (ConsumptionRate ∈ [0,1]) |
| `internal/budget_raise/types.go` | `BudgetRaiseResult` struct, `BudgetRaiseService` interface |
| `internal/budget_raise/compute.go` | `computeRecommendedBudget` (×1.5, cap 100,000), `computeNextMonthStart` (JST) |
| `internal/budget_raise/service.go` | `ComputeRecommendedBudget` + `Accept` (DynamoDB PUT) |
| `internal/budget_raise/service_test.go` | table tests + gopter PBT P-E-PBT-02 (monotonic bounds) |
| `internal/handlers/metrics_handler.go` | GET /api/metrics — `ErrNoBudgetSet` → 400 ERR_NO_BUDGET_SET |
| `internal/handlers/budget_raise_handler.go` | GET /api/budget/raise/recommendation + POST /api/budget/raise |
| `internal/handlers/metrics_handler_test.go` | handler tests (auth key: `"userId"`) |
| `internal/handlers/budget_raise_handler_test.go` | handler tests (auth key: `"userId"`) |

### 変更ファイル

| ファイル | 変更内容 |
|---|---|
| `go.mod` | `golang.org/x/sync` 追加 (errgroup)、Go 1.25.0 へ更新 |
| `main.go` | `MetricsService`/`BudgetRaiseService` DI 配線 + 3 ルート登録 |

---

## Frontend — `web/`

### 新規ファイル

| ファイル | 説明 |
|---|---|
| `lib/api/metrics.ts` | `fetchMetrics`, `fetchRaiseRecommendation`, `postBudgetRaise` |
| `hooks/useMetrics.ts` | LC-ME-07: `remainingBalance===0` → `/budget-empty` redirect |
| `hooks/useBudgetRaise.ts` | LC-ME-08: recommendation query + raise mutation (invalidates 3 keys) |
| `components/metrics/MetricsPanel.tsx` | LC-ME-09: DEG UX (bg-red-500 + animate-pulse when thresholdExceeded) |
| `components/metrics/RaiseModal.tsx` | LC-ME-11: ダメコピー + 増額 mutation |
| `app/budget-empty/page.tsx` | LC-ME-10: `useState(true)` で RaiseModal 自動オープン (P-ME-FE-DEG-01 §4.2) |
| `tests/useMetrics.test.tsx` | 正常データ取得 + remainingBalance===0 リダイレクト |
| `tests/useBudgetRaise.test.tsx` | recommendation 取得 + mutation 成功 |
| `tests/MetricsPanel.test.tsx` | loading skeleton + 正常表示 + ThresholdExceeded DEG UX |
| `tests/RaiseModal.test.tsx` | isOpen=false/true + reject + accept button enabled |

### 変更ファイル

| ファイル | 変更内容 |
|---|---|
| `hooks/useOrder.ts` | `invalidateQueries(['metrics'])` 追加 (BR-FE03) |
| `app/page.tsx` | `<MetricsPanel />` section 追加 |

---

## Infrastructure — `infra/`

### 変更ファイル

| ファイル | 変更内容 |
|---|---|
| `modules/api_gateway/routes.tf` | Unit E 3 ルート追加 (get_metrics / get_budget_raise_recommendation / post_budget_raise) |
| `modules/api_gateway/tests/api_gateway_basic.tftest.hcl` | `unit_e_routes_present` run ブロック追加 (4 asserts) |

---

## テスト結果

### Backend

```
ok  github.com/ryotinjpn/goro2pay/apps/api/internal/metrics       PASS
ok  github.com/ryotinjpn/goro2pay/apps/api/internal/budget_raise  PASS
ok  github.com/ryotinjpn/goro2pay/apps/api/internal/handlers      PASS
```

### Frontend

```
✓ tests/MetricsPanel.test.tsx    (3 tests)
✓ tests/RaiseModal.test.tsx      (4 tests)
✓ tests/useMetrics.test.tsx      (2 tests)
✓ tests/useBudgetRaise.test.tsx  (2 tests)
Test Files  4 passed (4)
Tests       11 passed (11)
```

---

## NFR パターン適用状況

| パターン | 適用先 | 状態 |
|---|---|---|
| P-ME-PARALLEL-01 (errgroup) | `metrics/service.go` | ✅ |
| P-ME-FE-DEG-01 (DEG UX) | `MetricsPanel.tsx`, `RaiseModal.tsx`, `budget-empty/page.tsx` | ✅ |
| P-E-PBT-01 (ConsumptionRate PBT) | `metrics/service_test.go` | ✅ |
| P-E-PBT-02 (RecommendedBudget PBT) | `budget_raise/service_test.go` | ✅ |
| P-INIT-01/P-DI-01 (DI) | `service.go` × 2, `main.go` | ✅ |
| P-OBS-01 (structured log) | `metrics_handler.go`, `budget_raise_handler.go` | ✅ |
| P-FE-LOAD-01 (loading skeleton) | `MetricsPanel.tsx` | ✅ |

---

## 論理コンポーネント完成状況

| ID | コンポーネント | 状態 |
|---|---|---|
| LC-ME-01 | MetricsService | ✅ |
| LC-ME-02 | MetricsHandler | ✅ |
| LC-ME-03 | BudgetRaiseService | ✅ |
| LC-ME-04 | BudgetRaiseHandler | ✅ |
| LC-ME-05 | DTOs (types.go) | ✅ |
| LC-ME-06 | Pure functions (compute.go) | ✅ |
| LC-ME-07 | useMetrics | ✅ |
| LC-ME-08 | useBudgetRaise | ✅ |
| LC-ME-09 | MetricsPanel | ✅ |
| LC-ME-10 | BudgetEmptyScreen | ✅ |
| LC-ME-11 | RaiseModal | ✅ |
