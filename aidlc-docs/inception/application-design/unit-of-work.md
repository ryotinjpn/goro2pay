# Unit of Work — ゴロゴロPay

**Document Version**: 1.0
**Created**: 2026-05-07
**Depth**: Comprehensive
**Prerequisite**: Application Design 承認済み、Units Generation Plan 回答済み

本ドキュメントは、ゴロゴロPay MVP の **Unit of Work**（論理的な開発単位）を正式に定義する。Application Design で示唆した 6 Unit 候補を、Plan ヒアリング結果に従って **5 Unit 構成** として確定する。

---

## 1. Unit of Work 定義と用語

### 1.1 本プロジェクトでの Unit of Work の定義

- **Unit = 論理的なコードモジュールの集合**（1 ビジネスドメイン = 1 Go パッケージ系列）
- **物理 Lambda 分割とは独立**: 本 MVP は単一モノリシック API Lambda（Q-E=B'）のため、すべての Unit のコードが同一 Lambda 内に相乗りする
- **各 Unit が per-unit Construction フェーズの単位** になる（Functional Design / NFR Requirements / NFR Design / Infrastructure Design / Code Generation を Unit 単位で実施）

### 1.2 用語定義

| 用語 | 定義 |
|---|---|
| Unit of Work | 論理的な開発単位（本ドキュメントの主題） |
| Service | ドメインロジックのコア（Unit 内に 1 つ以上存在、例: `WalletService`） |
| Module (Go パッケージ) | ディレクトリ単位の物理配置（例: `internal/wallet/`） |
| 横串コンポーネント | Unit に属さない共有リソース（例: `BedrockAdapter`） |
| per-unit Construction | Construction フェーズの各ステージを Unit 単位で実施する方式 |

---

## 2. Unit 一覧（5 Unit 構成）

Plan Q-A=B の決定に従い、以下の 5 Unit で構成する。ダメ化UX（NFR-DEG-*）は独立 Unit とせず、各 Unit の受入基準に横串として織り込む。

| Unit ID | Unit 名 (識別子) | Unit 名 (日本語) | Construction 深度 | 主要コンポーネント |
|---|---|---|---|---|
| **A** | `auth` | 認証 | Standard | AuthContextService, Cognito User Pool, AuthScreens |
| **B** | `budget` | ダメ予算 | Standard | WalletService, WalletRepository, BudgetSettingsRepository, IdempotencyRepository, BudgetResetLogRepository, SchedulerLambda |
| **C** | `order` | 代行手配コア | **Comprehensive** | OrderService, OrderHandler, OrderHistoryRepository, DeliveryAdapter, MainScreen（ボタン）, OrderCompletionScreen |
| **D** | `suggest` | 学習・先回り | Standard | SuggestService, SuggestHandler, 共有 BedrockAdapter 利用, 共有 FallbackSuggestProvider 利用 |
| **E** | `metrics` | ダメ化メトリクス | Standard | MetricsService, MetricsHandler, BudgetRaiseService, BudgetRaiseHandler, BudgetEmptyScreen, RaiseModal |

**横串コンポーネント**（Unit 非所属、Plan Q-B=A）:
- `BedrockAdapter`（`apps/api/internal/adapters/bedrock/`）— Unit C / D で共有
- `FallbackSuggestProvider`（`apps/api/internal/adapters/fallback/`）— Unit C / D で共有

---

## 3. Unit 詳細

### 3.1 Unit A: `auth`（認証）

#### 責務
- ユーザ認証機構全般（Cognito User Pool の利用）
- API リクエスト時の JWT claims 検証と `userId` の Context 注入
- Presentation 層のログイン / 新規登録画面

#### 含まれるコンポーネント

**Backend (Go)**
- `internal/auth/` パッケージ
  - `AuthContextService`（Gin middleware）
  - `AttachUserID()` / `UserIDFromContext()`
- 横断的に使われる認証ヘルパ群

**Frontend (Next.js)**
- `app/(auth)/login/page.tsx` → `LoginScreen`
- `app/(auth)/signup/page.tsx` → `SignupScreen`
- `app/page.tsx` （未認証時の `Landing`）
- `hooks/useAuth.ts`

