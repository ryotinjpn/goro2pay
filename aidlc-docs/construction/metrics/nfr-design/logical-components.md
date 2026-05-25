# Unit E (`metrics`) — Logical Components

**Document Version**: 1.0
**Created**: 2026-05-25
**Stage**: Construction / NFR Design
**Unit**: E — `metrics`（ダメ化メトリクス）
**Depth**: Standard
**Related**: [nfr-design-patterns.md](./nfr-design-patterns.md)（P-ME-*）、凍結契約 [unit-interfaces.md](../../interfaces/unit-interfaces.md) §6
**方針**: Q-DD5=A。Backend は Service / Handler を分離して計 6 LC、Frontend は 5 LC（計 11 LC）。共有コンポーネント（WalletReader / BudgetSettingsReader / OrderHistoryReader）は Unit B/C 所有として参照のみ。

---

## 1. コンポーネント識別子規則

- 形式: `LC-ME-{番号}`（Logical Component / Metrics / E）
- 共有コンポーネント（Unit B/C 所有）は参照のみ、重複定義しない

---

## 2. Backend コンポーネント（Unit E 所有）

### LC-ME-01: MetricsService

- **責務**: `GetMetrics(ctx, userID) (*Metrics, error)`（凍結契約 §6.1）の実装。P-ME-PARALLEL-01 に従い BudgetSettings + Wallet を errgroup で並列取得 → ErrNoBudgetSet 判定 → CountThisMonth 直列呼出 → ConsumptionRate / ThresholdExceeded / SummaryText 計算
- **依存先**: `WalletReader`（LC-B-xx 参照）/ `BudgetSettingsReader`（LC-B-xx 参照）/ `OrderHistoryReader`（LC-C-xx 参照）
- **パターン**: P-ME-PARALLEL-01 / P-OBS-01 / P-DI-01 / P-INIT-01
- **配置**: `apps/api/internal/metrics/service.go`

### LC-ME-02: MetricsHandler

- **責務**: `GET /api/metrics` の Gin handler。`auth.UserIDFromContext` で userID 取得 → `MetricsService.GetMetrics` 呼出 → エラーマッピング（ErrNoBudgetSet → 400, その他 → 500）→ JSON レスポンス
- **依存先**: `MetricsService`（LC-ME-01）/ `auth.UserIDFromContext`（Unit A）
- **配置**: `apps/api/internal/handlers/metrics_handler.go`

### LC-ME-03: BudgetRaiseService

- **責務**: `ComputeRecommendedBudget(ctx, userID) (int, error)` / `Accept(ctx, userID, newMonthlyBudget) (*BudgetRaiseResult, error)`（凍結契約 §6.1）の実装。推奨額 = `min(floor(current * 1.5), 100_000)`。Accept は JST 翌月 1 日 00:00 を server-side で算出して `BudgetSettingsWriter.Set` を呼ぶ
- **依存先**: `BudgetSettingsReader`（LC-B-xx 参照）/ `BudgetSettingsWriter`（LC-B-xx 参照）
- **パターン**: P-DI-01 / P-INIT-01
- **配置**: `apps/api/internal/budget_raise/service.go`

### LC-ME-04: BudgetRaiseHandler

- **責務**: `GET /api/budget/raise/recommendation` および `POST /api/budget/raise` の Gin handler。userID 取得 → BudgetRaiseService 呼出 → エラーマッピング（ErrInvalidBudget / ErrNoBudgetSet → 400）→ JSON レスポンス
- **依存先**: `BudgetRaiseService`（LC-ME-03）/ `auth.UserIDFromContext`（Unit A）
- **配置**: `apps/api/internal/handlers/budget_raise_handler.go`

### LC-ME-05: Metrics / BudgetRaiseResult（DTO）

- **責務**: 凍結契約 §6.1 の公開 DTO。`Metrics{DamageCount, ConsumptionRate, MonthlyBudget, RemainingBalance, ThresholdExceeded, SummaryText}` / `BudgetRaiseResult{NewMonthlyBudget, AppliedFrom}` / `ErrNoBudgetSet` / `ErrInvalidBudget`
- **配置**: `apps/api/internal/metrics/types.go` / `apps/api/internal/budget_raise/types.go`

### LC-ME-06: computeMetrics / computeRecommendedBudget（純関数）

- **責務**: テスト・PBT のために Service ロジックから切り出した純関数群。
  - `computeConsumptionRate(monthlyBudget, remainingBalance int) float64`
  - `computeMetrics(bs *BudgetSettings, w *Wallet, count int) *Metrics`
  - `computeRecommendedBudget(currentBudget int) int`
  - `computeNextMonthStart() time.Time`（JST 翌月 1 日 00:00）
- **パターン**: P-E-PBT-01 / P-E-PBT-02 の PBT 対象
- **配置**: `apps/api/internal/metrics/compute.go` / `apps/api/internal/budget_raise/compute.go`

---

## 3. Frontend コンポーネント（Unit E 所有）

### LC-ME-07: useMetrics（TanStack Query hook）

