# Unit C (`order`) — Logical Components

**Document Version**: 1.0
**Created**: 2026-05-23
**Stage**: Construction / NFR Design
**Unit**: C — `order` (代行手配コア)
**Depth**: Comprehensive
**Related**: [nfr-design-patterns.md](./nfr-design-patterns.md), [nfr-requirements.md](../nfr-requirements/nfr-requirements.md)

本ドキュメントは Unit C `order` の **論理コンポーネント一覧** を、設計パターン (P-RETRY-01〜P-FE-LOCK-01) を実装する具体的なクラス・関数・hook・atom として網羅する。Comprehensive 深度として全コンポーネントを ID 付き（`LC-ORDER-xx`）で管理する。

---

## 1. コンポーネント識別子規則

- 形式: `LC-ORDER-{番号}`
- 番号: 連番（カテゴリ無関係に通し番号、ID 安定化のため）
- 各コンポーネントに **責務 / 配置 / 依存先 / 由来パターン** を明記

---

## 2. Backend コンポーネント

### LC-ORDER-01: OrderService

- **責務**: 注文ユースケースのオーケストレーション（冪等性チェック / Wallet.Deduct / DeliveryAdapter.Place / OrderHistory.Insert / 結果組み立て）。フォールバック分岐は PlanBuilder に委譲。
- **配置**: `apps/api/internal/order/service.go`
- **公開 interface**: `OrderService` (凍結契約 §4.1)
- **公開メソッド**: `PlaceOrder(ctx, userID, req) (*PlaceOrderResult, error)` / `GetHistory(ctx, userID, limit) ([]*OrderRecord, error)`
- **依存先**:
  - `OrderHistoryRepository` (LC-ORDER-04)
  - `PlanBuilder` (LC-ORDER-08)
  - `DeliveryAdapter` (LC-ORDER-06)
  - `WalletService` (Unit B、unit-interfaces.md §3.1)
- **由来パターン**: P-DI-01 (manual DI)、P-OBS-02 (LogSummary 利用)

---

### LC-ORDER-02: OrderHandler

- **責務**: HTTP リクエストの受付・パース / OrderService 呼出 / レスポンス組み立て。Gin handler。
- **配置**: `apps/api/internal/order/handler.go`
- **登録ルート**:
  - `POST /api/orders` → `PlaceOrder`
  - `GET /api/orders` → `GetHistory`
- **エラーマッピング**（凍結契約 §4.3）:
  - `wallet.ErrInsufficientFunds` → 402 INSUFFICIENT_FUNDS
  - `wallet.ErrIdempotencyConflict` → 409 IDEMPOTENCY_CONFLICT (BR-C39)
  - `context.DeadlineExceeded` → 504 GATEWAY_TIMEOUT
  - その他 → 500 INTERNAL_ERROR
- **依存先**: `OrderService` (LC-ORDER-01)
- **由来パターン**: P-DI-01

---

### LC-ORDER-03: PlaceOrderRequest / PlaceOrderResult / OrderRecord（DTO）

- **責務**: Unit C 公開 DTO（凍結契約 §4.1）+ Unit C 内部追加属性
- **配置**: `apps/api/internal/order/types.go`
- **公開フィールド**（凍結契約整合）:
  - `PlaceOrderRequest`: Category / IdempotencyKey / SuggestionID
  - `PlaceOrderResult`: OrderID / StoreName / MenuName / Amount / RemainingBalance / Idempotent
  - `OrderRecord`: OrderID / UserID / Category / StoreName / MenuName / Amount / OrderedAt（凍結契約 §4.1 の 7 フィールド）
- **Unit C 内部追加属性**（domain-entities.md §2.1.2）:
  - `IdempotencyKey` / `DayOfWeek` / `Source` / `ExpiresAt`
- **由来**: 凍結契約 §4.1、Functional Design domain-entities.md

---

### LC-ORDER-04: OrderHistoryRepository

- **責務**: DynamoDB `GoroPay_OrderHistory` テーブルへの Insert / Query。
- **配置**: `apps/api/internal/repo/order_history/repository.go`
- **公開 interface**: `OrderHistoryRepository`
- **公開メソッド**:
  - `Insert(ctx, record *OrderRecord) error`
  - `Query(ctx, userID string, limit int) ([]*OrderRecord, error)`（最大 100、降順、TTL 90 日内）
- **依存先**: `dynamodb.Client` (P-INIT-01 で package-level 変数)
- **由来パターン**: P-INIT-01