**Infrastructure**
- Cognito User Pool
- Cognito App Client
- API Gateway Cognito Authorizer

#### 受入基準（stories.md）
- US-0-01（新規登録）
- US-0-02（ログイン）

#### 依存
- **依存先**: なし（最下位レイヤ）
- **依存元**: Unit B, C, D, E（全 Unit が認証済状態を前提）

#### Construction 深度
**Standard** — Cognito のほぼ標準構成、独自ロジックなし

#### ダメ化UX 織り込み
- 登録フローの総タップ数を 5 回以下に抑制（NFR-DEG-01: 低摩擦オンボーディング）
- 認証後のリダイレクト先は予算未設定なら BudgetSetup、設定済みなら MainScreen（US-0-04 と連携）

---

### 3.2 Unit B: `budget`（ダメ予算）

#### 責務
- 仮想ウォレット（残高）の取得・減算・リセット
- 月間ダメ予算の設定・変更
- 冪等性キーによる二重引き落とし防止
- 月初リセットのスケジュール処理（SchedulerLambda）
- 残高不変条件・履歴整合の保証（NFR-REL-01, NFR-REL-02）

#### 含まれるコンポーネント

**Backend (Go)**
- `internal/wallet/` パッケージ
  - `WalletService`（`GetBalance` / `SetBudget` / `Deduct` / `ResetAll`）
- `internal/repo/wallet_repo/`
- `internal/repo/budget_settings/`
- `internal/repo/idempotency/`
- `internal/repo/budget_reset_log/`
- `internal/handlers/wallet_handler.go` → `WalletHandler`

**Scheduler (Go、別 Lambda)**
- `apps/scheduler/` 配下全体
- `MonthlyResetHandler` → `WalletService.ResetAll` を呼び出し

**Frontend (Next.js)**
- `app/budget/page.tsx` → `BudgetSetupScreen`
- `hooks/useWallet.ts`
- `components/BalanceDisplay.tsx`（残高表示、MainScreen から利用）

**Infrastructure**
- DynamoDB: `GoroPay_Wallet`, `GoroPay_BudgetSettings`, `GoroPay_IdempotencyKeys`, `GoroPay_BudgetResetLog`
- EventBridge Scheduler（cron: 月初 00:00 JST）
- Lambda: `monthlyResetLambda`

#### 受入基準（stories.md）
- US-0-03（ダメ予算設定）
- US-1-02（残高常時可視化）
- US-1-04（残高不足時のダメになれない表示）
- US-1-05（二重引き落とし防止 / 冪等性）
- US-1-06（残高が負にならない）
- US-3-05（月初リセット）

#### 依存
- **依存先**: Unit A（`AuthContextService` の `userId`）
- **依存元**: Unit C, E（WalletService 呼び出し、BudgetSettings 参照）

#### Construction 深度
**Standard** — 金額計算・CAS・スケジュールの標準パターン

#### ダメ化UX 織り込み
- 残高表示はメイン画面で **常時・大きく可視化**（NFR-DEG-03）
- 予算初期値 30,000 円（idea.md のペルソナ設定）
- `BudgetSetupScreen` は 1 画面・1 入力で完結（NFR-DEG-01）

---

### 3.3 Unit C: `order`（代行手配コア）★ 本 MVP の心臓

#### 責務
- 「ご飯めんどくさい」ボタン押下から注文完了までのユースケースオーケストレーション
- Bedrock 推論 → Wallet 減算 → 外部アダプタ手配 → 履歴記録の一連フロー
- Bedrock 失敗時のリトライ・フォールバック制御（Plan Q-G=C）
- サジェスト経由の 1 タップ注文対応

#### 含まれるコンポーネント

**Backend (Go)**
- `internal/order/` パッケージ
  - `OrderService`（`PlaceOrder` / `GetHistory`）
- `internal/handlers/order_handler.go` → `OrderHandler`
- `internal/repo/order_history/`

**横串コンポーネント（Unit C / D で共有）**
- `internal/adapters/delivery/` → `DeliveryAdapter` + `MockDeliveryAdapter`
- `internal/adapters/bedrock/` → `BedrockAdapter` + `ClaudeBedrockAdapter`（共有）
- `internal/adapters/fallback/` → `FallbackSuggestProvider`（共有）

