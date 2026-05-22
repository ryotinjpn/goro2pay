# Budget Unit — Logical Components

**Document Version**: 1.0
**Created**: 2026-05-22
**Unit**: B (`budget` / ダメ予算)
**Construction Depth**: Standard
**Stage**: NFR Design / Construction
**Predecessors**: NFR Requirements / [nfr-design-patterns.md](./nfr-design-patterns.md)

本ドキュメントは Unit B の **論理コンポーネント** を定義する。論理レベルでの責務・入出力・他コンポーネントとの関係を明示し、実装本体は Code Generation で書く。

参照: [nfr-design-patterns.md](./nfr-design-patterns.md), [unit-interfaces.md](../../interfaces/unit-interfaces.md), [Functional Design](../functional-design/)

> **Unit A 共有コンポーネントの再利用**:
> - **LC-AUTH-05 `ContextAwareSlogHandler`**: Unit B の Lambda（API + Scheduler）でそのまま再利用する。`amount` / `newBalance` は handler 内で明示 attrs として渡す（P-OBS-01）。Unit B 独自の slog Handler は作成しない。
> - **LC-AUTH-09 `apiClient`**: Unit B の全 Frontend コンポーネントが `apiClient.request(...)` 経由で `/api/wallet/*` を呼ぶ。Unit B 独自の HTTP クライアントは作成しない。
> - **LC-AUTH-18 `BffProxyRouteHandler`**: `/api/wallet/*` リクエストは catch-all proxy（`web/app/api/[...path]/route.ts`）が API Gateway に転送する。Unit B 専用の Route Handler ファイルは不要。

---

## 1. コンポーネント一覧

| ID | 名前 | 配置 | 主要責務 |
|---|---|---|---|
| **LC-BUDGET-01** | `WalletHandler` | Backend / Go (`apps/api/internal/wallet/`) | HTTP リクエスト受付・レスポンス返却（GetBalance / SetBudget / Deduct） |
| **LC-BUDGET-02** | `WalletService` | Backend / Go (`apps/api/internal/wallet/`) | ビジネスロジック（残高確認・予算設定・残高減算・月次リセット） |
| **LC-BUDGET-03** | `WalletRepository` | Backend / Go (`apps/api/internal/repo/wallet_repo/`) | DynamoDB `GoroPay_Wallet` への CRUD（ConsistentRead / UpdateItem with ConditionExpression） |
| **LC-BUDGET-04** | `BudgetSettingsRepository` | Backend / Go (`apps/api/internal/repo/budget_settings/`) | DynamoDB `GoroPay_BudgetSettings` への CRUD |
| **LC-BUDGET-05** | `IdempotencyRepository` | Backend / Go (`apps/api/internal/repo/idempotency/`) | DynamoDB `GoroPay_IdempotencyKeys` への冪等性キー管理（TTL 24h） |
| **LC-BUDGET-06** | `BudgetResetLogRepository` | Backend / Go (`apps/api/internal/repo/budget_reset_log/`) | DynamoDB `GoroPay_BudgetResetLog` への書き込み（PutItem with ConditionExpression） |
| **LC-BUDGET-07** | `ResetSchedulerHandler` | Backend / Go (`apps/scheduler/main.go`) | EventBridge Scheduler から呼ばれる Lambda ハンドラ |
| **LC-BUDGET-08** | `useWalletHook` | Frontend / TypeScript (`web/hooks/useWallet.ts`) | 残高・月間予算の取得（TanStack Query, stale 30s, `['balance']` query key） |
| **LC-BUDGET-09** | `BalanceDisplay` | Frontend / React (`web/components/budget/BalanceDisplay.tsx`) | 残高表示 UI（初回スケルトン / 2 回目以降 placeholderData 薄表示） |
| **LC-BUDGET-10** | `InsufficientBalanceModal` | Frontend / React (`web/components/budget/InsufficientBalanceModal.tsx`) | 残高不足時モーダル（ダメ化文言 + 増額誘導ボタン） |
| **LC-BUDGET-11** | `insufficientBalanceAtom` | Frontend / Jotai atom (`web/state/budget.ts`) | 残高不足モーダルの開閉状態管理 |
| **LC-BUDGET-12** | `BudgetSetupScreen` | Frontend / React (`web/app/(authenticated)/budget/page.tsx`) | 月間予算設定画面（SetBudget フォーム） |
| *(再利用)* | `ContextAwareSlogHandler` | Backend / Go (`apps/api/internal/logging/handler.go`) | **LC-AUTH-05 を再利用**。Unit B 独自の実装は作成しない（P-OBS-01） |
| *(再利用)* | `apiClient` | Frontend / TypeScript (`web/lib/apiClient.ts`) | **LC-AUTH-09 を再利用**。Unit B の `/api/wallet/*` リクエストもこれを使う |
| *(再利用)* | `BffProxyRouteHandler` | Frontend / Next.js (`web/app/api/[...path]/route.ts`) | **LC-AUTH-18 を再利用**。Unit B 専用の Route Handler ファイルは不要 |