---

### LC-ORDER-05: BedrockAdapter

- **責務**: Bedrock Claude 3.5 Haiku Converse API 呼出。プロンプト組み立て、レスポンスパース、エラー伝播。
- **配置**: `apps/api/internal/adapters/bedrock/adapter.go`（横串、Unit C / D 共有）
- **公開 interface**: `BedrockAdapter`
- **公開メソッド**: `InferOrderPlan(ctx, history []OrderRecord) (*Plan, error)`
- **implementation**: `ClaudeBedrockAdapter`
- **依存先**: `bedrockruntime.Client` (P-INIT-01 で package-level 変数)、`RetryClassifier` (LC-ORDER-07)
- **由来パターン**: P-INIT-01、P-RETRY-01
- **テストモード**: `NewClaudeBedrockAdapterWithClient(c BedrockRuntimeAPI)` で mock client 注入可能

---

### LC-ORDER-06: DeliveryAdapter / MockDeliveryAdapter

- **責務**: 外部デリバリー手配のアダプタ層（MVP は MockDeliveryAdapter のみ実装）。
- **配置**: `apps/api/internal/adapters/delivery/adapter.go`（横串、Unit C / D 共有）
- **公開 interface**: `DeliveryAdapter`
- **公開メソッド**: `Place(ctx, plan *Plan, idempotencyKey string) error`
- **implementation**: `MockDeliveryAdapter`（noop で `nil` を返す、ログ出力のみ）
- **由来**: Application Design Q-F=A、unit-of-work.md §3.3

---

### LC-ORDER-07: RetryClassifier / BedrockRetryClassifier

- **責務**: Bedrock 呼出失敗時の「リトライすべきか / 即フォールバックか」を判定。AWS SDK v2 のエラー型を `errors.Is` / `errors.As` で判定。
- **配置**: `apps/api/internal/adapters/bedrock/retry.go`（横串、Unit C / D 共有）
- **公開 interface**: `RetryClassifier`
- **公開メソッド**: `ShouldRetry(err error) bool`
- **implementation**: `BedrockRetryClassifier`
- **判定対象**:
  - リトライ: `*types.ThrottlingException` / `*types.ServiceUnavailableException` / `context.DeadlineExceeded` / `*smithy.GenericAPIError` (FaultServer)
  - 即フォールバック: `*types.ValidationException` / `*types.AccessDeniedException` / `*types.ResourceNotFoundException`
- **由来パターン**: P-RETRY-01

---

### LC-ORDER-08: PlanBuilder / BedrockPlanBuilder

- **責務**: Bedrock 呼出 + リトライ + フォールバック分岐の一連のロジックを集約。`Plan` を返す（成功 or フォールバック）。
- **配置**: `apps/api/internal/order/plan_builder.go`（Unit C 内、Unit D 共有時は横串移動可能）
- **公開 interface**: `PlanBuilder`
- **公開メソッド**: `Build(ctx, history []OrderRecord) (*Plan, error)`
- **implementation**: `BedrockPlanBuilder`
- **依存先**:
  - `BedrockAdapter` (LC-ORDER-05)
  - `RetryClassifier` (LC-ORDER-07)
  - `FallbackSuggestProvider` (LC-ORDER-09)
- **分岐ロジック**:
  - Bedrock 呼出（NFRC-C06/C07: リトライ 1 回 / タイムアウト 1.5s）
  - Bedrock 失敗 + 履歴 ≥ 5 件 → `FallbackSuggestProvider.BuildFromHistory(history)`
  - Bedrock 失敗 + 履歴 < 5 件 → `FallbackSuggestProvider.Default()`
- **由来パターン**: P-PLAN-01

---

### LC-ORDER-09: FallbackSuggestProvider

- **責務**: Bedrock 失敗時の Plan 生成（履歴最頻パターン or Default 5 店舗ランダム選択）。
- **配置**: `apps/api/internal/adapters/fallback/provider.go`（横串、Unit C / D 共有）
- **公開 interface**: `FallbackSuggestProvider`
- **公開メソッド**:
  - `BuildFromHistory(history []OrderRecord) *Plan`（最頻パターン）
  - `Default() *Plan`（5 店舗ランダム選択）