**Frontend (Next.js)**
- `app/page.tsx`（MainScreen の「ご飯めんどくさい」ボタン部分）
- `app/order/[id]/complete/page.tsx` → `OrderCompletionScreen`
- `components/GoroButton.tsx`
- `hooks/useOrder.ts`

**Infrastructure**
- DynamoDB: `GoroPay_OrderHistory`（TTL 90 日）
- Amazon Bedrock: Claude 系モデル（InvokeModel / Converse 権限）

#### 受入基準（stories.md）
- US-1-01（コアユースケース: ご飯めんどくさい押下）
- US-1-03（注文履歴記録）
- US-1-07（完了画面の自動遷移）
- US-2-03（サジェストからの 1 タップ注文、Unit D と共同）
- US-X-01（決定疲れ解放の体験）

#### 依存
- **依存先**:
  - Unit A（認証）
  - Unit B（`WalletService.Deduct`）
  - Unit D（`SuggestService.ResolveSuggestion` — サジェスト経由注文時）
  - 横串: `BedrockAdapter`, `DeliveryAdapter`, `FallbackSuggestProvider`
- **依存元**: なし（最上位ユースケース）

#### Construction 深度
**Comprehensive** — 本 MVP の技術的見せ場

Comprehensive で記述する内容（Construction フェーズで展開）:
- 詳細シーケンス図（Bedrock 成功 / リトライ / フォールバック / 冪等性分岐 の全パターン）
- 状態遷移図（注文の全状態）
- Bedrock プロンプトテンプレート（入出力例付き）
- 境界値テスト具体値
- Property-Based Test プロパティ定義
  - P-1: 残高不変（同一 idempotencyKey の多重送信で残高は 1 回分しか減らない）
  - P-2: 履歴整合（全注文履歴の sum == 初期残高 - 現在残高 + リセット差分）
  - P-3: 冪等性レスポンス一貫性（同一 key の 2 回目以降は初回と同じ payload）
- CloudWatch アラーム条件（体感 3 秒超えの p95 検知等）
- 失敗時の補償トランザクション戦略（本 MVP は未実装、設計だけ残す）

#### ダメ化UX 織り込み
- ボタン押下から完了画面表示まで **3 秒以内**（NFR-PERF-01, NFR-DEG-01）
- 操作タップ数は 1 回のみ（NFR-DEG-01）
- 完了画面は 5 秒後に自動でメイン画面に戻る（US-1-07、NFR-DEG-01）
- Bedrock 失敗時もフォールバックで即時解決体験を維持（NFR-DEG-05）

---

### 3.4 Unit D: `suggest`（学習・先回り）

#### 責務
- アプリ起動時の先回りサジェスト生成
- 過去利用履歴の解析と Bedrock への推論リクエスト
- Bedrock 失敗時のフォールバック（履歴最頻パターン or デフォルト）
- SuggestionID の採番・一時保存（Unit C の 1 タップ注文で復元される）

#### 含まれるコンポーネント

**Backend (Go)**
- `internal/suggest/` パッケージ
  - `SuggestService`（`GetSuggestion` / `ResolveSuggestion`）
- `internal/handlers/suggest_handler.go` → `SuggestHandler`
- 横串 `BedrockAdapter`, `FallbackSuggestProvider`（Unit C と共有）
- 履歴参照のため `internal/repo/order_history/`（Unit C が所有、D は読取参照）

**Frontend (Next.js)**
- `components/SuggestionCard.tsx`（MainScreen 内で表示）
- `hooks/useSuggestion.ts`

**Infrastructure**
- DynamoDB: `GoroPay_OrderHistory`（読取、Unit C と共有）
- 追加テーブル: SuggestionID の一時保存（TTL 30 分）※ Infrastructure Design で確定
- Amazon Bedrock（Claude）

#### 受入基準（stories.md）
- US-2-01（起動時サジェスト表示）
- US-2-02（履歴不足時の非表示）
- US-2-03（1 タップ注文の復元部分、Unit C と共同）
- US-2-04（行動履歴の蓄積と学習）
- US-X-02（自己委譲の体験）

