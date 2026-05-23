# Unit Interfaces — ゴロゴロPay (Construction Wave 1 並列化用 Interface 契約集)

**Document Version**: 1.0
**Created**: 2026-05-21
**Purpose**: Construction フェーズの Wave 1 (各 Unit の設計ドキュメント 20 件並列実行) を可能にするための **Unit 間 Interface 契約の凍結版**
**Source**: `aidlc-docs/inception/application-design/{services.md, component-methods.md, components.md, component-dependency.md, unit-of-work.md}`

---

## 0. 本ドキュメントの位置付け

### 目的
- 5 Unit (auth / budget / order / suggest / metrics) と横串コンポーネント (BedrockAdapter / DeliveryAdapter / FallbackSuggestProvider) の **Interface 契約** を 1 ドキュメントに集約
- Construction Wave 1 で各 Unit の Functional Design / NFR Requirements / NFR Design / Infrastructure Design を **並列実行** できるよう、Unit 境界の整合性を事前確定
- 並列実行中に各 Unit のサブエージェントが他 Unit の内部仕様を仮定しなくて済むようにする

### 凍結スコープ
本ドキュメントで凍結するのは以下のみ:
1. **Go interface 定義** (公開メソッド・引数・戻り値・エラー型)
2. **公開 DTO 構造体** (Unit 境界をまたぐ型のフィールド構成)
3. **REST API ルーティング** (path / HTTP method / status code)
4. **Frontend ↔ Backend 契約** (Request / Response JSON スキーマ)
5. **DynamoDB テーブルキー設計** (PK / SK / GSI / TTL 属性のみ、内部属性は Infra Design で確定)
6. **Sentinel error の存在** (具体メッセージ・コードは Functional Design で確定可能)

### スコープ外 (Wave 1 中に各 Unit が確定する)
- メソッド本体の実装ロジック・擬似コード
- Bedrock プロンプトの本文
- リトライ回数・バックオフ係数の数値
- DynamoDB 内部属性 (Balance 以外の項目等)
- UI コンポーネントの細部 (色・余白・コピー本文)
- アラーム閾値の数値

### 変更ルール
本ドキュメントの **公開 interface・公開 DTO** に変更が必要になった場合、変更を加える前に本ドキュメントを先に更新し、影響を受ける Unit に共有する (各 Unit 内部の変更は自由)。

---

## 1. Unit 一覧と境界

| Unit ID | Identifier | パッケージ Prefix | 公開する主要 Interface |
|---|---|---|---|
| A | `auth` | `internal/auth` | `AuthMiddleware`, `UserIDFromContext` |
| B | `budget` | `internal/wallet`, `internal/repo/{wallet_repo,budget_settings,idempotency,budget_reset_log}` | `WalletService` |
| C | `order` | `internal/order`, `internal/repo/order_history` | `OrderService` |
| D | `suggest` | `internal/suggest` | `SuggestService` |
| E | `metrics` | `internal/metrics`, `internal/budget_raise` | `MetricsService`, `BudgetRaiseService` |
| 横串 | — | `internal/adapters/{bedrock,delivery,fallback}` | `BedrockAdapter`, `DeliveryAdapter`, `FallbackSuggestProvider` |

### 1.1 依存方向 (循環なし)

```
        Unit C (order) ← 最上位ユースケース
          │
          ├──→ Unit B (budget)
          ├──→ Unit D (suggest)
          ├──→ 横串 BedrockAdapter
          ├──→ 横串 DeliveryAdapter
          ├──→ 横串 FallbackSuggestProvider
          └──→ Unit C 所有: OrderHistoryRepository (D / E から読取参照)

        Unit D (suggest)
          ├──→ 横串 BedrockAdapter
          ├──→ 横串 FallbackSuggestProvider
          └──→ Unit C 所有: OrderHistoryRepository (読取のみ)

        Unit E (metrics)
          ├──→ Unit B 所有: WalletRepository, BudgetSettingsRepository (読取)
          └──→ Unit C 所有: OrderHistoryRepository (読取)

        Unit B (budget)
          └──→ (依存なし)

        Unit A (auth)
          └──→ (依存なし)
```