---

## 2. Backend 論理コンポーネント

### LC-BUDGET-01: `WalletHandler`

| 項目 | 内容 |
|---|---|
| 配置 | `apps/api/internal/wallet/handler.go`（仮） |
| 公開 IF | `GetBalance(c *gin.Context)`, `SetBudget(c *gin.Context)`, `Deduct(c *gin.Context)` |
| 入力 | `gin.Context`（`userID` は Unit A の `AttachUserIDMiddleware` が注入済み） |
| 出力 | JSON レスポンス（[unit-interfaces.md §3.3](../../interfaces/unit-interfaces.md)） |
| エラーマッピング | `ErrInsufficientBalance` → 402, `ErrBudgetOutOfRange` → 400, `ErrIdempotencyConflict` → 409, その他 → 500 |
| ログ | `amount` / `newBalance` を明示 attrs で渡す（LC-AUTH-05 再利用、P-OBS-01）:<br>`slog.InfoContext(ctx, "deduct success", "action", "deduct", "amount", amount, "newBalance", newBalance)` |

### LC-BUDGET-02: `WalletService`

| 項目 | 内容 |
|---|---|
| 配置 | `apps/api/internal/wallet/service.go`（仮） |
| 公開 IF | [unit-interfaces.md §3.1](../../interfaces/unit-interfaces.md) の `WalletService` interface |
| 依存 | LC-BUDGET-03, LC-BUDGET-04, LC-BUDGET-05, LC-BUDGET-06 |
| `Deduct` ロジック | 1. `IdempotencyRepository.TryAcquire` → 冪等キー取得<br>2. `WalletRepository.Deduct`（ConditionExpression）→ P-REL-01<br>3. `IdempotencyRepository.SaveResponse` |
| `SetBudget` ロジック | 1. `BudgetSettingsRepository.Upsert`<br>2. `WalletRepository.SetBalance`（新規作成または差分調整）<br>エラー時はそのまま返す（P-REL-02） |
| `ResetAll` ロジック | Wallet テーブルを Scan → 各ユーザの `balance` を `BudgetSettings.monthlyBudget` にリセット → 失敗を `[]error` に収集（P-OBS-02） |

### LC-BUDGET-03: `WalletRepository`

| 項目 | 内容 |
|---|---|
| 配置 | `apps/api/internal/repo/wallet_repo/wallet_repo.go`（仮） |
| 公開 IF | `Get(ctx, userID) (*WalletRecord, error)`, `Deduct(ctx, userID, amount) (newBalance int, error)`, `SetBalance(ctx, userID, balance) error` |
| DynamoDB操作 | `Get`: GetItem with `ConsistentRead: true`<br>`Deduct`: UpdateItem with `ConditionExpression: balance >= :amount`<br>`SetBalance`: UpdateItem（条件なし） |
| エラー変換 | `ConditionalCheckFailedException` → `wallet.ErrInsufficientBalance`（P-REL-01） |
| テーブル | `GoroPay_Wallet`（PK: `userId`） |
| 注意 | Unit E の `WalletReader` interface を実装する（[unit-interfaces.md §3.2](../../interfaces/unit-interfaces.md)） |

### LC-BUDGET-04: `BudgetSettingsRepository`

| 項目 | 内容 |
|---|---|
| 配置 | `apps/api/internal/repo/budget_settings/budget_settings_repo.go`（仮） |
| 公開 IF | `Get(ctx, userID) (*BudgetSettings, error)`, `Upsert(ctx, userID, monthlyBudget, effectiveFrom) error` |
| DynamoDB操作 | `Get`: GetItem<br>`Upsert`: PutItem（条件なし、上書き） |
| テーブル | `GoroPay_BudgetSettings`（PK: `userId`） |
| 注意 | Unit E の `BudgetSettingsReader` / `BudgetSettingsWriter` interface を実装する（[unit-interfaces.md §3.2, §6.3](../../interfaces/unit-interfaces.md)） |