#### 依存
- **依存先**:
  - Unit A（認証）
  - Unit C（OrderHistoryRepository の読取、横串 Adapter 共有）
- **依存元**: Unit C（`SuggestService.ResolveSuggestion`）

#### Construction 深度
**Standard** — Bedrock の推論呼び出しとプロンプト設計が中心、Unit C で詳細化した Bedrock 連携パターンを流用

#### ダメ化UX 織り込み
- サジェストカードはメイン画面上部、視線誘導の主役（NFR-DEG-02）
- 「YES」タップだけで Unit C の注文フローに合流（1 タップで完結、NFR-DEG-01）
- コピーは自虐的・依存促進的に（「そろそろご飯めんどくさいですよね？」、NFR-DEG-05）

---

### 3.5 Unit E: `metrics`（ダメ化メトリクス）

#### 責務
- 今月のダメ化回数・予算消化率の集計
- 残高 0 円時の「今月ダメになれません」体験の提供
- 翌月予算の増額推奨値算出と適用（退化ループの完成点）

#### 含まれるコンポーネント

**Backend (Go)**
- `internal/metrics/` パッケージ
  - `MetricsService`（`GetMetrics`）
- `internal/budget_raise/` パッケージ
  - `BudgetRaiseService`（`ComputeRecommendedBudget` / `Accept`）
- `internal/handlers/metrics_handler.go` → `MetricsHandler`
- `internal/handlers/budget_raise_handler.go` → `BudgetRaiseHandler`
- 読取参照: `internal/repo/wallet_repo/`, `internal/repo/budget_settings/`, `internal/repo/order_history/`

**Frontend (Next.js)**
- `components/MetricsPanel.tsx`（今月のダメ化回数・消化率）
- `app/budget-empty/page.tsx` → `BudgetEmptyScreen`
- `components/RaiseModal.tsx`（増額誘導モーダル）
- `hooks/useMetrics.ts`, `hooks/useBudgetRaise.ts`

**Infrastructure**
- 既存テーブル参照のみ（新規テーブルなし）

#### 受入基準（stories.md）
- US-3-01（今月のダメ化回数表示）
- US-3-02（予算消化率表示と警告色演出）
- US-3-03（残高 0 時の画面遷移）
- US-3-04（翌月予算の増額誘導）
- US-X-03（退化ループの完成体験）

#### 依存
- **依存先**:
  - Unit A（認証）
  - Unit B（WalletRepository / BudgetSettingsRepository の読取）
  - Unit C（OrderHistoryRepository の読取）
- **依存元**: なし（最上位の集計 Unit）

#### Construction 深度
**Standard** — 集計・表示ロジックが中心、独自の複雑性は低い

#### ダメ化UX 織り込み
- 消化率 80% 超で赤色警告（NFR-DEG-03: 不安の演出）
- 増額推奨値は `min(current * 1.5, 100_000)`（NFR-DEG-04: 退化ループ）
- 増額モーダルの UI コピーは「翌月予算を増額しますか？ 推奨: ¥XX,XXX」（NFR-DEG-05）

---

## 4. コード組織戦略（Greenfield 向け）

### 4.1 リポジトリ構造（Unit の物理配置）