- **5 店舗ラインナップ**（BR-C07、`internal/adapters/fallback/stores.go` の定数配列）:
  | Index | Store | Menu | Amount | Category |
  |---|---|---|---|---|
  | 0 | ゴロゴロ食堂 | おまかせ定食 | ¥1,000 | food |
  | 1 | ぐうたら亭 | 手抜き丼 | ¥800 | food |
  | 2 | ダメ屋 | やる気なしカレー | ¥1,200 | food |
  | 3 | 怠惰キッチン | 何でもよし弁当 | ¥1,500 | food |
  | 4 | ふぬけ食堂 | しょうがない定食 | ¥900 | food |
- **由来**: BR-C06 / BR-C07、unit-of-work.md §3.3

---

### LC-ORDER-10: LogSummary

- **責務**: NFRC-C12 の Unit C 固有 11 項目を `PlaceOrder` リクエスト中に蓄積し、完了時に 1 件のサマリログ（`place_order_complete` event）として出力。
- **配置**: `apps/api/internal/order/logger.go`
- **構造**:
  - 11 項目の状態を保持する struct（`idempotencyKey` / `idempotent` / `orderId` / `bedrockLatencyMs` / `bedrockAttempt` / `fallbackTriggered` / `category` / `amount` / `storeName` / `menuName` / `historyCount` / `source`）
  - 各項目の setter メソッド
  - `LogComplete()` メソッド（`slog.InfoContext` で 11 項目を一括出力）
- **使用パターン**: `defer summary.LogComplete()` で完了時保証
- **依存先**: `slog` (標準ライブラリ)、`ContextAwareSlogHandler` (Unit A LC-AUTH-05)
- **由来パターン**: P-OBS-02
- **PII 防御**: Bedrock プロンプト本文・レスポンス本文のフィールドを意図的に含めない（NFRC-C24 整合）

---

### LC-ORDER-11: EventLogger

- **責務**: Bedrock リトライ発動時 / フォールバック発動時の WARNING ログを `slog.WarnContext` で出力。
- **配置**: `apps/api/internal/order/event_logger.go`
- **公開関数**:
  - `LogBedrockRetry(ctx, attempt int, errorClass string, elapsedMs int64)`
  - `LogFallbackTriggered(ctx, reason string, historyCount int, fallbackType string)`
- **呼出元**: `PlanBuilder` (LC-ORDER-08) 内
- **由来パターン**: P-OBS-03

---

### LC-ORDER-12: LatencyMiddleware

- **責務**: Gin handler chain の最初で全 API リクエストのレイテンシを計測し、`request_complete` ログに `latencyMs` / `statusCode` / `path` を付与。
- **配置**: `apps/api/internal/middleware/latency.go`（横串、全 Unit で利用）
- **公開関数**: `LatencyMiddleware() gin.HandlerFunc`
- **由来パターン**: P-OBS-01
- **NFRC-C13-1 アラーム要件**: 本 middleware で E2E レイテンシが必ず付与されるため、CloudWatch Logs メトリクスフィルタによる p95 検知の信頼性が最大化

---

### LC-ORDER-13: MeasureHelper

- **責務**: 各ステップ（Bedrock / Wallet / Adapter / History / PlanBuilder）のレイテンシを計測する generic 関数。
- **配置**: `apps/api/internal/observability/measure.go`（横串、全 Unit で利用）
- **公開関数**: `Measure[T any](ctx, name string, fn func() (T, error)) (T, time.Duration, error)`
- **由来パターン**: P-OBS-01
- **使用例**: `result, elapsed, err := observability.Measure(ctx, "bedrock", func() (*Plan, error) { return adapter.InferOrderPlan(ctx, history) })`

---

### LC-ORDER-14: BedrockClientInit (package init)

- **責務**: `bedrockruntime.Client` を Lambda INIT フェーズで初期化、package-level 変数 `defaultClient` に保持。
- **配置**: `apps/api/internal/adapters/bedrock/adapter.go` の `init()` 関数
- **挙動**: AWS Config ロード → `bedrockruntime.NewFromConfig(cfg)` → 失敗時は `panic`
- **由来パターン**: P-INIT-01
- **NFRC-C18 達成根拠**: INIT フェーズで burst CPU を活用し SDK 初期化を 250〜400ms で完了、INVOKE フェーズは SDK 再利用で 600ms 以内

---

### LC-ORDER-15: DynamoClientInit (package init)