### LC-BUDGET-05: `IdempotencyRepository`

| 項目 | 内容 |
|---|---|
| 配置 | `apps/api/internal/repo/idempotency/idempotency_repo.go`（仮） |
| 公開 IF | `TryAcquire(ctx, key, payload) (acquired bool, existing *IdempotencyRecord, error)`, `SaveResponse(ctx, key, response) error` |
| DynamoDB操作 | `TryAcquire`: PutItem with `ConditionExpression: attribute_not_exists(key)`<br>`SaveResponse`: UpdateItem（response フィールドを更新） |
| TTL | `expiresAt` = now + 24h（DynamoDB TTL 自動削除） |
| テーブル | `GoroPay_IdempotencyKeys`（PK: `key`） |

### LC-BUDGET-06: `BudgetResetLogRepository`

| 項目 | 内容 |
|---|---|
| 配置 | `apps/api/internal/repo/budget_reset_log/budget_reset_log_repo.go`（仮） |
| 公開 IF | `Save(ctx, resetDate, userID, prevBalance, newBalance) error` |
| DynamoDB操作 | PutItem with `ConditionExpression: attribute_not_exists(resetDate) AND attribute_not_exists(userId)`（P-REL-03） |
| テーブル | `GoroPay_BudgetResetLog`（PK: `resetDate`（`YYYY-MM`）、SK: `userId`） |

### LC-BUDGET-07: `ResetSchedulerHandler`

| 項目 | 内容 |
|---|---|
| 配置 | `apps/scheduler/main.go` |
| ランタイム | Go（Lambda、EventBridge Scheduler から起動） |
| 振る舞い | `WalletService.ResetAll(ctx)` を呼び、結果（processedUsers / failedUsers）を INFO ログ出力 |
| タイムアウト | 30 秒（Q-N3=A） |
| メモリ | 128MB（Q-N9=A） |
| スケジュール | `cron(0 15 L * ? *)` UTC = 月末最終日 0:00 JST（Q-B8=A） |
| slog | LC-AUTH-05 `ContextAwareSlogHandler` を再利用（P-OBS-01） |

---

## 3. Frontend 論理コンポーネント

### LC-BUDGET-08: `useWalletHook`

| 項目 | 内容 |
|---|---|
| 配置 | `web/hooks/useWallet.ts`（仮） |
| 公開 IF | [unit-interfaces.md §9](../../interfaces/unit-interfaces.md) と一致:<br>`{ balance, monthlyBudget, isLoading, refetch() }` |
| 内部実装 | TanStack Query `useQuery({ queryKey: ['balance'], staleTime: 30_000, placeholderData: keepPreviousData })`（P-DEG-01） |
| API 呼び出し | `apiClient.request({ path: '/api/wallet' })`（**LC-AUTH-09 を再利用**、Authorization ヘッダ自動付与） |
| query key | `['balance']` — Unit C との **クロスユニット契約**（[unit-interfaces.md §9.1](../../interfaces/unit-interfaces.md)） |

> **クロスユニット依存（P-DEG-03）**: Unit C の `useOrder()` は注文成功時に `queryClient.invalidateQueries({ queryKey: ['balance'] })` を呼ぶ必要がある。`queryKey: ['balance']` は Unit B と Unit C の凍結契約。

### LC-BUDGET-09: `BalanceDisplay`

| 項目 | 内容 |
|---|---|
| 配置 | `web/components/budget/BalanceDisplay.tsx`（仮） |
| Props | `{ balance: number; monthlyBudget: number; isLoading: boolean; isFetching: boolean }` |
| 表示分岐 | `isLoading=true` → **スケルトン UI**（初回）<br>`isFetching=true && isLoading=false` → **薄い表示**（opacity 0.5）<br>通常 → **残高 + 月間予算を表示**（P-DEG-01） |
| 連携 | LC-BUDGET-08 `useWalletHook` からデータを受け取る |

### LC-BUDGET-10: `InsufficientBalanceModal`