**全 Unit の暗黙の前提**: API リクエスト時に Unit A の `AuthMiddleware` が `userID` を `gin.Context` に注入していること。

---

## 2. Unit A: `auth` — 認証

### 2.1 公開 Go API

```go
package auth

import "github.com/gin-gonic/gin"

// AttachUserID は API Gateway Cognito Authorizer から渡される JWT claims
// (sub) を gin.Context に "userID" キーで注入する Gin middleware を返す
func AttachUserID() gin.HandlerFunc

// UserIDFromContext は middleware で載せた userID を取り出す
// claims が空・不正な場合は ErrUnauthorized を返す
func UserIDFromContext(c *gin.Context) (string, error)

// ErrUnauthorized は認証に失敗した場合の sentinel error
var ErrUnauthorized = errors.New("unauthorized")
```

### 2.2 Frontend ↔ Backend 契約

認証は Cognito Hosted UI または Amplify Auth ライブラリ経由で実施するため、本 MVP のバックエンド API は **JWT 検証以外の認証エンドポイントを持たない**。フロントエンドからの API リクエストには `Authorization: Bearer <id_token>` ヘッダを必須とする (API Gateway Cognito Authorizer が検証)。

### 2.3 Cognito 構成 (公開固定値のみ)

| 項目 | 値 |
|---|---|
| User Pool ID | Infra Design で確定 (env: `COGNITO_USER_POOL_ID`) |
| App Client ID | Infra Design で確定 (env: `COGNITO_APP_CLIENT_ID`) |
| Sign-in attribute | email |
| MFA | OFF (MVP) |
| Password policy | デフォルト (8 文字以上) |

---

## 3. Unit B: `budget` — ダメ予算

### 3.1 公開 Go API

```go
package wallet

type WalletService interface {
    GetBalance(ctx context.Context, userID string) (*WalletSnapshot, error)
    SetBudget(ctx context.Context, userID string, monthlyBudget int) error
    Deduct(ctx context.Context, userID string, amount int, idempotencyKey string) (*DeductResult, error)
    ResetAll(ctx context.Context) (*ResetResult, error)
}

type WalletSnapshot struct {
    UserID        string
    Balance       int       // 円、>= 0
    MonthlyBudget int       // 円、1..100_000
    UpdatedAt     time.Time
}

type DeductResult struct {
    NewBalance int
    Idempotent bool
}

type ResetResult struct {
    ProcessedUsers int
    Errors         []error
}

// Sentinel errors
var (
    ErrInsufficientBalance = errors.New("insufficient balance")
    ErrIdempotencyConflict = errors.New("idempotency key conflict with different payload")
    ErrBudgetOutOfRange    = errors.New("monthly budget out of range")
)
```

### 3.2 Repository 公開 Interface (Unit E の読取参照のみ公開)

```go
package wallet_repo

type WalletReader interface {
    Get(ctx context.Context, userID string) (*WalletRecord, error)
}

type WalletRecord struct {
    UserID    string
    Balance   int
    UpdatedAt time.Time
}
```

```go
package budget_settings

type BudgetSettingsReader interface {
    Get(ctx context.Context, userID string) (*BudgetSettings, error)
}

type BudgetSettings struct {
    UserID        string
    MonthlyBudget int
    EffectiveFrom time.Time
}
```

> ※ Write 系メソッドは Unit B 内部で Service 経由で呼ばれるのみ。他 Unit からは公開しない。

### 3.3 REST API

| Method | Path | 認証 | 概要 |
|---|---|---|---|
| GET | `/api/wallet` | 必須 | 残高 + 月間予算取得 |
| POST | `/api/wallet/budget` | 必須 | 月間予算設定 |

#### `GET /api/wallet` レスポンス
```json
{
  "balance": 28800,
  "monthlyBudget": 30000,
  "updatedAt": "2026-05-21T10:23:45+09:00"
}
```