```
goro2pay/
├── apps/
│   ├── api/                              # Unit A + B + C + D + E の API Lambda
│   │   ├── Dockerfile                    # LWA + Go バイナリ
│   │   ├── main.go                       # 依存注入 + Gin 起動
│   │   ├── internal/
│   │   │   ├── auth/                     # Unit A: 認証 middleware
│   │   │   ├── wallet/                   # Unit B: WalletService
│   │   │   ├── order/                    # Unit C: OrderService ★Comprehensive
│   │   │   ├── suggest/                  # Unit D: SuggestService
│   │   │   ├── metrics/                  # Unit E: MetricsService
│   │   │   ├── budget_raise/             # Unit E: BudgetRaiseService
│   │   │   ├── handlers/                 # Gin handlers（全 Unit の入口）
│   │   │   ├── adapters/                 # 横串（Unit 非所属）
│   │   │   │   ├── delivery/             # Unit C 主利用
│   │   │   │   ├── bedrock/              # Unit C / D 共有
│   │   │   │   └── fallback/             # Unit C / D 共有
│   │   │   ├── repo/                     # 各 Unit 所有
│   │   │   │   ├── wallet_repo/          # Unit B
│   │   │   │   ├── budget_settings/      # Unit B
│   │   │   │   ├── idempotency/          # Unit B
│   │   │   │   ├── budget_reset_log/     # Unit B
│   │   │   │   └── order_history/        # Unit C（D/E から読取参照）
│   │   │   └── apperrors/                # 横串 sentinel errors
│   │   └── go.mod
│   └── scheduler/                        # Unit B: SchedulerLambda
│       ├── Dockerfile or zip             # LWA 不要の通常 Lambda
│       ├── main.go
│       └── internal/
│           └── (WalletService は api 側から共有、初期は go workspaces で解決)
│
├── pkg/                                   # 将来の Unit 分離に備え、共有ドメイン型があれば置く
│   └── (本 MVP では空、必要になった時点で抽出)
│
├── web/                                   # 全 Unit の UI（Unit ごとに page/component をグループ）
│   ├── app/
│   │   ├── layout.tsx                    # 横串: Provider 群
│   │   ├── page.tsx                      # Unit C: MainScreen（ボタン部分）
│   │   │                                 # + Unit D: SuggestionCard
│   │   │                                 # + Unit E: MetricsPanel
│   │   ├── (auth)/                       # Unit A
│   │   │   ├── login/page.tsx
│   │   │   └── signup/page.tsx
│   │   ├── budget/page.tsx               # Unit B: BudgetSetupScreen
│   │   ├── budget-empty/page.tsx         # Unit E: BudgetEmptyScreen
│   │   └── order/[id]/complete/page.tsx  # Unit C: OrderCompletionScreen
│   ├── components/                       # 各 Unit のプレゼンテーション部品
│   │   ├── GoroButton.tsx                # Unit C
│   │   ├── BalanceDisplay.tsx            # Unit B
│   │   ├── SuggestionCard.tsx            # Unit D
│   │   ├── MetricsPanel.tsx              # Unit E
│   │   └── RaiseModal.tsx                # Unit E
│   ├── hooks/                            # Unit 別カスタムフック
│   │   ├── useAuth.ts                    # Unit A
│   │   ├── useWallet.ts                  # Unit B
│   │   ├── useOrder.ts                   # Unit C
│   │   ├── useSuggestion.ts              # Unit D
│   │   ├── useMetrics.ts                 # Unit E
│   │   └── useBudgetRaise.ts             # Unit E
│   ├── state/
│   │   └── atoms.ts                      # Jotai atoms（Unit 別にファイル分割も検討）
│   └── lib/
│       ├── api.ts                        # REST client（横串）
│       └── queryClient.ts                # TanStack Query（横串）
│
├── infra/                                 # Terraform IaC
│   ├── modules/
│   │   ├── lambda_api/                   # Unit 横串（API Lambda + ECR）
│   │   ├── lambda_scheduler/             # Unit B（月初リセット）
│   │   ├── cognito/                      # Unit A
│   │   ├── dynamodb/                     # Unit B + C 所有テーブル
│   │   ├── amplify/                      # Unit A/B/C/D/E 共通（PWA 配信）
│   │   ├── api_gateway/                  # Unit 横串
│   │   └── bedrock/                      # Unit C + D の IAM
│   └── envs/
│       ├── dev/
│       └── prd/                          # 将来用
│
└── aidlc-docs/                            # 設計・計画ドキュメント（コード対象外）
```

### 4.2 Unit と物理配置の対応ルール

- **1 Unit = 1 Go パッケージ系列**（`internal/<unit>/`）
- **Handler は別ディレクトリに集約**（Gin ルーティング側の都合）
- **Repository は Unit 所有のものだけ自分の `internal/repo/<name>/` に**（`order_history` は Unit C 所有、D/E は読取参照）
- **横串 Adapter は Unit 非所属**（`internal/adapters/<name>/`）
- **Frontend は Unit と物理ディレクトリが 1:1 ではない**（React の画面コンポジションの都合で混在）。ただし hooks は Unit 別に 1:1 対応

### 4.3 Go workspaces 運用