| 項目 | 内容 |
|---|---|
| 配置 | `web/components/budget/InsufficientBalanceModal.tsx`（仮） |
| Props | `{ open: boolean; onClose(): void; onRaise(): void }` |
| 表示内容 | ダメ化文言（例: 「残りダメ予算が足りません…」）+ 増額誘導ボタン（`/budget` 遷移）+ 閉じるボタン（P-DEG-02） |
| 表示トリガー | LC-BUDGET-11 `insufficientBalanceAtom` が `true` になったとき |
| 連携 | Unit C の 402 受信 → LC-BUDGET-11 `atom.set(true)` → このコンポーネントが open |
| 注意 | 文言・アニメーション詳細は Code Generation で確定 |

### LC-BUDGET-11: `insufficientBalanceAtom`

| 項目 | 内容 |
|---|---|
| 配置 | `web/state/budget.ts`（仮） |
| 型 | `atom<boolean>(false)` |
| 振る舞い | Unit C の `useOrder()` が 402 を受信したときに `true` にセット。Modal が閉じたときに `false` にリセット（P-DEG-02） |
| Subscribe | LC-BUDGET-10 `InsufficientBalanceModal` |

### LC-BUDGET-12: `BudgetSetupScreen`

| 項目 | 内容 |
|---|---|
| 配置 | `web/app/(authenticated)/budget/page.tsx`（仮） |
| 振る舞い | 月間予算設定フォーム（1〜100,000 円）→ `POST /api/wallet/budget` を `apiClient.request(...)` 経由で呼ぶ（**LC-AUTH-09 を再利用**） |
| 成功時 | `queryClient.invalidateQueries({ queryKey: ['balance'] })` で残高を再取得 |
| エラー時 | バリデーションエラー（400）→ インラインメッセージ表示 |

---

## 4. Unit A 再利用コンポーネント（変更なし使用）

Unit B は以下の Unit A コンポーネントを **変更なしで再利用** する。新しいインスタンスや派生実装は作成しない。

| Unit A LC | 再利用方法 | 注意事項 |
|---|---|---|
| **LC-AUTH-05** `ContextAwareSlogHandler` | Unit B の Lambda 起動時に `slog.SetDefault(...)` で登録。`amount` / `newBalance` は明示 attrs で渡す（P-OBS-01） | context から 8 項目が自動付与される |
| **LC-AUTH-09** `apiClient` | Unit B の全 Frontend コンポーネントが `apiClient.request(...)` 経由で `/api/wallet/*` を呼ぶ | Authorization: Bearer <accessToken> は自動付与 |
| **LC-AUTH-18** `BffProxyRouteHandler` | `/api/wallet/*` は catch-all proxy が受けて API Gateway に転送 | Unit B 専用の Route Handler ファイル（`web/app/api/wallet/route.ts` 等）は作成しない |

---

## 5. コンポーネント関係図