#### `POST /api/wallet/budget` リクエスト / レスポンス
```json
// Request
{ "monthlyBudget": 30000 }

// Response (200)
{ "monthlyBudget": 30000, "appliedFrom": "2026-06-01T00:00:00+09:00" }
```

### 3.4 DynamoDB テーブル (Unit B 所有、キー設計のみ凍結)

| テーブル | PK | SK | TTL 属性 | 主要属性 |
|---|---|---|---|---|
| `GoroPay_Wallet` | `userId` | (なし) | — | `balance`, `updatedAt` |
| `GoroPay_BudgetSettings` | `userId` | (なし) | — | `monthlyBudget`, `effectiveFrom`, `raiseHistory` |
| `GoroPay_IdempotencyKeys` | `key` | (なし) | `expiresAt` (24h) | `payload`, `createdAt` |
| `GoroPay_BudgetResetLog` | `resetDate` (`YYYY-MM`) | `userId` | — | `prevBalance`, `newBalance`, `at` |

### 3.5 Scheduler Lambda 連携

`apps/scheduler/main.go` の `HandleMonthlyReset(ctx, event)` が `WalletService.ResetAll(ctx)` を呼ぶ。EventBridge Scheduler の cron `cron(0 15 L * ? *)` (UTC) = 月末 15:00 UTC = 月初 00:00 JST で起動。

---

## 4. Unit C: `order` — 代行手配コア (Comprehensive)

### 4.1 公開 Go API

```go
package order

type OrderService interface {
    PlaceOrder(ctx context.Context, userID string, req PlaceOrderRequest) (*PlaceOrderResult, error)
    GetHistory(ctx context.Context, userID string, limit int) ([]*OrderRecord, error)
}

type PlaceOrderRequest struct {
    Category       string  // "food" | "errand" | ...
    IdempotencyKey string  // 必須、ULID 推奨
    SuggestionID   *string // Suggest 経由の 1 タップ注文時のみ
}

type PlaceOrderResult struct {
    OrderID          string
    StoreName        string
    MenuName         string
    Amount           int
    RemainingBalance int
    Idempotent       bool
}

type OrderRecord struct {
    OrderID   string
    UserID    string
    Category  string
    StoreName string
    MenuName  string
    Amount    int
    OrderedAt time.Time
}
```

### 4.2 OrderHistoryRepository 公開 Interface (Unit D / E が読取参照)

```go
package order_history

type OrderHistoryReader interface {
    ListRecent(ctx context.Context, userID string, limit int) ([]OrderRecord, error)
    CountThisMonth(ctx context.Context, userID string) (int, error)
    SumThisMonth(ctx context.Context, userID string) (int, error)
}
```

> ※ `Insert` は Unit C 内部のみ。Unit D / E は読取のみ。

### 4.3 REST API

| Method | Path | 認証 | 概要 |
|---|---|---|---|
| POST | `/api/orders` | 必須 | 注文実施 (コアユースケース) |
| GET | `/api/orders` | 必須 | 注文履歴取得 |

#### `POST /api/orders` リクエスト
```json
{
  "category": "food",
  "idempotencyKey": "01HW3XKQ7EZ8XN1MX1C7Q4Z9TY",
  "suggestionId": null
}
```

#### `POST /api/orders` レスポンス (201)
```json
{
  "orderId": "01HW3XL...",
  "storeName": "CoCo壱番屋",
  "menuName": "ポークカレー",
  "amount": 1200,
  "remainingBalance": 28800,
  "idempotent": false
}
```

#### `POST /api/orders` エラー
| HTTP | code | 意味 |
|---|---|---|
| 400 | `VALIDATION_FAILED` | パラメータ不正 |
| 401 | `UNAUTHORIZED` | JWT 不正 |
| 402 | `INSUFFICIENT_BALANCE` | 残高不足 (BudgetEmptyScreen 誘導) |
| 409 | `IDEMPOTENCY_CONFLICT` | 同一 key で異なる payload |
| 500 | `INTERNAL_ERROR` | 想定外 |