- **責務**: `dynamodb.Client` を Lambda INIT フェーズで初期化、package-level 変数 `defaultClient` に保持。
- **配置**: `apps/api/internal/repo/order_history/repository.go` の `init()` 関数
- **挙動**: AWS Config ロード → `dynamodb.NewFromConfig(cfg)` → 失敗時は `panic`
- **由来パターン**: P-INIT-01

---

## 3. テスト専用 Backend コンポーネント

### LC-ORDER-16: MockBedrockAdapter

- **責務**: `BedrockAdapter` interface の関数フィールド（closure）注入型 mock。NFRC-C16 4 シナリオ対応。
- **配置**: `apps/api/internal/adapters/bedrock/mock.go`
- **構造**:
  - `InferOrderPlanFunc func(ctx, history) (*Plan, error)` 関数フィールド
  - `Calls int` 呼び出し回数記録
  - `InferOrderPlan` メソッドが `Func` を呼ぶ（未設定時はデフォルト成功応答）
- **由来パターン**: P-MOCK-01
- **Unit D との共有**: Unit D テストでも再利用（横串）

---

### LC-ORDER-17: MockDeliveryAdapter (extended)

- **責務**: 既存の production `MockDeliveryAdapter` を closure 注入対応に拡張、テスト時の挙動制御。
- **配置**: `apps/api/internal/adapters/delivery/mock.go`
- **構造**: `PlaceFunc func(ctx, plan, idempotencyKey) error` + `Calls int`
- **由来パターン**: P-MOCK-01

---

### LC-ORDER-18: MockFallbackProvider

- **責務**: `FallbackSuggestProvider` interface の closure 注入型 mock。
- **配置**: `apps/api/internal/adapters/fallback/mock.go`
- **構造**:
  - `BuildFromHistoryFunc func(history) *Plan`
  - `DefaultFunc func() *Plan`
  - `Calls int`
- **由来パターン**: P-MOCK-01

---

### LC-ORDER-19: WalletStub (test-only)

- **責務**: Unit B `WalletService` interface の in-memory 実装（テスト専用）。同一 idempotencyKey の重複検知、残高管理を再現。
- **配置**: `apps/api/internal/order/wallet_stub_test.go`
- **構造**:
  - `balances map[string]int` (userID → balance)
  - `idempotencyRecords map[string]*WalletDeductResult` (key → result)
  - `Deduct` / `GetBalance` メソッド実装
- **由来パターン**: P-PBT-01
- **PBT 用途**: 状態遷移検証（同一 idempotencyKey の N 回送信で残高 -X が 1 回のみ）

---

### LC-ORDER-20: InmemoryHistory (test-only)

- **責務**: `OrderHistoryRepository` interface の in-memory 実装（テスト専用）。
- **配置**: `apps/api/internal/repo/order_history/inmemory.go`
- **構造**:
  - `records map[string]*OrderRecord` (orderID → record)
  - `Insert` で重複検知、`Query` で降順 / LIMIT 対応
- **由来パターン**: P-PBT-01

---

## 4. Frontend コンポーネント

### LC-ORDER-21: GoroButton

- **責務**: メイン画面の「ご飯めんどくさい」ボタン。`useOrder` hook を使用し、`disabled` 属性で連打抑制 + ローディング表示。
- **配置**: `web/components/GoroButton.tsx`
- **依存先**:
  - `useOrder` (LC-ORDER-22)
- **a11y**: HTML `disabled` 属性、ARIA label、キーボード操作対応
- **由来パターン**: P-FE-LOCK-01、P-FE-ERR-01

---

### LC-ORDER-22: useOrder

- **責務**: `PlaceOrder` mutation の React hook。エラーハンドリング（`mapOrderError`）、連打抑制（`useDisableLock`）、invalidate（`['orderHistory']` / `['balance']`）を統合。
- **配置**: `web/hooks/useOrder.ts`
- **依存先**:
  - TanStack Query `useMutation`
  - `mapOrderError` (LC-ORDER-26)
  - `useDisableLock` (LC-ORDER-25)
  - `useToast` (LC-ORDER-29)
  - `useRouter` (Next.js)
  - `apiClient.placeOrder` (`web/lib/api/orders.ts`)
- **設定**:
  - `retry: 0`（NFRC-C19、FD Q-12=A）
  - `onSuccess`: `triggerLock` + `invalidateQueries(['orderHistory'])` + `invalidateQueries(['balance'])`
  - `onError`: `triggerLock` + `mapOrderError(err)` の戻り値で `router.push` / `showToast` / silent