```
┌──── Frontend (Next.js) ─────────────────────────────────────────────┐
│                                                                       │
│  ┌──────────────────┐    ┌─────────────────────────────────────────┐ │
│  │ BudgetSetupScreen│    │ BalanceDisplay (LC-09)                  │ │
│  │ (LC-12)          │    │ skeleton / placeholderData / normal     │ │
│  └────────┬─────────┘    └──────────────┬──────────────────────────┘ │
│           │                              │ isLoading / isFetching     │
│           │                              ▼                            │
│           │              ┌──────────────────────────────┐            │
│           │              │ useWalletHook (LC-08)         │            │
│           │              │ queryKey: ['balance']         │◄── Unit C  │
│           │              │ staleTime: 30s                │ invalidate │
│           │              └──────────────┬───────────────┘            │
│           │                              │ apiClient.request(...)     │
│           │                              ▼                            │
│           └────────────►┌──────────────────────────────┐             │
│                         │ apiClient (LC-AUTH-09)        │             │
│                         │ (Unit A 共有、再利用)          │             │
│                         └──────────────┬───────────────┘             │
│                                        │ /api/wallet/* fetch         │
│  ┌─────────────────────────────────┐   │ Authorization: Bearer       │
│  │ insufficientBalanceAtom (LC-11) │   │                             │
│  │ atom<boolean>(false)             │   │                             │
│  └────────┬────────────────────────┘   │                             │
│           │ true (Unit C が 402 受信時) │                             │
│           ▼                             │                             │
│  ┌──────────────────────────────────┐  │                             │
│  │ InsufficientBalanceModal (LC-10)  │  │                             │
│  │ ダメ化文言 + 増額誘導ボタン       │  │                             │
│  └──────────────────────────────────┘  │                             │
└────────────────────────────────────────┼─────────────────────────────┘
                                          │
                ┌── Next.js Server ────────┼──────────────────────────────┐
                │                          ▼                               │
                │  ┌────────────────────────────────────────────────────┐  │
                │  │ BffProxyRouteHandler (LC-AUTH-18)                  │  │
                │  │ /api/[...path]/route.ts (Unit A 共有、再利用)       │  │
                │  │ Authorization: Bearer <accessToken> を透過         │  │
                │  └─────────────────────────┬──────────────────────────┘  │
                └─────────────────────────────┼──────────────────────────────┘
                                              │ HTTPS → API Gateway
                                              ▼
                ┌── Lambda (Go + Gin + LWA) ──────────────────────────────────┐
                │                                                              │
                │  [AttachUserIDMiddleware (LC-AUTH-01)]                        │
                │  [RequestContextMiddleware (LC-AUTH-04)]                       │
                │                    ▼                                          │
                │  ┌────────────────────────────────┐                          │
                │  │ WalletHandler (LC-BUDGET-01)    │                          │
                │  └────────────────┬───────────────┘                          │
                │                   ▼                                           │
                │  ┌────────────────────────────────┐                          │
                │  │ WalletService (LC-BUDGET-02)    │                          │
                │  └──┬──────┬──────┬───────────────┘                          │
                │     │      │      │      │                                    │
                │     ▼      ▼      ▼      ▼                                    │
                │   LC-03  LC-04  LC-05  LC-06                                  │
                │  Wallet Budget Idem. Reset                                    │
                │   Repo   Repo   Repo   Log                                    │
                │     │      │      │      │                                    │
                │     └──────┴──────┴──────┘                                   │
                │                    │                                          │
                │                    ▼ DynamoDB                                 │
                │  GoroPay_Wallet / GoroPay_BudgetSettings /                   │
                │  GoroPay_IdempotencyKeys / GoroPay_BudgetResetLog            │
                │                                                              │
                │  ┌────────────────────────────────────────────────────────┐  │
                │  │ ContextAwareSlogHandler (LC-AUTH-05, 再利用)            │  │
                │  │ stdout → CloudWatch Logs                               │  │
                │  └────────────────────────────────────────────────────────┘  │
                └──────────────────────────────────────────────────────────────┘

                ┌── Scheduler Lambda (Go) ─────────────────────────────────────┐
                │  ResetSchedulerHandler (LC-BUDGET-07)                         │
                │  └─► WalletService.ResetAll() (LC-BUDGET-02)                  │
                │  EventBridge Scheduler: cron(0 15 L * ? *) UTC                │
                └──────────────────────────────────────────────────────────────┘
```

---

## 6. コンポーネント → パターン対応

| LC ID | 関連パターン |
|---|---|
| LC-BUDGET-01 | P-OBS-01, P-REL-01（エラーマッピング） |
| LC-BUDGET-02 | P-REL-01, P-REL-02, P-REL-03, P-OBS-02 |
| LC-BUDGET-03 | P-REL-01, P-PERF-01 |
| LC-BUDGET-04 | P-REL-02 |
| LC-BUDGET-05 | P-PERF-01（逐次実行） |
| LC-BUDGET-06 | P-REL-03 |
| LC-BUDGET-07 | P-OBS-02 |
| LC-BUDGET-08 | P-DEG-01, P-DEG-03 |
| LC-BUDGET-09 | P-DEG-01 |
| LC-BUDGET-10 | P-DEG-02 |
| LC-BUDGET-11 | P-DEG-02 |
| LC-BUDGET-12 | P-REL-02（SetBudget） |
| LC-AUTH-05（再利用） | P-OBS-01 |
| LC-AUTH-09（再利用） | P-DEG-03（invalidate 後の refetch 経路） |
| LC-AUTH-18（再利用） | BFF プロキシ（Unit A P-RES-01 パターン） |

---

## 7. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| **Infrastructure Design** | DynamoDB テーブル物理設計（PK/SK/GSI/TTL 詳細）、Lambda IAM Role（DynamoDB アクセス権限）、EventBridge Scheduler の Terraform 実装 |
| **Code Generation** | LC-BUDGET-01〜12 の実装本体、LC-AUTH-05/09/18 再利用コード、`go.mod` / `package.json` のバージョン pin、`gopter` PBT テストコード |