#### `GET /api/orders?limit=20` レスポンス (200)
```json
{
  "items": [
    { "orderId": "...", "category": "food", "storeName": "...", "menuName": "...", "amount": 1200, "orderedAt": "..." }
  ]
}
```

### 4.4 DynamoDB テーブル (Unit C 所有)

| テーブル | PK | SK | GSI | TTL 属性 | 主要属性 |
|---|---|---|---|---|---|
| `GoroPay_OrderHistory` | `userId` | `orderId` (ULID, 時系列ソート可) | `gsi_byCreatedAt` (PK: `userId`, SK: `orderedAt`) | `expiresAt` (90d) | `category`, `storeName`, `menuName`, `amount`, `orderedAt` |

### 4.5 内部依存 (他 Unit / 横串)

```go
type OrderService struct {
    Wallet              WalletService                   // Unit B
    BedrockAdapter      bedrock_adapter.BedrockAdapter  // 横串
    DeliveryAdapter     delivery.DeliveryAdapter        // 横串
    Fallback            fallback.FallbackSuggestProvider // 横串
    SuggestResolver     SuggestResolver                  // Unit D の subset (4.6)
    OrderHistoryWriter  order_history.OrderHistoryWriter // 内部
}

// SuggestResolver は SuggestService の "ResolveSuggestion" のみを必要とする
type SuggestResolver interface {
    ResolveSuggestion(ctx context.Context, suggestionID string) (*suggest.SuggestionPlan, error)
}
```

---

## 5. Unit D: `suggest` — 学習・先回り

### 5.1 公開 Go API

```go
package suggest

type SuggestService interface {
    GetSuggestion(ctx context.Context, userID string) (*Suggestion, error)
    ResolveSuggestion(ctx context.Context, suggestionID string) (*SuggestionPlan, error)
}

type Suggestion struct {
    HasSuggestion bool
    SuggestionID  string
    Title         string
    Plan          *SuggestionPlan
    FallbackUsed  bool // 内部ログ用、API レスポンスには出さない
}

type SuggestionPlan struct {
    StoreName string
    MenuName  string
    Amount    int
    Category  string
}
```

### 5.2 REST API

| Method | Path | 認証 | 概要 |
|---|---|---|---|
| GET | `/api/suggest` | 必須 | 起動時サジェスト取得 |

#### `GET /api/suggest` レスポンス (200, 履歴十分)
```json
{
  "hasSuggestion": true,
  "suggestionId": "01HW3...",
  "title": "そろそろご飯めんどくさいですよね？",
  "plan": {
    "storeName": "CoCo壱番屋",
    "menuName": "ポークカレー",
    "amount": 1200,
    "category": "food"
  }
}
```

#### `GET /api/suggest` レスポンス (200, 履歴不足)
```json
{ "hasSuggestion": false }
```

### 5.3 DynamoDB テーブル (Unit D 所有)

| テーブル | PK | SK | TTL 属性 | 主要属性 |
|---|---|---|---|---|
| `GoroPay_Suggestion` | `suggestionId` | (なし) | `expiresAt` (30 分) | `userId`, `plan` (JSON), `createdAt` |

### 5.4 内部依存

```go
type SuggestService struct {
    OrderHistoryReader order_history.OrderHistoryReader // Unit C の読取
    BedrockAdapter     bedrock_adapter.BedrockAdapter   // 横串
    Fallback           fallback.FallbackSuggestProvider // 横串
    SuggestionStore    SuggestionStore                  // Unit D 内部
}
```

---

## 6. Unit E: `metrics` — ダメ化メトリクス

### 6.1 公開 Go API

```go
package metrics

type MetricsService interface {
    GetMetrics(ctx context.Context, userID string) (*Metrics, error)
}

type Metrics struct {
    DamageCount       int
    ConsumptionRate   float64 // 0.0..1.0
    MonthlyBudget     int
    RemainingBalance  int
    ThresholdExceeded bool    // ConsumptionRate > 0.8
    SummaryText       string
}
```