`apps/api` と `apps/scheduler` で `WalletService` を共有するため:
- 初期は **Go workspaces** (`go.work`) を使い、2 モジュール間で共有
- 共有コードが増えた時点で `pkg/` にドメイン型を抽出

---

## 5. Unit 間通信パターン

Plan Q-D=A の決定に従い、**Unit 間は Go interface 経由の同期関数呼び出し** とする。

### 5.1 通信パターンの具体例

```go
// Unit C の OrderService は Unit B の WalletService を interface 越しに呼ぶ
package order

type WalletService interface {
    Deduct(ctx context.Context, userID string, amount int, idempotencyKey string) (*DeductResult, error)
}

type OrderService struct {
    wallet WalletService  // interface 経由、実装は main.go で注入
    // ...
}
```

### 5.2 依存方向のルール

- **上位 Unit が下位 Unit の interface を定義**（下位 Unit は他 Unit の型を知らない）
- **interface 名は利用者側（上位 Unit）の概念で命名**（例: Unit C から見ると `WalletService`）
- **実装は下位 Unit が提供し、`main.go` で配線**

この方式は**ヘキサゴナル風のポート & アダプタ**構造を Unit 間に適用したもので、将来の Lambda 分離時には各 interface を gRPC / REST クライアントに置き換えれば済む。

### 5.3 将来の Unit 分離への布石

本 MVP は単一 Lambda だが、将来以下のような分離が可能:
- **Unit ごと Lambda 分離**: interface の背後を HTTP クライアントに差し替え
- **Unit を独立マイクロサービス化**: API Gateway でルート分離

---

## 6. per-unit Construction の実施順序と深度

### 6.1 実施順序（Plan Q-C=A）

```
Unit A (auth)     → Construction 実施（Standard）
  ↓
Unit B (budget)   → Construction 実施（Standard）
  ↓
Unit C (order)    → Construction 実施（Comprehensive）★本 MVP の山場
  ↓
Unit D (suggest)  → Construction 実施（Standard）
  ↓
Unit E (metrics)  → Construction 実施（Standard）
  ↓
Build and Test（全 Unit 統合）
```

### 6.2 各 Unit の Construction ステージと期待成果物

| ステージ | Unit A | Unit B | Unit C | Unit D | Unit E |
|---|---|---|---|---|---|
| Functional Design | Standard | Standard | **Comprehensive** | Standard | Standard |
| NFR Requirements | Standard | Standard | **Comprehensive** | Standard | Standard |
| NFR Design | Standard | Standard | **Comprehensive** | Standard | Standard |
| Infrastructure Design | Standard | Standard | **Comprehensive** | Standard | Standard |
| Code Generation | （ALWAYS） | （ALWAYS） | （ALWAYS） | （ALWAYS） | （ALWAYS） |

**総ドキュメント数の見込み**: 5 Unit × 4 ステージ + 5 Unit × Code Generation Plan = **約 25 ドキュメント**  
Comprehensive 深度の Unit C を含めた想定ページ数は 140〜300 ページ規模（Plan Q-E 議論の §B 案と整合）。

### 6.3 締切（2026-05-10）との関係

- Inception フェーズの Units Generation 完了で **締切遵守**（本 PR で達成）
- Construction は締切外、Phase B として任意実施（execution-plan.md §5）
- Unit C を最優先（デモ映え & 技術的見せ場）とし、ハッカソン本番に向けて Unit A → B → C まで進められれば MVP デモ可能

---

## 7. 審査観点へのトレーサビリティ

| 審査観点 | 対応 |
|---|---|
| ビジネス意図の明確さ | 各 Unit の「ダメ化UX 織り込み」項でペルソナのフェーズ体験と紐付け |
| 創造性とテーマ適合性 | Unit C / D / E がダメ化フェーズ 1 / 2 / 3 にそれぞれ対応、ストーリーラインが Unit 構成に貫通 |
| Unit 分解の適切さ | 本ドキュメントそのもの。5 Unit の境界・責務・依存が明確、共有コンポーネントの扱いも定義 |
| ドキュメント品質 | Unit ごとの詳細セクション、コード組織戦略、通信パターン、Construction 計画の 4 視点で立体化 |