- **戻り値**: `{ ...mutation, disabled: mutation.isPending || isLocked }`
- **由来パターン**: P-FE-ERR-01、P-FE-LOCK-01

---

### LC-ORDER-23: useOrderHistory

- **責務**: `GetHistory` query の React hook。`placeholderData: keepPreviousData` で再 fetch 中も前回データ表示維持。
- **配置**: `web/hooks/useOrderHistory.ts`
- **設定**:
  - `queryKey: ['orderHistory']`
  - `queryFn: fetchOrderHistory`
  - `placeholderData: keepPreviousData`
  - `staleTime: 60_000`（NFRC-C19）
  - `gcTime: 300_000`（NFRC-C19）
  - `refetchOnWindowFocus: true`（NFRC-C19）
- **由来パターン**: P-FE-LOAD-01

---

### LC-ORDER-24: OrderHistoryList / OrderHistorySkeleton

- **責務**:
  - `OrderHistoryList`: 履歴データ表示（`isFetching` 時は `opacity: 0.5`）
  - `OrderHistorySkeleton`: 初回ロード中のスケルトン表示
- **配置**:
  - `web/components/OrderHistoryList.tsx`
  - `web/components/OrderHistorySkeleton.tsx`
- **依存先**: `useOrderHistory` (LC-ORDER-23)
- **空状態**: 0 件時は「履歴がまだありません」ダメ化文言、サジェストカードは出さない
- **由来パターン**: P-FE-LOAD-01

---

### LC-ORDER-25: useDisableLock

- **責務**: 連打抑制 1 秒の振る舞い hook。`triggerLock()` で lock 開始、`isLocked` で disable 状態を返す。
- **配置**: `web/hooks/useDisableLock.ts`
- **公開 API**: `useDisableLock(durationMs: number) → { isLocked: boolean, triggerLock: () => void }`
- **状態管理**:
  - `lockedUntil: number`（lock 解除時刻 epoch ms）を `useState` で保持
  - `isLocked = Date.now() < lockedUntil` 派生値
  - `useEffect` cleanup で `clearTimeout`、unmount 時にタイマー破棄
- **由来パターン**: P-FE-LOCK-01
- **将来の拡張**: 他箇所（増額ボタン、履歴削除等）でも再利用可能

---

### LC-ORDER-26: OrderErrorMapper (mapOrderError)

- **責務**: HTTP エラー / ネットワークエラーから `OrderErrorAction` 型へのマッピング純関数。
- **配置**: `web/lib/errorMappers.ts`
- **公開関数**: `mapOrderError(err: unknown): OrderErrorAction`
- **戻り値型**:
  ```typescript
  type OrderErrorAction =
    | { type: 'navigate'; path: string; transitionMs?: number }
    | { type: 'toast'; text: string; durationMs?: number }
    | { type: 'silent' };
  ```
- **マッピング**:
  - `ApiError(402)` → `{ type: 'navigate', path: '/budget-empty', transitionMs: 0 }`
  - `ApiError(409)` → `{ type: 'silent' }`（冪等性衝突は実質成功扱い、BR-C39）
  - `ApiError(500)` / `NetworkError` / その他 → `{ type: 'toast', text: getRandomToast(), durationMs: 5000 }`
- **由来パターン**: P-FE-ERR-01
- **依存先**: `getRandomToast` (LC-ORDER-27)

---

### LC-ORDER-27: ToastVariants (getRandomToast)

- **責務**: 自虐トースト 3 種ローテーションの純関数 + readonly 配列定義。
- **配置**: `web/lib/toasts.ts`
- **公開関数**: `getRandomToast(): ToastVariant`
- **公開定数**: `TOAST_VARIANTS: readonly ToastVariant[]`
- **3 種候補**（最終文言は Code Generation で確定）:
  1. 「サーバーがやる気を失いました…もう一度お試しください」
  2. 「システムがふぬけてます。少し待ってあげてください」
  3. 「今日はちょっとダメ化に失敗しました。再挑戦しますか？」
- **由来パターン**: P-FE-TOAST-01

---

### LC-ORDER-28: ToastsAtom

- **責務**: トースト一覧のグローバル状態（Jotai atom）。
- **配置**: `web/state/toastAtoms.ts`
- **公開 atom**: `toastsAtom: atom<ToastItem[]>`
- **`ToastItem` 型**: `{ id: string; text: string; durationMs: number }`
- **由来パターン**: P-FE-TOAST-02

---