```go
package budget_raise

type BudgetRaiseService interface {
    ComputeRecommendedBudget(ctx context.Context, userID string) (int, error)
    Accept(ctx context.Context, userID string, newMonthlyBudget int) (*BudgetRaiseResult, error)
}

type BudgetRaiseResult struct {
    NewMonthlyBudget int
    AppliedFrom      time.Time // 翌月 1 日 00:00 JST
}
```

### 6.2 REST API

| Method | Path | 認証 | 概要 |
|---|---|---|---|
| GET | `/api/metrics` | 必須 | 今月のダメ化メトリクス取得 |
| GET | `/api/budget/raise/recommendation` | 必須 | 増額推奨値取得 |
| POST | `/api/budget/raise` | 必須 | 増額適用 (翌月 1 日から) |

#### `GET /api/metrics` レスポンス (200)
```json
{
  "damageCount": 12,
  "consumptionRate": 0.82,
  "monthlyBudget": 30000,
  "remainingBalance": 5400,
  "thresholdExceeded": true,
  "summaryText": "今月のダメ化回数: 12 回、消化額 ¥24,600"
}
```

#### `GET /api/budget/raise/recommendation` レスポンス (200)
```json
{ "currentMonthlyBudget": 30000, "recommendedMonthlyBudget": 45000 }
```

#### `POST /api/budget/raise` リクエスト / レスポンス
```json
// Request
{ "newMonthlyBudget": 45000 }

// Response (200)
{ "newMonthlyBudget": 45000, "appliedFrom": "2026-06-01T00:00:00+09:00" }
```

### 6.3 内部依存 (他 Unit の読取のみ)

```go
type MetricsService struct {
    Wallet             wallet_repo.WalletReader
    BudgetSettings     budget_settings.BudgetSettingsReader
    OrderHistory       order_history.OrderHistoryReader
}

type BudgetRaiseService struct {
    BudgetSettings     budget_settings.BudgetSettingsReader
    BudgetSettingsWriter budget_settings.BudgetSettingsWriter // Unit B が公開する Write Interface
}
```

> Unit B は `BudgetSettingsWriter` を **Unit E 向けに公開** する (本 MVP の例外的な公開):
>
> ```go
> package budget_settings
>
> type BudgetSettingsWriter interface {
>     Set(ctx context.Context, userID string, monthlyBudget int, effectiveFrom time.Time) error
> }
> ```

### 6.4 DynamoDB

新規テーブルなし。既存 (Unit B / C 所有) を読取参照のみ。

---

## 7. 横串 Adapters (Unit 非所属)

### 7.1 BedrockAdapter

```go
package bedrock_adapter

type BedrockAdapter interface {
    InferOrderPlan(ctx context.Context, req InferOrderPlanInput) (*InferOrderPlanOutput, error)
    InferSuggestion(ctx context.Context, req InferSuggestionInput) (*InferSuggestionOutput, error)
}

type InferOrderPlanInput struct {
    UserID   string
    History  []OrderHistoryBrief
    NowJST   time.Time
    Category string
}

type InferOrderPlanOutput struct {
    StoreName string
    MenuName  string
    Amount    int
    Rationale string
}

type InferSuggestionInput struct {
    UserID  string
    History []OrderHistoryBrief
    NowJST  time.Time
}

type InferSuggestionOutput struct {
    HasSuggestion bool
    Title         string
    Plan          *suggest.SuggestionPlan
}

type OrderHistoryBrief struct {
    Category  string
    StoreName string
    MenuName  string
    Amount    int
    OrderedAt time.Time
}

var ErrBedrockUnavailable = errors.New("bedrock unavailable after retry")
```

**実装**: `ClaudeBedrockAdapter` (Converse API 利用、モデル ID は env `BEDROCK_MODEL_ID`)

### 7.2 DeliveryAdapter

