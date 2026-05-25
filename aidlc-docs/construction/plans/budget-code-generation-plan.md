# Budget Unit (Unit B) — Code Generation Plan

**Document Version**: 0.1 (Draft, awaiting user approval)
**Created**: 2026-05-24
**Unit**: B (`budget` / ダメ予算)
**Construction Depth**: Standard
**Stage**: Code Generation (Construction Phase)
**Prerequisite**: Functional Design / NFR Requirements / NFR Design / Infrastructure Design 全て承認済み (PR #62, #74, #75, #77 マージ済み)

---

## 1. Plan の目的と範囲

本 Plan は、Unit B（`budget` / ダメ予算）の **Code Generation ステージ** を遂行するための作業計画を定義する。設計フェーズで凍結した仕様（FD Q-B1〜Q-B12 / NFR Req Q-N1〜Q-N11 / NFR Design Q-D1〜Q-D8 / Infra Design Q-I1〜Q-I6）に従い、Backend (Go) / Frontend (Next.js) / Infrastructure (Terraform) のコード・テスト・設定を生成する。

### 1.1 ストーリー範囲（unit-of-work-story-map.md より）

Unit B が実装するストーリー:

| Story ID | 内容 | 受入基準ポイント |
|---|---|---|
| **US-0-03** | ダメ予算を設定する | 1,000〜100,000 円の範囲、1,000 円刻み、即時反映、変更時の差分調整 |
| **US-0-04** | 初回メイン画面（残高初期表示） | Wallet 自動作成、`balance == monthlyBudget` |
| **US-1-02** | 残高を常に視認できる | MainScreen 上部固定、TanStack Query 30s + 即 invalidate |
| **US-1-04** | 残高不足時にダメになれない | 402 INSUFFICIENT_BALANCE、増額誘導モーダル |
| **US-1-05** | 二重引き落とし防止 | 冪等性キー `{userID}:{ulid}`、TTL 24h、失敗結果も保存 |
| **US-1-06** | 残高が負にならない | DynamoDB ConditionExpression: `balance >= :amount` |
| **US-3-05** | 月初リセット | EventBridge Scheduler、完全リセット、`(resetDate, userID)` 冪等性 |

### 1.2 Unit 依存と契約

unit-interfaces.md §3:
- **依存先**:
  - Unit A `AttachUserID` middleware（`gin.Context` から userID 取得）
  - Unit A `LC-AUTH-05 ContextAwareSlogHandler` 再利用（P-OBS-01）
  - Unit A `LC-AUTH-09 apiClient` 再利用（Frontend）
  - Unit A `LC-AUTH-18 BffProxyRouteHandler` 再利用（Frontend、catch-all proxy で `/api/wallet/*` を透過）
- **依存元**:
  - Unit C `OrderService.PlaceOrder`（`WalletService.Deduct` を Go 関数呼び出しで使用）
  - Unit E `MetricsService` / `BudgetRaiseService`（`WalletReader` / `BudgetSettingsReader` / `BudgetSettingsWriter` を読取・限定書込）
- **公開 Interface**: `WalletService`（`GetBalance` / `SetBudget` / `Deduct` / `ResetAll`）
- **公開 DTO**: `WalletSnapshot` / `DeductResult` / `ResetResult`
- **公開 Sentinel Error**: `wallet.ErrInsufficientBalance` / `wallet.ErrIdempotencyConflict` / `wallet.ErrBudgetOutOfRange`
- **Repository 公開 Interface**: `wallet_repo.WalletReader` / `budget_settings.BudgetSettingsReader` / `budget_settings.BudgetSettingsWriter`
- **公開 Output (Terraform module)**:
  - `module.budget`: 4 テーブルの `_table_name` / `_table_arn`、`scheduler_lambda_function_name` / `scheduler_lambda_arn`、`dynamodb_policy_arn`

### 1.3 NFR Design / Infrastructure Design 由来の主要設計判断

| 判断 | 由来 | 影響 |
|---|---|---|
| `ConditionalCheckFailedException` 即マップ（リトライなし） | P-REL-01 / Q-D1 | `internal/repo/wallet_repo/wallet_repo.go` |
| `SetBudget` 部分失敗はそのまま返す（補償なし） | P-REL-02 / Q-D2 | `internal/wallet/service.go` |
| `BudgetResetLog` 冪等性 = `attribute_not_exists` PutItem | P-REL-03 / Q-N4 | `internal/repo/budget_reset_log/` |
| Deduct p95 ≤ 500ms = 逐次実行 + ConsistentRead | P-PERF-01 / Issue #25 | `internal/wallet/service.go` の Deduct 実装 |
| `ContextAwareSlogHandler` 再利用 + amount/newBalance 明示 attrs | P-OBS-01 / Q-D5 | Handler / Service / Scheduler の slog 呼び出し |
| `ResetAll` ユーザ単位 ERROR + 集計 INFO | P-OBS-02 / Q-D8 | `internal/wallet/service.go` の ResetAll 実装 |
| `gopter` PBT（Unit A と統一、3 プロパティ） | P-TEST-01 / Q-D6 | `internal/wallet/service_pbt_test.go` |
| `BalanceDisplay` 初回スケルトン + `placeholderData` | P-DEG-01 / Q-D3 | `web/components/budget/BalanceDisplay.tsx` |
| `InsufficientBalanceModal` + Jotai atom | P-DEG-02 / Q-D4 | `web/components/budget/InsufficientBalanceModal.tsx` + `web/state/budget.ts` |
| Unit C invalidate `['balance']` クロスユニット契約 | P-DEG-03 / Q-D7 | Unit C 側の `useOrder` を更新（Unit B 側は受け側のみ） |
| 新規 `infra/modules/budget/` モジュール | Q-I1 = A | DynamoDB 4 + Scheduler Lambda + EventBridge + IAM + Log Group |
| Scheduler Lambda zip 方式（手動ビルド） | Q-I2 = A | `apps/scheduler/Makefile` で `bootstrap` 生成 |
| 既存 `api_gateway/routes.tf` に Unit B ルート追加 | Q-I (Unit C と同方式) | `infra/modules/api_gateway/` への追記 |
| 既存 `lambda_api/iam.tf` に DynamoDB ポリシー attach | Q-I (Unit C と同方式) | `additional_policy_arns` 経由 |
| API Lambda 環境変数追加（4 テーブル名） | Infra Design §5 | `envs/dev/main.tf` 経由で `module.lambda_api` に注入 |

### 1.4 Wallet スタブ撤廃

Unit A Code Generation で配線した `apps/api/wallet_stub.go` (`unconfiguredWalletService`) は Unit B 実装到達時点で **削除し、本物の `WalletService` 実装を `main.go` の DI で配線する** 。Unit C の `order.WalletService` interface (`internal/order/wallet.go`) はそのまま、Adapter を Unit B Service に差し替える。

---

## 2. プロジェクト構造（生成対象）

既存構造を踏襲、Unit B で追加・拡張するファイルを `★` で示す。

```
goro2pay/
├── apps/
│   ├── api/                                # API Lambda (Unit A/C 既存)
│   │   ├── main.go                         # ★ DI 配線追加 (Wallet 系)、wallet_stub.go 削除
│   │   ├── wallet_stub.go                  # ★ 削除（Unit B 実装到達のため）
│   │   └── internal/
│   │       ├── wallet/                     # ★ 新規パッケージ
│   │       │   ├── doc.go                  # ★ パッケージドキュメント
│   │       │   ├── errors.go               # ★ Sentinel: ErrInsufficientBalance / ErrIdempotencyConflict / ErrBudgetOutOfRange
│   │       │   ├── types.go                # ★ DTO: WalletSnapshot / DeductResult / ResetResult
│   │       │   ├── service.go              # ★ WalletService 本実装 (GetBalance/SetBudget/Deduct/ResetAll)
│   │       │   ├── service_test.go         # ★ ユニットテスト (mock repo)
│   │       │   ├── service_pbt_test.go     # ★ gopter PBT (3 プロパティ)
│   │       │   ├── handler.go              # ★ HTTP Handler (GET /api/wallet, POST /api/wallet/budget)
│   │       │   ├── handler_test.go         # ★ Handler テスト
│   │       │   ├── validation.go           # ★ VR-B-01〜06 純関数バリデーション
│   │       │   └── validation_test.go      # ★ 境界値テスト
│   │       ├── repo/
│   │       │   ├── wallet_repo/            # ★ 新規
│   │       │   │   ├── types.go            # WalletRecord (公開), WalletReader (Unit E 公開)
│   │       │   │   ├── repo.go             # DynamoDB 実装 (Get/DeductConditional/SetBalance/Create/ResetTo/ListAllUserIDs)
│   │       │   │   └── repo_test.go        # 軽量ユニット (mock dynamodb client)
│   │       │   ├── budget_settings/        # ★ 新規
│   │       │   │   ├── types.go            # BudgetSettings, RaiseLog, BudgetSettingsReader / Writer
│   │       │   │   ├── repo.go             # DynamoDB 実装 (Get/Set/UpdateRaiseHistory)
│   │       │   │   └── repo_test.go
│   │       │   ├── idempotency/            # ★ 新規
│   │       │   │   ├── types.go            # IdempotencyRecord
│   │       │   │   ├── repo.go             # DynamoDB 実装 (TryAcquire/SaveResponse)
│   │       │   │   └── repo_test.go
│   │       │   └── budget_reset_log/       # ★ 新規
│   │       │       ├── types.go
│   │       │       ├── repo.go             # DynamoDB 実装 (Get/Insert)
│   │       │       └── repo_test.go
│   │       └── order/
│   │           └── wallet.go                # 既存変更なし (Unit C 公開 IF)
│   └── scheduler/                          # ★ 新規 (月初リセット Lambda)
│       ├── main.go                         # ★ ResetSchedulerHandler (lambda.Start で起動)
│       ├── main_test.go                    # ★ ResetAll 呼び出しの軽量テスト
│       ├── go.mod                          # ★ 共有モジュールに含めるか別 module かは Step 1 で確定
│       └── Makefile                        # ★ GOOS=linux GOARCH=arm64 で bootstrap ビルド
│
├── web/                                    # Next.js Frontend (Unit A/C 既存)
│   ├── app/
│   │   └── (authenticated)/
│   │       └── budget/
│   │           ├── page.tsx                # ★ BudgetSetupScreen
│   │           └── page.test.tsx           # ★ Vitest コンポーネントテスト
│   ├── components/
│   │   └── budget/                         # ★ 新規ディレクトリ
│   │       ├── BalanceDisplay.tsx          # ★ 残高表示 (skeleton / placeholderData / normal)
│   │       ├── BalanceDisplay.test.tsx     # ★
│   │       ├── BudgetForm.tsx              # ★ クイックボタン + 数値入力 + 送信
│   │       ├── BudgetForm.test.tsx         # ★
│   │       ├── InsufficientBalanceModal.tsx # ★ ダメ化文言 + 増額誘導
│   │       └── InsufficientBalanceModal.test.tsx  # ★
│   ├── hooks/
│   │   ├── useWallet.ts                    # ★ TanStack Query, queryKey: ['balance'], staleTime 30s, placeholderData
│   │   ├── useWallet.test.ts               # ★
│   │   ├── useSetBudget.ts                 # ★ mutation + invalidate ['balance'] + redirect
│   │   └── useSetBudget.test.ts            # ★
│   ├── state/
│   │   ├── budget.ts                       # ★ insufficientBalanceAtom
│   │   └── budget.test.ts                  # ★
│   ├── lib/
│   │   └── api/
│   │       └── wallet.ts                   # ★ apiClient ラッパ（getWallet / postWalletBudget）
│
└── infra/                                  # Terraform (Unit A/C 既存)
    ├── envs/
    │   └── dev/
    │       └── main.tf                     # ★ module "budget" 呼出を追加 + lambda_api への env / policy 注入
    └── modules/
        ├── budget/                         # ★ 新規
        │   ├── README.md                   # ★
        │   ├── versions.tf                 # ★
        │   ├── variables.tf                # ★
        │   ├── locals.tf                   # ★ tags / 命名 prefix
        │   ├── dynamodb.tf                 # ★ 4 テーブル (Wallet/BudgetSettings/IdempotencyKeys/BudgetResetLog)
        │   ├── scheduler_lambda.tf         # ★ archive_file + Lambda + permission + Schedule
        │   ├── iam.tf                      # ★ Scheduler Lambda Role + Policy + EventBridge Scheduler Role + Policy + DynamoDB Policy (API Lambda 用、output)
        │   ├── log_groups.tf               # ★ Scheduler Lambda 用 Log Group
        │   ├── outputs.tf                  # ★ 4 table name/arn + scheduler arn + dynamodb_policy_arn
        │   └── tests/                      # ★ tftest 4 件
        │       ├── budget_basic.tftest.hcl
        │       ├── budget_dynamodb_tables.tftest.hcl
        │       ├── budget_scheduler_lambda.tftest.hcl
        │       └── budget_iam_least_privilege.tftest.hcl
        ├── api_gateway/                    # 既存
        │   └── routes.tf                   # ★ GET /api/wallet + POST /api/wallet/budget 追加
        └── lambda_api/                     # 既存
            └── (additional_policy_arns で attach、env で 4 テーブル名注入。envs/dev/main.tf で接続するのみで module 本体への変更は不要)
```

---

## 3. Step 一覧

各 Step は **Step 番号順に** 実行する。各 Step 完了時に Plan の checkbox `[ ]` を `[x]` に更新する。

### Step 1 — プロジェクト構造セットアップ

- [ ] 1.1 `apps/api/internal/wallet/` ディレクトリ作成（doc.go プレースホルダ）
- [ ] 1.2 `apps/api/internal/repo/{wallet_repo,budget_settings,idempotency,budget_reset_log}/` ディレクトリ作成
- [ ] 1.3 `apps/scheduler/` ディレクトリ作成（go.mod は **既存 `apps/api/go.mod` の `apps/api/...` パッケージを共有 import** する方針で別 module 化しない。`apps/scheduler/main.go` は `apps/api/internal/wallet` を直接 import）
- [ ] 1.4 `web/components/budget/` / `web/state/budget.ts` プレースホルダ作成
- [ ] 1.5 `infra/modules/budget/` ディレクトリ作成（versions.tf / variables.tf / locals.tf スケルトン）
- [ ] **完了確認**: `find apps/api/internal/wallet apps/api/internal/repo/wallet_repo ... -type d` で 4 + 1 ディレクトリ存在

### Step 2 — Backend Domain (errors / types / validation) 生成

unit-interfaces.md §3.1 と FD §4 に整合する公開シグネチャを実装する。

- [ ] 2.1 `internal/wallet/errors.go`: `ErrInsufficientBalance` / `ErrIdempotencyConflict` / `ErrBudgetOutOfRange` の `errors.New` 定義
- [ ] 2.2 `internal/wallet/types.go`: `WalletSnapshot` / `DeductResult` / `ResetResult` の構造体定義（凍結 IF §3.1 一致）
- [ ] 2.3 `internal/wallet/validation.go`: VR-B-01（範囲）/ VR-B-02（刻み）/ VR-B-03（amount > 0）/ VR-B-04（idempotencyKey 形式）/ VR-B-05（userID プレフィックス一致）/ VR-B-06（userID 必須）の純関数群
- [ ] **対応ストーリー**: US-0-03（VR-B-01/02）、US-1-05（VR-B-04/05）、US-1-06（VR-B-03）

### Step 3 — Backend Domain Unit Test

- [ ] 3.1 `internal/wallet/validation_test.go`: 各 VR-B-* の境界値（OK / NG）table-driven test
- [ ] 3.2 `errors.go` の sentinel が `errors.Is` で識別可能なことを確認するテスト
- [ ] **完了確認**: `cd apps/api && go test ./internal/wallet/... -run "TestValidate" -v` で全 PASS

### Step 4 — Backend Repository Layer 生成

各 Repository は AWS SDK Go v2 (`github.com/aws/aws-sdk-go-v2/service/dynamodb`) を使用。Unit A・Unit C と同じ pattern で **package-level `init()` で SDK Client を初期化** する（P-INIT-01、Unit A/C と統一）。

- [ ] 4.1 `internal/repo/wallet_repo/types.go`: `WalletRecord` 構造体（公開）+ `WalletReader` interface（Unit E 公開）
- [ ] 4.2 `internal/repo/wallet_repo/repo.go`:
  - `init()` で `dynamodb.NewFromConfig` 初期化、テーブル名は env `DDB_TABLE_WALLET`
  - `Get(ctx, userID)` — GetItem with `ConsistentRead: true`（PR-B-05）
  - `Create(ctx, userID, balance)` — PutItem with `attribute_not_exists(userId)`
  - `DeductConditional(ctx, userID, amount)` — UpdateItem with `ConditionExpression: balance >= :amount`、ConditionFailed → `wallet.ErrInsufficientBalance`（P-REL-01）
  - `UpdateBalance(ctx, userID, delta)` — `SET balance = balance + :delta`（増額用、SetBudget 変更パス）
  - `SetBalance(ctx, userID, balance)` — UpdateItem 無条件（減額打ち切り、ResetTo 兼用）
  - `ListAllUserIDs(ctx)` — Scan（ResetAll 専用）
- [ ] 4.3 `internal/repo/budget_settings/types.go`: `BudgetSettings` / `RaiseLog` / `BudgetSettingsReader` / `BudgetSettingsWriter`（Unit E 向け）
- [ ] 4.4 `internal/repo/budget_settings/repo.go`:
  - `Get(ctx, userID)`
  - `Set(ctx, userID, monthlyBudget, effectiveFrom)` — Upsert（PutItem）。`createdAt` は既存値保持、`updatedAt` のみ更新
- [ ] 4.5 `internal/repo/idempotency/types.go`: `IdempotencyRecord`
- [ ] 4.6 `internal/repo/idempotency/repo.go`:
  - `TryAcquire(ctx, key, payload, ttl)` — PutItem with `attribute_not_exists(key)`、ConditionFailed → `(acquired=false, existing=...)` を返す
  - `SaveResponse(ctx, key, response)` — UpdateItem
- [ ] 4.7 `internal/repo/budget_reset_log/types.go` / `repo.go`:
  - `Get(ctx, resetDate, userID)`
  - `Insert(ctx, log)` — PutItem with `attribute_not_exists(resetDate) AND attribute_not_exists(userId)`（P-REL-03）
- [ ] **対応ストーリー**: US-1-05（idempotency）、US-1-06（DeductConditional）、US-3-05（reset_log）

### Step 5 — Backend Repository Unit Test

各 Repository は DynamoDB Local もしくは `aws-sdk-go-v2` の `httptest` mock を使ったテストを実装。Unit A・Unit C で `mock_provider` を使った tftest があるので、ここでは **interface mock を使った Service 層テストを優先** し、Repository 層は最低限の sanity 確認に留める（時間制約）。

- [ ] 5.1 `internal/repo/wallet_repo/repo_test.go`: `DeductConditional` の ConditionalCheckFailed → `ErrInsufficientBalance` 変換のみテスト（mock dynamodb client）
- [ ] 5.2 `internal/repo/idempotency/repo_test.go`: `TryAcquire` の ConditionalCheckFailed → `acquired=false` のテスト
- [ ] 5.3 他 Repository は Service 層テストでカバー（`Service.Deduct` / `Service.SetBudget` / `Service.ResetAll`）
- [ ] **完了確認**: `go test ./internal/repo/...` PASS

### Step 6 — Backend Service Layer (WalletService) 生成

- [ ] 6.1 `internal/wallet/service.go`:
  - `Service` struct（依存: `WalletRepository` / `BudgetSettingsRepository` / `IdempotencyRepository` / `BudgetResetLogRepository` / `time.Now func` ← テスタビリティ用）
  - `NewService(...)` constructor
  - `GetBalance(ctx, userID)` (UC-B-03) — Wallet.Get + BudgetSettings.Get の合成
  - `SetBudget(ctx, userID, monthlyBudget)` (UC-B-01/02) — VR-B-01/02 → 初回判定（DR-B-01）→ 初回パス（BudgetSettings + Wallet 作成）or 変更パス（DR-B-02 差分調整、減額時 `min`）
  - `Deduct(ctx, userID, amount, idempotencyKey)` (UC-B-04) — VR-B-03/04/05 → TryAcquire → 冪等性ヒット判定（DR-B-03 hash 一致 / 不一致）→ DeductConditional → SaveResponse（成功 / 失敗結果も保存、PR-B-04）
  - `ResetAll(ctx)` (UC-B-05) — `ListAllUserIDs` → 各ユーザで `BudgetResetLog.Get` で skip 判定（DR-B-05）→ `BudgetSettings.Get` → `Wallet.ResetTo` → `BudgetResetLog.Insert`、失敗は `[]error` 集約（P-OBS-02）、`amount`/`newBalance` を slog 出力
- [ ] 6.2 公開 IF が unit-interfaces.md §3.1 の `WalletService` interface を満たすことを **コンパイル時チェック** で保証:
  ```go
  var _ WalletService = (*Service)(nil)
  ```
- [ ] 6.3 `Deduct` の payload hash 計算ヘルパー（`sha256`）を追加（DR-B-03）
- [ ] **対応ストーリー**: 全 7 ストーリー
- [ ] **対応 ルール**: VR-B-01〜06、DR-B-01〜05、CR-B-01〜05、PR-B-01〜06

### Step 7 — Backend Service Layer Unit Test

- [ ] 7.1 `internal/wallet/service_test.go`:
  - mock repo を function-field で構築（Unit C と同 pattern P-MOCK-01）
  - `TestService_GetBalance`: 正常 / Wallet 未作成
  - `TestService_SetBudget_FirstTime`: 初回パス（Wallet 作成）
  - `TestService_SetBudget_Increase`: 変更パス + 増額（balance += delta）
  - `TestService_SetBudget_Decrease`: 変更パス + 減額（min 打ち切り）
  - `TestService_SetBudget_OutOfRange`: VR-B-01 違反 → `ErrBudgetOutOfRange`
  - `TestService_Deduct_Success`
  - `TestService_Deduct_InsufficientBalance` → 失敗結果が IdempotencyRecord に保存されること（PR-B-04）
  - `TestService_Deduct_IdempotencyHit_SamePayload`: `Idempotent=true` 返却、DeductConditional 呼ばれない
  - `TestService_Deduct_IdempotencyConflict_DifferentPayload`: `ErrIdempotencyConflict`
  - `TestService_Deduct_FailureReplay`: 既存失敗 IdempotencyRecord → 同じエラー再生
  - `TestService_ResetAll_AllSuccess`
  - `TestService_ResetAll_PartialFailure`: 1 ユーザ失敗 → ERROR ログ + 集計 INFO + `Errors` に積まれる、他ユーザは継続
  - `TestService_ResetAll_SkipExisting`: 既存 BudgetResetLog → スキップ（DR-B-05）
- [ ] **完了確認**: `go test ./internal/wallet/... -v` PASS

### Step 8 — Backend Service Layer PBT (gopter)

- [ ] 8.1 `internal/wallet/service_pbt_test.go`:
  - `TestProperty_BalanceInvariant`: 任意の `(balance, amount)` で `balance >= amount` のとき `Deduct` 成功 → `newBalance == balance - amount`、`balance < amount` のとき `ErrInsufficientBalance`
  - `TestProperty_SetBudgetIdempotency`: `SetBudget(x); SetBudget(x)` の結果が一回呼びと同じ（変更パスでも delta=0 で no-op）
  - `TestProperty_ValidateMonthlyBudgetBoundary`: `[1, 100000]` 範囲内 + `% 1000 == 0` → success、それ以外 → `ErrBudgetOutOfRange`
- [ ] 8.2 Generation 数 100（Unit A 同様デフォルト、P-TEST-01）
- [ ] **完了確認**: `go test ./internal/wallet/... -run "TestProperty" -v` PASS、shrink で最小反例が表示されない（= プロパティが満たされる）

### Step 9 — Backend Handler / Routing 生成

- [ ] 9.1 `internal/wallet/handler.go`:
  - `Handler` struct (Service 依存)
  - `GetBalance(c *gin.Context)` — `auth.UserIDFromContext` → `Service.GetBalance` → JSON `{balance, monthlyBudget, updatedAt}` 200
  - `SetBudget(c *gin.Context)` — JSON Bind → `Service.SetBudget` → JSON `{monthlyBudget, appliedFrom}` 200（appliedFrom は now）
  - エラーマッピング: `ErrBudgetOutOfRange` → 400 / `ErrInsufficientBalance` → 402 / `ErrIdempotencyConflict` → 409 / `auth.ErrUnauthorized` → 401 / その他 → 500（凍結 IF §8）
  - slog: `action="get_balance" | "set_budget"`、Deduct 系は Unit C 経由で呼ばれるためここでは `amount`/`newBalance` 出力なし
- [ ] 9.2 `apps/api/main.go` への配線追加:
  - `wallet_stub.go` 削除
  - `wallet_repo.NewRepository()` / `budget_settings.NewRepository()` / `idempotency.NewRepository()` / `budget_reset_log.NewRepository()` 構築
  - `walletSvc := wallet.NewService(...)` 構築
  - 既存の `unconfiguredWalletService` を `walletSvc` に差し替え（`order.WalletService` interface には Adapter 経由で接続：`order.WalletService` の `Deduct(userID, idempotencyKey, amount)` 呼び出しを Unit B の `wallet.Service.Deduct(ctx, userID, amount, idempotencyKey)` に変換）
  - `walletHandler := wallet.NewHandler(walletSvc)` 構築
  - `api.GET("/wallet", walletHandler.GetBalance)` / `api.POST("/wallet/budget", walletHandler.SetBudget)` 登録
- [ ] 9.3 Adapter ファイル新設: `apps/api/internal/wallet/order_adapter.go`（or `apps/api/wallet_order_adapter.go`）
  - `order.WalletService` interface を実装し、内部で `wallet.Service.Deduct` に委譲
  - 引数順序の差異（`(userID, idempotencyKey, amount)` vs `(userID, amount, idempotencyKey)`）を吸収
  - 戻り値変換（`wallet.DeductResult` → `order.WalletDeductResult`、`OrderID` は idempotencyKey 由来 or 上位生成 ULID）
  - sentinel error 変換: `wallet.ErrInsufficientBalance` → `order.ErrInsufficientFunds`、`wallet.ErrIdempotencyConflict` → `order.ErrIdempotencyConflict`
- [ ] **対応ストーリー**: US-0-03、US-0-04、US-1-02、US-1-04、US-1-05、US-1-06

### Step 10 — Backend Handler Unit Test

- [ ] 10.1 `internal/wallet/handler_test.go`:
  - mock service を `httptest` で叩く
  - `TestHandler_GetBalance_OK`
  - `TestHandler_SetBudget_OK`
  - `TestHandler_SetBudget_BadRequest`（VR 違反 → 400）
  - `TestHandler_SetBudget_OutOfRange`（`ErrBudgetOutOfRange` → 400）
  - `TestHandler_Unauthorized`（userID なし → 401）
- [ ] 10.2 `internal/wallet/order_adapter_test.go`:
  - `TestAdapter_Deduct_Success`
  - `TestAdapter_Deduct_InsufficientBalance` → `order.ErrInsufficientFunds`
  - `TestAdapter_Deduct_IdempotencyConflict` → `order.ErrIdempotencyConflict`

### Step 11 — Scheduler Lambda 生成

- [ ] 11.1 `apps/scheduler/main.go`:
  - `lambda.Start(handler)` パターン
  - 起動時に `slog.SetDefault(logging.NewContextAwareSlogHandler(...))` で Unit A の Handler を再利用（P-OBS-01）
  - `handler(ctx, event interface{})` 内で `wallet.Service` を構築（Repository は env から SDK 初期化）→ `walletSvc.ResetAll(ctx)` 呼び出し
  - 集計 INFO ログ出力
- [ ] 11.2 `apps/scheduler/main_test.go`:
  - `WalletService` の関数型 mock で `ResetAll` が呼ばれることだけ確認（軽量）
- [ ] 11.3 `apps/scheduler/Makefile`:
  - `build` ターゲット: `GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o bootstrap .`
  - `clean` ターゲット
- [ ] 11.4 `apps/scheduler/README.md`: ビルド手順 + デプロイ前提 (`make build` 必須)
- [ ] **対応ストーリー**: US-3-05

### Step 12 — Frontend API ラッパ + Hooks 生成

- [ ] 12.1 `web/lib/api/wallet.ts`:
  - `getWallet(): Promise<{balance, monthlyBudget, updatedAt}>` — `apiClient.request({path: '/api/wallet'})`（LC-AUTH-09 再利用）
  - `postWalletBudget(monthlyBudget): Promise<{monthlyBudget, appliedFrom}>` — `apiClient.request({path: '/api/wallet/budget', method: 'POST', body: {monthlyBudget}})`
  - エラーマッピング（400 / 401 / 5xx）: `apiClient` 既存と統一
- [ ] 12.2 `web/hooks/useWallet.ts`:
  - `useQuery({ queryKey: ['balance'], queryFn: getWallet, staleTime: 30_000, placeholderData: keepPreviousData })`
  - return: `{balance, monthlyBudget, isLoading, isFetching, refetch}`（凍結 IF §9 と一致、`isFetching` は内部用に追加 — 公開シグネチャから漏れる場合は `BalanceDisplay` 内で直接 `useQuery` を呼ぶ方式に変更）

  > **凍結 IF 整合判断**: 凍結 IF §9 は `{balance, monthlyBudget, isLoading, refetch}` のみ公開。`isFetching` は **公開しない** とし、`BalanceDisplay` が internally `useQuery({ queryKey: ['balance'] })` を直接 subscribe する（FD §3.1 注通り）

- [ ] 12.3 `web/hooks/useWallet.ts` の最終形: 凍結 IF §9 と完全一致 (`{balance, monthlyBudget, isLoading, refetch}`)
- [ ] 12.4 `web/hooks/useSetBudget.ts`:
  - `useMutation({ mutationFn: postWalletBudget, onSuccess: () => { invalidate(['balance']); router.push('/'); } })`
  - return: `{mutate, isPending, error}`
- [ ] **クロスユニット契約**: `queryKey: ['balance']` を Unit B 側で確定。Unit C 側の `useOrder` が成功時に invalidate 呼び出すこと（既存 Unit C 実装を確認、未実装なら ★ 追記）

### Step 13 — Frontend Hooks Unit Test

- [ ] 13.1 `web/hooks/useWallet.test.ts`: mock `apiClient` で 200 / 401 を返し、`isLoading` 遷移 / refetch 動作を確認
- [ ] 13.2 `web/hooks/useSetBudget.test.ts`: mock `apiClient` で 200 を返し、`onSuccess` で `invalidateQueries` が呼ばれることを spy で確認

### Step 14 — Frontend State (Jotai atom)

- [ ] 14.1 `web/state/budget.ts`: `insufficientBalanceAtom = atom<boolean>(false)`
- [ ] 14.2 `web/state/budget.test.ts`: 初期値 false、true → false 遷移の確認

### Step 14.5 — Frontend Unit C 連携確認 (P-DEG-03)

- [ ] 14.5.1 既存 `web/hooks/useOrder.ts` の onSuccess に `queryClient.invalidateQueries({queryKey: ['balance']})` が **無ければ追加**（P-DEG-03 / クロスユニット契約）
- [ ] 14.5.2 既存 `web/hooks/useOrder.ts` の 402 ハンドリングに `setInsufficientBalance(true)` を **無ければ追加**（P-DEG-02）

### Step 15 — Frontend Components 生成

- [ ] 15.1 `web/components/budget/BalanceDisplay.tsx`:
  - 内部で `useQuery({ queryKey: ['balance'], ... })` を直接 subscribe（FD §3.1 注通り、isError アクセス用）
  - 表示分岐: `isLoading=true` → スケルトン / `isFetching && data` → opacity 0.5 / 通常 → 残高 + ラベル + ResetCountdown
  - `data-testid` 付与: `budget-balance-display`、`budget-balance-amount`
  - 警告色（消化率 > 0.8 で赤）
- [ ] 15.2 `web/components/budget/BudgetForm.tsx`:
  - クイックボタン 5 個（10/30/50/80/100k、30k に `★` ハイライト）+ 数値入力 + 送信ボタン
  - リアルタイムバリデーション（VR-B-01 / VR-B-02）
  - `data-testid`: `budget-form`、`budget-quick-{value}`、`budget-input`、`budget-submit`
- [ ] 15.3 `web/components/budget/InsufficientBalanceModal.tsx`:
  - `useAtom(insufficientBalanceAtom)` で open 制御
  - 「残りダメ予算が足りません…」+ 「もっとダメになる」ボタン（`router.push('/budget')`）+ 閉じるボタン
  - `data-testid`: `insufficient-balance-modal`、`raise-budget-button`、`close-modal-button`
- [ ] 15.4 `web/app/(authenticated)/budget/page.tsx`: `BudgetForm` を埋め込んだ `BudgetSetupScreen`
- [ ] **MainScreen 統合**: 既存 `web/app/page.tsx` に `BalanceDisplay` を埋め込む。Unit C `MainScreen` 既存実装を変更しないために、page.tsx 上部に直接 import（既存 layout を破壊しない）

### Step 16 — Frontend Components Unit Test (Vitest)

- [ ] 16.1 `BalanceDisplay.test.tsx`: 初回ロード時スケルトン / 2 回目以降 placeholderData 表示 / 警告色判定
- [ ] 16.2 `BudgetForm.test.tsx`: クイックボタンクリック / 数値入力 / バリデーションエラー表示 / 送信ボタン disabled 制御
- [ ] 16.3 `InsufficientBalanceModal.test.tsx`: atom true → open / 増額誘導ボタンクリック → router.push / 閉じる → atom false
- [ ] 16.4 `web/app/(authenticated)/budget/page.test.tsx`: 初回ユーザ default 30k 表示 / 既存ユーザ現在値 prefill

### Step 17 — Infrastructure: 新規 `infra/modules/budget/` 生成

- [ ] 17.1 `versions.tf`: terraform >= 1.6 / aws ~> 5.x（Unit A・C と統一）
- [ ] 17.2 `variables.tf`: `env`, `region`, `api_lambda_role_name`（Unit A `lambda_api` の Role 名 — DynamoDB Policy attach 用）
- [ ] 17.3 `locals.tf`: `name_prefix = "gp-${var.env}"`、tags（Unit=budget 含む）
- [ ] 17.4 `dynamodb.tf`: 4 テーブル定義（infrastructure-design.md §3.1）
  - `wallet`: PK `userId`、PROVISIONED 1/1
  - `budget_settings`: PK `userId`、PROVISIONED 1/1
  - `idempotency_keys`: PK `key`、TTL `expiresAt`
  - `budget_reset_log`: PK `resetDate` / SK `userId`
  - 全テーブル AES256 暗号化、PITR=false
- [ ] 17.5 `iam.tf`:
  - `aws_iam_role.scheduler_lambda`（assume: lambda.amazonaws.com）
  - `aws_iam_role_policy.scheduler_lambda` — Wallet Get/Update/Scan + BudgetSettings Get + BudgetResetLog PutItem/GetItem + Logs (infrastructure-design.md §3.4.2)
  - `aws_iam_role.eventbridge_scheduler`（assume: scheduler.amazonaws.com）
  - `aws_iam_role_policy.eventbridge_scheduler_invoke` — `lambda:InvokeFunction` on Scheduler Lambda only
  - **API Lambda 用 Policy として `aws_iam_policy.dynamodb_access`**（Unit C `order_history` と同 pattern）— Get/Put/Update/Delete/Query/Scan on 4 テーブル → outputs で `dynamodb_policy_arn` を出力 → `envs/dev/main.tf` 経由で `module.lambda_api.additional_policy_arns` に渡す
- [ ] 17.6 `log_groups.tf`: `/aws/lambda/gp-${env}-scheduler-fn`、retention 7 日
- [ ] 17.7 `scheduler_lambda.tf`:
  - `data "archive_file" "scheduler"` — `apps/scheduler/bootstrap` を zip 化
  - `aws_lambda_function.scheduler` — runtime `provided.al2023`、handler `bootstrap`、arch `arm64`、memory 128MB、timeout 30s、env DDB_TABLE_*
  - `aws_lambda_permission.eventbridge_invoke_scheduler`
  - `aws_scheduler_schedule.monthly_reset` — `cron(0 15 L * ? *)`、`flexible_time_window.mode = "OFF"`、`retry_policy.maximum_retry_attempts = 0`
- [ ] 17.8 `outputs.tf`:
  - `wallet_table_name` / `wallet_table_arn`
  - `budget_settings_table_name` / `budget_settings_table_arn`
  - `idempotency_keys_table_name` / `idempotency_keys_table_arn`
  - `budget_reset_log_table_name` / `budget_reset_log_table_arn`
  - `scheduler_lambda_function_name` / `scheduler_lambda_arn`
  - `dynamodb_policy_arn`（API Lambda attach 用）
- [ ] 17.9 `README.md`: モジュール概要 + 入出力一覧 + ビルド前提（`make -C apps/scheduler build` 必須）

### Step 18 — Infrastructure: 既存 `api_gateway/routes.tf` への Unit B ルート追記

- [ ] 18.1 `infra/modules/api_gateway/routes.tf` (既存) に追記:
  - `aws_apigatewayv2_route.get_wallet` — `GET /api/wallet`
  - `aws_apigatewayv2_route.post_wallet_budget` — `POST /api/wallet/budget`
  - 両方とも `authorization_type = "JWT"` + Cognito Authorizer（Unit C と同 pattern）
- [ ] 18.2 既存ファイルが Unit C 用に整理済みかを確認（必要なら `routes.tf` 新設）

### Step 19 — Infrastructure: `envs/dev/main.tf` 配線

- [ ] 19.1 `envs/dev/main.tf` に以下を追加:
  ```hcl
  module "budget" {
    source = "../../modules/budget"
    env    = local.env
    region = local.region
  }
  ```
- [ ] 19.2 既存 `module "lambda_api"` の `additional_policy_arns` に `module.budget.dynamodb_policy_arn` を追加（Unit C order_history / bedrock と同 pattern）
- [ ] 19.3 既存 `module "lambda_api"` への `environment` 追加方法を確認:
  - 既存 `lambda_api/api_lambda.tf` で `environment` をどう受け取っているか確認
  - 4 テーブル名（`DDB_TABLE_WALLET` / `DDB_TABLE_BUDGET_SETTINGS` / `DDB_TABLE_IDEMPOTENCY` / `DDB_TABLE_BUDGET_RESET_LOG`）を `lambda_api` モジュール側で受け取れるよう、必要なら variable 拡張
- [ ] 19.4 既存 `module "api_gateway"` の input variables に Unit B ルート追加に必要な変数（Cognito Authorizer ID 等）が含まれていることを確認

### Step 20 — Infrastructure: terraform-test 4 件

- [ ] 20.1 `tests/budget_basic.tftest.hcl` — `mock_provider` で plan が成立することのみ確認
- [ ] 20.2 `tests/budget_dynamodb_tables.tftest.hcl` — 4 テーブルの billing_mode / hash_key / TTL 設定が NFR 一致
- [ ] 20.3 `tests/budget_scheduler_lambda.tftest.hcl` — Scheduler Lambda の memory=128 / timeout=30 / arch=arm64 / runtime=provided.al2023 / cron `0 15 L * ? *`
- [ ] 20.4 `tests/budget_iam_least_privilege.tftest.hcl` — Scheduler Role が Wallet Scan / BudgetSettings Get のみ（OrderHistory には触れない）、EventBridge Role が `lambda:InvokeFunction` のみ
- [ ] **完了確認**: `cd infra/modules/budget && terraform init && terraform test` で全 PASS

### Step 21 — Documentation: code/ 配下の Markdown 要約生成

- [ ] 21.1 `aidlc-docs/construction/budget/code/README.md` — 生成成果物 一覧（Backend / Frontend / Infra ファイル数 + コアファイル）
- [ ] 21.2 `aidlc-docs/construction/budget/code/backend-summary.md` — wallet パッケージ + 4 Repository + Adapter + Scheduler Lambda 概要
- [ ] 21.3 `aidlc-docs/construction/budget/code/frontend-summary.md` — Hooks 2 + Components 3 + State 1 + API ラッパ 1 概要
- [ ] 21.4 `aidlc-docs/construction/budget/code/infrastructure-summary.md` — module/budget/ 一覧 + envs/dev 配線 + tftest 4 件
- [ ] 21.5 `aidlc-docs/construction/budget/code/deployment-runbook.md` — デプロイ手順:
  1. `make -C apps/scheduler build` で `bootstrap` 生成
  2. `cd infra/envs/dev && terraform init && terraform plan && terraform apply`
  3. CodePipeline で API Lambda が再ビルド・デプロイされる（既存 CI/CD）
  4. Frontend は Amplify Hosting で自動デプロイ
  5. 動作確認: `/budget` でフォーム送信、MainScreen で `BalanceDisplay` 表示、`POST /api/orders` で 402 → モーダル表示

### Step 22 — Build & Sanity Check

- [ ] 22.1 `cd apps/api && go build ./... && go vet ./...` PASS
- [ ] 22.2 `cd apps/api && go test ./... -count=1` 全 PASS（Unit A/C テストの regression 含む）
- [ ] 22.3 `cd apps/scheduler && go build ./... && go vet ./...` PASS
- [ ] 22.4 `cd web && npm run lint && npx vitest run` 全 PASS
- [ ] 22.5 `cd infra/modules/budget && terraform init && terraform validate && terraform fmt -check` PASS
- [ ] 22.6 `cd infra/modules/budget && terraform test` 全 PASS
- [ ] 22.7 `cd infra/envs/dev && terraform init && terraform validate && terraform plan` でエラーなし（実 apply は別フェーズ）

### Step 23 — Plan Checkbox 更新 + aidlc-state.md 更新

- [ ] 23.1 全 Step の `[ ]` を `[x]` に更新（実行時に都度更新）
- [ ] 23.2 `aidlc-docs/aidlc-state.md` の Unit B 進捗ブロックの **Code Generation (Unit B)** を `[x]` に
- [ ] 23.3 `aidlc-docs/audit.md` に Code Generation 完了エントリを追記

---

## 4. ストーリー × Step トレーサビリティ

| Story ID | 主要 Step |
|---|---|
| US-0-03 ダメ予算を設定する | 2 (validation), 6 (SetBudget), 9 (handler), 12-13 (useSetBudget), 15-16 (BudgetForm) |
| US-0-04 初回メイン画面（残高初期表示） | 6 (SetBudget 初回 → Wallet 自動作成), 12 (useWallet), 15 (BalanceDisplay) |
| US-1-02 残高を常に視認できる | 12 (useWallet staleTime/placeholderData), 14.5 (Unit C invalidate), 15 (BalanceDisplay) |
| US-1-04 残高不足時にダメになれない | 9 (handler 402 mapping), 14 (atom), 15 (InsufficientBalanceModal) |
| US-1-05 二重引き落とし防止 | 4.6 (idempotency repo), 6 (Deduct + DR-B-03), 7 (idempotency tests), 8 (PBT) |
| US-1-06 残高が負にならない | 4.2 (DeductConditional), 6 (Deduct), 7-8 (tests) |
| US-3-05 月初リセット | 4.7 (reset_log repo), 6 (ResetAll), 7 (tests), 11 (Scheduler Lambda), 17.7 (EventBridge) |

---

## 5. NFR / Pattern × Step トレーサビリティ

| NFR / Pattern | 該当 Step |
|---|---|
| P-REL-01 ConditionalCheck 即マップ | 4.2 (DeductConditional) |
| P-REL-02 SetBudget 部分失敗 | 6.1 (Service.SetBudget) |
| P-REL-03 BudgetResetLog 冪等性 | 4.7 (Insert with ConditionExpression) |
| P-PERF-01 Deduct 500ms | 4.2 (ConsistentRead) + 6.1 (逐次実行) |
| P-OBS-01 ContextAwareSlogHandler 再利用 | 9 (Handler), 11 (Scheduler), `amount`/`newBalance` 明示 attrs |
| P-OBS-02 ResetAll エラーログ | 6.1 (`ResetAll` ループ) |
| P-TEST-01 gopter PBT (3 props) | 8 |
| P-DEG-01 BalanceDisplay Loading | 15.1 |
| P-DEG-02 InsufficientBalanceModal | 14, 14.5.2, 15.3 |
| P-DEG-03 invalidate from Unit C | 14.5.1 (Unit C 側の補強) |
| NFR-COMP-01〜03 仮想ウォレット | 全実装で「整数値のみ」を堅持 |

---

## 6. 既存実装との衝突ポイントと対処

| ポイント | 既存状態 | 対処 |
|---|---|---|
| `apps/api/wallet_stub.go` | Unit A で配線済みの `unconfiguredWalletService` | Step 9 で **削除し**、本物の `wallet.Service` を Adapter 経由で `order.WalletService` に注入 |
| `apps/api/main.go` の `walletStub` 配線 | `newUnconfiguredWalletService()` を `order.NewService` に渡している | Step 9 で `wallet.NewService(...)` + Adapter 構築に差し替え |
| `web/hooks/useOrder.ts` の 402 / invalidate | 現状 `['balance']` invalidate / `insufficientBalanceAtom.set(true)` の有無不明 | Step 14.5 で確認 → 不足分を補追記 |
| `web/app/page.tsx` (MainScreen) | 既存 Unit C 実装 | Step 15 で `BalanceDisplay` を埋め込み（既存レイアウトを破壊しない最小追加） |
| `infra/modules/api_gateway/routes.tf` | Unit A/C のルート定義済み | Step 18 で Unit B ルート 2 本を追記 |
| `infra/modules/lambda_api/` の env | DDB_TABLE_ORDER_HISTORY 等 Unit C 由来 env を受け取る variable がある想定 | Step 19 で variable 拡張または既存 pattern に従って配線 |

---

## 7. 並列化の余地（参考、本 Plan は逐次実行）

| 並列ブロック | Step |
|---|---|
| Backend Domain + Repository | Step 2-5（並列実行可、ただし依存関係に注意） |
| Frontend + Backend | Step 12-16 と Step 6-11 は独立して進行可 |
| Infrastructure | Step 17-20 は Backend / Frontend と独立 |

ただし Build & Sanity (Step 22) は最後に実行し、全領域の整合性を保証する。

---

## 8. 凍結 IF / 設計成果物への参照

- [unit-interfaces.md §3](../interfaces/unit-interfaces.md) — Unit B 公開 IF 全体
- [functional-design/business-logic-model.md](../budget/functional-design/business-logic-model.md) — UC-B-01〜06
- [functional-design/business-rules.md](../budget/functional-design/business-rules.md) — VR / DR / CR / PR
- [functional-design/domain-entities.md](../budget/functional-design/domain-entities.md) — エンティティ定義
- [functional-design/frontend-components.md](../budget/functional-design/frontend-components.md) — UI 仕様
- [nfr-requirements.md](../budget/nfr-requirements/nfr-requirements.md) — NFR 数値
- [nfr-design/nfr-design-patterns.md](../budget/nfr-design/nfr-design-patterns.md) — P-REL/PERF/OBS/TEST/DEG パターン
- [nfr-design/logical-components.md](../budget/nfr-design/logical-components.md) — LC-BUDGET-01〜12
- [infrastructure-design/infrastructure-design.md](../budget/infrastructure-design/infrastructure-design.md) — Terraform 凍結

---

## 9. 完了基準

- 全 23 Step の checkbox が `[x]`
- Build & Sanity (Step 22) が全項目 PASS
- 全ユニットテスト・PBT・tftest が PASS
- `aidlc-docs/aidlc-state.md` の Code Generation (Unit B) が `[x]`
- `aidlc-docs/audit.md` に完了記録