### LC-ORDER-29: useToast

- **責務**: トースト表示 API hook。`showToast(text, durationMs)` で `toastsAtom` に追加、`setTimeout` で自動消去。
- **配置**: `web/hooks/useToast.ts`
- **公開 API**: `useToast() → { toasts: ToastItem[]; showToast: (text, durationMs?) => void }`
- **依存先**: `toastsAtom` (LC-ORDER-28)、Jotai `useAtom`
- **由来パターン**: P-FE-TOAST-02

---

### LC-ORDER-30: ToastHost

- **責務**: トースト一覧表示コンポーネント。`app/layout.tsx` の `<body>` 直下に配置、`<Providers>` 配下。最大 3 件表示、4 件目以降はキューイング。
- **配置**: `web/components/ToastHost.tsx`
- **依存先**: `useToast` (LC-ORDER-29)、`Toast` (LC-ORDER-31)
- **a11y**: `role="region"` + `aria-live="polite"`
- **由来パターン**: P-FE-TOAST-02

---

### LC-ORDER-31: Toast

- **責務**: 個別トーストコンポーネント。アニメーション、自動消去、a11y 対応。
- **配置**: `web/components/Toast.tsx`
- **a11y**: `role="status"`
- **由来パターン**: P-FE-TOAST-02

---

### LC-ORDER-32: OrderCompletionScreen

- **責務**: 注文完了画面。`useEffect` + `setTimeout(5000)` でメイン画面に自動遷移（NFRC-C21）。
- **配置**: `web/app/order/[id]/complete/page.tsx`
- **依存先**: `useRouter` (Next.js)、Server Component で完了データ取得（BFF 経由）
- **由来**: FD frontend-components.md、NFRC-C21 / NFRC-C23

---

### LC-ORDER-33: ApiClientOrders (apiClient.placeOrder / apiClient.getOrderHistory)

- **責務**: `/api/orders` エンドポイント呼出のラッパ関数。`Authorization: Bearer <accessToken>` ヘッダ付与。
- **配置**: `web/lib/api/orders.ts`
- **依存先**: Unit A `apiClient` (LC-AUTH-09)
- **公開関数**:
  - `placeOrder(req: PlaceOrderRequest): Promise<PlaceOrderResult>`
  - `fetchOrderHistory(limit?: number): Promise<OrderRecord[]>`
- **由来**: Unit A AccessToken 採用 + BFF パターン

---

### LC-ORDER-34: UlidGenerator

- **責務**: Frontend での ULID 生成（idempotencyKey 用）。
- **配置**: `web/lib/ulid.ts`
- **公開関数**: `generateUlid(): string`
- **依存先**: `ulid` npm パッケージ
- **由来**: FD Q-7=A（Frontend ULID 発行）、frontend-components.md

---

## 5. Unit A 再利用コンポーネント

Unit C で再利用する Unit A の論理コンポーネント:

| Unit A コンポーネント | Unit C での利用 |
|---|---|
| **LC-AUTH-05 ContextAwareSlogHandler** | Unit C 共通 8 項目（`level / timestamp / userId / action / traceId / requestId / email_hash / userAgent`）の自動付与 |
| **LC-AUTH-09 apiClient** | Frontend からの API 呼出（`Authorization: Bearer <accessToken>` 付与）。`web/lib/api/orders.ts` から利用 |
| **LC-AUTH-18 BffProxyRouteHandler** | Next.js Server Route Handler。`/api/*` を API Gateway へ透過プロキシ |
| **AttachUserID middleware** | API Gateway Cognito Authorizer から渡される JWT claims を `gin.Context` に注入（凍結契約 §1.1 の前提） |

## 6. Unit B 連携

Unit C は Unit B の以下のコンポーネントを利用（凍結契約 §3.1）:

| Unit B コンポーネント | Unit C での利用 |
|---|---|
| **WalletService.Deduct** | 残高引き落とし。`(ctx, userID, idempotencyKey, amount) → DeductResult` |
| **wallet.ErrInsufficientFunds** | 402 INSUFFICIENT_FUNDS にマッピング（LC-ORDER-02） |
| **wallet.ErrIdempotencyConflict** | 409 IDEMPOTENCY_CONFLICT にマッピング（BR-C39、LC-ORDER-02） |
| **DeductResult.OrderID** | 冪等命中時の OrderID 復元（BR-C15） |

## 7. Unit D 連携