```go
package delivery

type DeliveryAdapter interface {
    PlaceOrder(ctx context.Context, req PlaceOrderInput) (*PlaceOrderOutput, error)
}

type PlaceOrderInput struct {
    UserID    string
    Category  string
    StoreName string
    MenuName  string
    Amount    int
}

type PlaceOrderOutput struct {
    ExternalOrderID string
    Status          string // "accepted" 等
    ETAMinutes      int
}
```

**実装**: `MockDeliveryAdapter` (固定応答、本 MVP のみ)

### 7.3 FallbackSuggestProvider

```go
package fallback

type FallbackSuggestProvider interface {
    BuildFromHistory(ctx context.Context, history []bedrock_adapter.OrderHistoryBrief) (*suggest.SuggestionPlan, error)
    Default() *suggest.SuggestionPlan
}
```

---

## 8. グローバルエラー → HTTP マッピング

各 Unit の Handler 層で以下のマッピングを共通実装する (`internal/apperrors/` に集約)。

| Sentinel error (Go) | HTTP | code (JSON) |
|---|---|---|
| `auth.ErrUnauthorized` | 401 | `UNAUTHORIZED` |
| `wallet.ErrInsufficientBalance` | 402 | `INSUFFICIENT_BALANCE` |
| `wallet.ErrIdempotencyConflict` | 409 | `IDEMPOTENCY_CONFLICT` |
| `wallet.ErrBudgetOutOfRange` | 400 | `VALIDATION_FAILED` |
| `bedrock_adapter.ErrBedrockUnavailable` | 200 (フォールバック透過化) | (内部ログのみ) |
| `validator.ValidationErrors` | 400 | `VALIDATION_FAILED` |
| その他 | 500 | `INTERNAL_ERROR` |

### エラーレスポンス JSON
```json
{
  "code": "INSUFFICIENT_BALANCE",
  "message": "今月のダメ予算を使い切りました"
}
```

---

## 9. Frontend Hooks Interface (TypeScript)

各 Unit のフロントエンド実装が公開するフックの型契約。実装本体は各 Unit 内部、契約のみ凍結。

```typescript
// Unit A
function useAuth(): {
  user: { userId: string; email: string } | null;
  login(email: string, password: string): Promise<void>;
  signup(email: string, password: string): Promise<void>;
  logout(): Promise<void>;
  isAuthenticated: boolean;
};

// Unit B
function useWallet(): {
  balance: number;
  monthlyBudget: number;
  isLoading: boolean;
  refetch(): void;
};

// Unit C
function useOrder(): {
  placeOrder(input: { suggestionId?: string }): Promise<{
    orderId: string;
    storeName: string;
    menuName: string;
    amount: number;
    remainingBalance: number;
  }>;
  isPlacing: boolean;
};

// Unit D
function useSuggestion(): {
  suggestion: {
    hasSuggestion: boolean;
    suggestionId?: string;
    title?: string;
    plan?: { storeName: string; menuName: string; amount: number; category: string };
  } | null;
  isLoading: boolean;
};

// Unit E
function useMetrics(): {
  damageCount: number;
  consumptionRate: number;
  monthlyBudget: number;
  remainingBalance: number;
  thresholdExceeded: boolean;
  summaryText: string;
};

function useBudgetRaise(): {
  recommendedBudget: number;
  accept(newBudget: number): Promise<void>;
  dismiss(): void;
};
```

### 9.1 クロスユニット TanStack Query key 契約

Unit C の注文完了後に Unit B の残高表示を即時更新するため、以下の query key 契約を凍結する。

| Query Key | 所有 Unit | 参照 Unit | 用途 |
|---|---|---|---|
| `['balance']` | Unit B (`useWallet`) | Unit C (`useOrder`) | 注文完了時に Unit C が `invalidateQueries({ queryKey: ['balance'] })` を呼ぶ |

**Unit C の実装責務**:
- `useOrder().placeOrder()` が成功したとき、`queryClient.invalidateQueries({ queryKey: ['balance'] })` を呼ぶ
- この invalidate により Unit B の `useWallet` が即時 refetch し、`BalanceDisplay` に最新残高が反映される