- **責務**: `GET /api/metrics` を TanStack Query でフェッチ。`queryKey: ['metrics']`、`staleTime: 0`（常に最新）、`retry: 1`。
  - `isError` かつ `ERR_NO_BUDGET_SET` → `/budget` へリダイレクト
  - `data.remainingBalance === 0` → `/budget-empty` へリダイレクト（BR-FE01）
- **パターン**: P-FE-LOAD-01 / P-ME-FE-DEG-01（BudgetEmpty リダイレクト）
- **配置**: `web/src/hooks/useMetrics.ts`

### LC-ME-08: useBudgetRaise（TanStack Query hook）

- **責務**: `GET /api/budget/raise/recommendation` のフェッチ（queryKey: `['budget-raise-recommendation']`）+ `POST /api/budget/raise` の mutation。mutation 成功後に `['metrics']` / `['wallet']` / `['budget-raise-recommendation']` を invalidate
- **パターン**: P-FE-LOAD-01
- **配置**: `web/src/hooks/useBudgetRaise.ts`

### LC-ME-09: MetricsPanel

- **責務**: MainScreen に埋め込まれる Metrics 表示コンポーネント。ダメ化回数 / 消化率プログレスバー / 残高 / SummaryText を表示。`ThresholdExceeded` に応じて P-ME-FE-DEG-01 §4.1 のスタイルを切替。`useMetrics` で最新データを取得
- **依存**: `useMetrics`（LC-ME-07）
- **パターン**: P-ME-FE-DEG-01（animate-pulse）/ P-FE-LOAD-01（isLoading skeleton）
- **配置**: `web/src/components/MetricsPanel.tsx`

### LC-ME-10: BudgetEmptyScreen

- **責務**: `/budget-empty` ページ。マウント時に `useState(true)` で RaiseModal を自動表示（P-ME-FE-DEG-01 §4.2）。ダメコピー「今月はもうダメになれません」を表示。`RaiseModal` を `isOpen` / `onClose` で制御
- **依存**: `useBudgetRaise`（LC-ME-08）/ `RaiseModal`（LC-ME-11）
- **パターン**: P-ME-FE-DEG-01（RaiseModal 強制表示）
- **配置**: `web/src/app/budget-empty/page.tsx`

### LC-ME-11: RaiseModal

- **責務**: 増額誘導モーダル。`GET /api/budget/raise/recommendation` で推奨額を表示し、`POST /api/budget/raise` で増額を適用。ダメコピー一覧（P-ME-FE-DEG-01 §4.3）を使用。「今月はがんばる」で閉じる（`onClose` call）
- **依存**: `useBudgetRaise`（LC-ME-08）
- **パターン**: P-ME-FE-DEG-01（ダメコピー）/ P-FE-LOAD-01（isLoading 中はボタン無効化）
- **配置**: `web/src/components/RaiseModal.tsx`

---

## 4. 共有コンポーネント（他 Unit 所有、参照のみ）

| 参照 LC | 名称 | Unit E での用途 |
|---|---|---|
| LC-B-xx（Unit B 所有） | `WalletReader` | `MetricsService` の残高読取 |
| LC-B-xx（Unit B 所有） | `BudgetSettingsReader` | `MetricsService` / `BudgetRaiseService` の予算設定読取 |
| LC-B-xx（Unit B 所有） | `BudgetSettingsWriter` | `BudgetRaiseService.Accept` の予算更新 |
| LC-C-xx（Unit C 所有） | `OrderHistoryReader` | `MetricsService` の CountThisMonth 呼出 |
| LC-AUTH-xx（Unit A 所有） | `AttachUserID` middleware | 全エンドポイントの認証（NFRE-E10） |

---

## 5. コンポーネント依存グラフ

```
[Unit A] AttachUserID
    │
    ▼
LC-ME-02: MetricsHandler ──→ LC-ME-01: MetricsService
LC-ME-04: BudgetRaiseHandler ──→ LC-ME-03: BudgetRaiseService

LC-ME-01 ──errgroup──→ [Unit B] WalletReader
         ──errgroup──→ [Unit B] BudgetSettingsReader
         ──直列──────→ [Unit C] OrderHistoryReader.CountThisMonth
         ──純関数────→ LC-ME-06: computeMetrics / computeConsumptionRate

LC-ME-03 ──────────→ [Unit B] BudgetSettingsReader
         ──────────→ [Unit B] BudgetSettingsWriter
         ──純関数──→ LC-ME-06: computeRecommendedBudget / computeNextMonthStart

[Frontend]
LC-ME-09: MetricsPanel ──→ LC-ME-07: useMetrics
LC-ME-10: BudgetEmptyScreen ──→ LC-ME-08: useBudgetRaise
                             ──→ LC-ME-11: RaiseModal
LC-ME-11: RaiseModal ──→ LC-ME-08: useBudgetRaise

LC-ME-07: useMetrics (queryKey: ['metrics'])
          ← invalidate from: useOrder mutation success (Unit C)
          ← invalidate from: useBudgetRaise mutation success (LC-ME-08)
```

---

## 6. 文書管理

- **凍結契約への影響**: なし（LC-ME-* は Unit E 内部）
- **次ステージ**: Infrastructure Design（Standard）