Unit C は Unit D の以下のコンポーネントを利用（凍結契約）:

| Unit D コンポーネント | Unit C での利用 |
|---|---|
| **SuggestService.ResolveSuggestion** | サジェスト経由注文時の Plan 復元（FD Q-5）、Bedrock 再呼出しなし |

## 8. クロスユニット query key 契約

| Query Key | 所有 Unit | 利用 Unit |
|---|---|---|
| `['balance']` | Unit B | Unit C `useOrder.onSuccess` で invalidate（unit-interfaces.md §9.1） |
| `['orderHistory']` | Unit C | Unit C `useOrder.onSuccess` で invalidate（自所有） |

---

## 9. コンポーネント関係図

```
[Frontend]
                                       ┌──────────────────────┐
                                       │  app/layout.tsx       │
                                       │  ├─ <Providers>       │
                                       │  │  └─ children        │
                                       │  └─ <ToastHost> (30)   │
                                       └──────────────────────┘
                                                  │
       ┌──────────────────────────────────────┼─────────────────────┐
       │                                                │                                     │
   [GoroButton (21)]                  [OrderHistoryList (24)]  [OrderCompletionScreen (32)]
       │                                                │
       │ uses                                       │ uses
       ▼                                                ▼
   [useOrder (22)]  ──invalidate──▶ [useOrderHistory (23)]
       │                                                │
       ├─uses─▶ [useDisableLock (25)]
       ├─uses─▶ [mapOrderError (26)] ──uses──▶ [getRandomToast (27)]
       ├─uses─▶ [useToast (29)] ──uses──▶ [toastsAtom (28)]
       └─uses─▶ apiClient.placeOrder (33) ──HTTP──▶ [BFF Route Handler] (LC-AUTH-18)
                                                                                  │
                                                                                  │ /api/orders
                                                                                  ▼
[Backend]                                                          [API Gateway + Cognito Authorizer]
                                                                                  │
                                                                                  ▼
                                                              [LatencyMiddleware (12)] (Gin)
                                                                                  │
                                                                                  ▼
                                                                  [OrderHandler (02)]
                                                                                  │
                                                                                  ▼
                                                                  [OrderService (01)]
                                                                                  │
                              ┌─────────────────────────────┼────────────────────────┐
                              │                                                │                                              │
                  uses [PlanBuilder (08)]      uses [WalletService (Unit B)]   uses [DeliveryAdapter (06)]
                              │                                                                                            │
                              │ uses                                                                                  │ uses
                              ▼                                                                                            ▼
                  [BedrockAdapter (05)] + [RetryClassifier (07)]   [MockDeliveryAdapter (06)]
                              │                                                                                            │
                              │ uses                                                                                   │ uses
                              ▼                                                                                            ▼
                  [FallbackSuggestProvider (09)]                          (noop, log only)
                              │
                              ▼
                  [OrderHistoryRepository (04)] ──uses──▶ DynamoDB

[全 PlaceOrder で使用]
                  [LogSummary (10)] ──defer LogComplete──▶ slog ──▶ CloudWatch Logs
                  [EventLogger (11)] ──slog.Warn──▶ slog ──▶ CloudWatch Logs
                  [MeasureHelper (13)] ──slog.Info──▶ slog ──▶ CloudWatch Logs

[初期化 (init)]
                  [BedrockClientInit (14)] ──init()──▶ defaultClient
                  [DynamoClientInit (15)] ──init()──▶ defaultClient
```

---

## 10. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| **Infrastructure Design** | DynamoDB `OrderHistory` テーブル定義 / Bedrock IAM Role / API Gateway ルート / Lambda 関数 / CloudWatch Alarms / メトリクスフィルタ / AWS Budgets Terraform |
| **Code Generation** | LC-ORDER-01〜34 の具体実装、Bedrock プロンプトテンプレート、`gopter` テストコード、`MockBedrockAdapter` 実装、自虐トースト最終文言 |
| **Build and Test** | dev 環境 Bedrock 実呼出し検証手順、CI ワークフロー、E2E テストシナリオ |

---

## 11. 文書管理

- **承認**: ユーザ承認待ち（Construction フェーズの per-unit ループ承認ゲート）
- **凍結契約への影響**: なし（公開 interface は LC-ORDER-01 OrderService / LC-ORDER-03 DTO のみで、いずれも凍結契約整合済み）
- **次ステージ**: Infrastructure Design（per-unit、Comprehensive 深度）