**変更ルール**: `['balance']` の query key を変更する場合、Unit B と Unit C の両方を同時に更新すること。

---

## 10. 環境変数 (Unit 横串で参照する設定)

各 Unit の Infrastructure Design で具体値を確定するが、**変数名のみ凍結**。

| 変数名 | 利用 Unit | 用途 |
|---|---|---|
| `AWS_REGION` | 全 | `ap-northeast-1` 固定 |
| `COGNITO_USER_POOL_ID` | A | Cognito 検証 (API Gateway 側で利用) |
| `COGNITO_APP_CLIENT_ID` | A | Frontend 認証 |
| `DDB_TABLE_WALLET` | B, E | DynamoDB テーブル名 |
| `DDB_TABLE_BUDGET_SETTINGS` | B, E | 〃 |
| `DDB_TABLE_IDEMPOTENCY` | B | 〃 |
| `DDB_TABLE_BUDGET_RESET_LOG` | B | 〃 |
| `DDB_TABLE_ORDER_HISTORY` | C, D, E | 〃 |
| `DDB_TABLE_SUGGESTION` | D | 〃 |
| `BEDROCK_MODEL_ID` | C, D | Claude モデル ID (例: `apac.anthropic.claude-sonnet-4-6-20251201-v1:0`) |
| `BEDROCK_REGION` | C, D | Bedrock リージョン |
| `LOG_LEVEL` | 全 | `info` / `debug` |

---

## 11. 並列実行ガイド (Wave 1 サブエージェント向け)

### 11.1 各 Unit のサブエージェントが従うルール
- 本ドキュメントに記載された **公開 interface・公開 DTO** を変更しない
- 変更が必要だと判断したら、本ドキュメント更新の Issue を立てて他 Unit と調整する
- 内部実装 (private 関数、内部構造体、Repository 内部メソッド等) は自由に決めてよい
- 各 Unit のドキュメントから本ドキュメントを `[unit-interfaces.md](../../interfaces/unit-interfaces.md)` で参照する

### 11.2 Wave 1 で並列実行可能な Issue 一覧 (20 件)
- Unit A: #17 #18 #19 #20
- Unit B: #24 #25 #26 #27
- Unit C: #32 #33 #34 #35
- Unit D: #43 #44 #45 #46
- Unit E: #50 #51 #52 #53

### 11.3 並列実行時の出力先 (衝突回避)
各 Unit の成果物は **異なるディレクトリ** に出力されるため、ファイル衝突は発生しない:
- `aidlc-docs/construction/auth/{functional-design,nfr-requirements,nfr-design,infrastructure-design}/`
- `aidlc-docs/construction/budget/...`
- `aidlc-docs/construction/order/...`
- `aidlc-docs/construction/suggest/...`
- `aidlc-docs/construction/metrics/...`

---

## 12. 共有 Frontend インフラ (Unit A 実装 → 全 Unit 再利用)

Unit A が実装する以下のコンポーネントは **Unit B〜E 全て** で共有する。各 Unit は独自の HTTP クライアントや BFF Route Handler を作成しない。

| LC (Unit A) | 共有する Unit | 用途 |
|---|---|---|
| **LC-AUTH-09** `apiClient` (`web/lib/apiClient.ts`) | B, C, D, E | `/api/*` への authenticated fetch（`Authorization: Bearer <accessToken>` 自動付与、BFF 経由） |
| **LC-AUTH-18** `BffProxyRouteHandler` (`web/app/api/[...path]/route.ts`) | B, C, D, E | catch-all proxy — `/api/*` を API Gateway に転送。各 Unit は専用 Route Handler ファイルを新設しない |

**変更ルール**: `apiClient` の公開 IF（`request(input)` シグネチャ）や `BffProxyRouteHandler` を変更する場合、影響を受ける全 Unit（B〜E）に共有すること。

---

## 13. 変更履歴

| Date | Version | 変更内容 | 影響 Unit |
|---|---|---|---|
| 2026-05-21 | 1.0 | 初版凍結 (Wave 1 開始前) | 全 Unit |
