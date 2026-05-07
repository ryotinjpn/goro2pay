# Components — ゴロゴロPay

**Document Version**: 1.0
**Created**: 2026-05-07
**Depth**: Comprehensive

本ドキュメントは、Application Design Plan (approved) に基づく **高レベルコンポーネント定義** と責務を記述する。詳細なビジネスロジックは Construction フェーズの Functional Design（per-unit）で展開する。

---

## 1. コンポーネントレイヤ構成

```
┌───────────────────────────────────────────────────────────────┐
│ Presentation Layer (Next.js / App Router / Amplify Hosting)   │
│   UI Components · Screens · Auth Integration                  │
└───────────────────────────────────────────────────────────────┘
                                │ REST (JSON over HTTPS)
                                ▼
┌───────────────────────────────────────────────────────────────┐
│ API Gateway (REST) + Cognito Authorizer                        │
└───────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌───────────────────────────────────────────────────────────────┐
│ Application Layer (Lambda × 2 / Go + Gin + LWA)               │
│   API Lambda (Gin Router)        │   Scheduler Lambda          │
│   ├ Handlers                     │   └ MonthlyResetHandler     │
│   ├ Services                     │                             │
│   └ Adapters                     │                             │
└───────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌───────────────────────────────────────────────────────────────┐
│ Infrastructure Layer                                           │
│   DynamoDB · Amazon Bedrock · Amazon Cognito · EventBridge    │
│   Scheduler · CloudWatch Logs                                  │
└───────────────────────────────────────────────────────────────┘
```

---

## 2. Presentation Layer（Next.js / Client Components）

### 2.1 PresentationApp
- **目的**: ゴロゴロPay の PWA クライアント全体のルート
- **実装**: Next.js (App Router) フル機能 on Amplify Hosting
- **責務**:
  - 画面ルーティング（ランディング / ログイン / メイン / 予算設定 / 完了 等）
  - Cognito 認証状態の保持と JWT 付与
  - Jotai の Provider とルート atom の初期化
  - TanStack Query の QueryClientProvider 設定
  - PWA manifest / Service Worker 登録
- **利用ストーリー**: 全ストーリーの UI エントリポイント

### 2.2 AuthScreens
- **目的**: ランディング / 新規登録 / ログイン画面
- **実装**: App Router の `app/(auth)/*` セグメント、Client Component
- **責務**: Cognito User Pool との連携（amplify-js または amazon-cognito-identity-js）
- **利用ストーリー**: US-0-01, US-0-02

### 2.3 BudgetSetupScreen
- **目的**: 初回登録後のダメ予算設定画面
- **実装**: App Router の `app/budget/page.tsx`
- **責務**: 予算額の入力・バリデーション・API への送信
- **利用ストーリー**: US-0-03

### 2.4 MainScreen
- **目的**: ゴロゴロPay の中核画面。残高・ダメ化回数・消化率・大きな「ご飯めんどくさい」ボタン・先回りサジェストカードを表示
- **実装**: App Router の `app/page.tsx`
- **責務**:
  - 起動時に SuggestService の API を呼んでサジェストを取得
  - 残高・メトリクスを TanStack Query で取得
  - ボタン押下で OrderService 呼び出し、完了画面への遷移
  - 残高 0 時の増額誘導モーダル表示
- **利用ストーリー**: US-0-04, US-1-01, US-1-02, US-2-01, US-2-03, US-3-01, US-3-02, US-3-03, US-3-04

### 2.5 OrderCompletionScreen
- **目的**: 注文完了表示画面、5 秒後に自動遷移
- **実装**: モーダル or 独立ルート `app/order/[id]/complete`
- **責務**: 注文結果と残予算の表示、タイマーによる自動遷移
- **利用ストーリー**: US-1-01, US-1-07

### 2.6 BudgetEmptyScreen
- **目的**: 残高 0 時の「今月ダメになれません」画面
- **実装**: モーダル or ルート `app/budget-empty/page.tsx`
- **責務**: 今月のダメ化サマリ表示、増額提案モーダル表示
- **利用ストーリー**: US-3-03, US-3-04

### 2.7 State Atoms & Queries（横串）
- **目的**: クライアント状態管理
- **実装**: Jotai atoms + TanStack Query hooks
- **主な atom / query**:
  - `userAtom` (current authenticated user)
  - `walletBalanceAtom` (derived from walletQuery)
  - `walletQuery` (GET /wallet)
  - `metricsQuery` (GET /metrics)
  - `suggestQuery` (GET /suggest、起動時のみ)
  - `orderMutation` (POST /orders)
  - `modalStateAtom` (無力感モーダル等の開閉)

---

## 3. Application Layer（Go + Gin + LWA）

### 3.1 ApiLambda (goroPayApi)
- **目的**: 全 API エンドポイントを束ねる単一 Lambda
- **実装**: Go + Gin + Lambda Web Adapter (コンテナイメージ)
- **ランタイム**: `provided.al2023` + LWA extension + Gin HTTP server on `localhost:8080`
- **責務**:
  - Gin ルーティング
  - JWT claims (userId) の Context 注入（Gin middleware）
  - Handlers → Services → Adapters/Repositories の呼び出し
  - Bedrock 呼び出しのリトライ + フォールバック
  - 構造化ログ出力
- **API ルート**:
  ```
  GET    /health              → 認証不要
  POST   /auth/logout         → ログアウト（JWT 失効、任意）
  GET    /wallet              → 残高取得
  POST   /wallet/budget       → 予算設定・変更
  POST   /orders              → 代行手配（「ご飯めんどくさい」）
  GET    /orders              → 注文履歴
  GET    /metrics             → ダメ化メトリクス取得
  GET    /suggest             → 起動時サジェスト取得
  POST   /budget/raise        → 翌月予算の増額提案を確定
  ```

### 3.2 SchedulerLambda (monthlyResetLambda)
- **目的**: 月初 00:00 JST に全ユーザの残高を予算額にリセット
- **実装**: Go（Gin 不要、シンプルな Lambda handler）
- **トリガ**: EventBridge Scheduler（Cron: `cron(0 15 L * ? *)` UTC = JST 翌月1日 00:00）
- **責務**:
  - 全ユーザの BudgetSettings を走査
  - Wallet テーブルの残高を各ユーザの monthlyBudget にリセット
  - リセット履歴を BudgetResetLog に記録
- **利用ストーリー**: US-3-05

### 3.3 Handlers（Gin ハンドラ群、ApiLambda 内）

| ハンドラ | HTTP メソッド/パス | 責務 | 利用ストーリー |
|---|---|---|---|
| `HealthHandler` | `GET /health` | ヘルスチェック応答 | - |
| `WalletHandler.GetBalance` | `GET /wallet` | 残高取得 | US-1-02 |
| `WalletHandler.SetBudget` | `POST /wallet/budget` | 予算の設定・変更 | US-0-03 |
| `OrderHandler.PlaceOrder` | `POST /orders` | 代行手配（コア） | US-1-01, US-1-04, US-1-05, US-1-06, US-2-03 |
| `OrderHandler.GetHistory` | `GET /orders` | 履歴取得 | US-1-03 |
| `MetricsHandler.GetMetrics` | `GET /metrics` | ダメ化回数・消化率 | US-3-01, US-3-02 |
| `SuggestHandler.GetSuggestion` | `GET /suggest` | 起動時サジェスト | US-2-01, US-2-02 |
| `BudgetRaiseHandler.Accept` | `POST /budget/raise` | 増額提案の確定 | US-3-04 |

### 3.4 Services（ビジネスロジック）

| サービス | 責務 | 詳細 |
|---|---|---|
| `AuthContextService` | JWT claims から `userId` を Context に取り出す | Gin middleware |
| `WalletService` | 残高取得・減算・予算設定・リセット | Idempotency + 条件付き書き込み |
| `OrderService` | 代行手配ユースケースのオーケストレーション | Bedrock 呼び出し + Wallet 減算 + 履歴記録 + Adapter 呼び出し |
| `SuggestService` | 起動時サジェストの生成 | 履歴取得 + Bedrock 推論 + リトライ/フォールバック |
| `MetricsService` | 今月のダメ化回数・消化率計算 | 履歴と予算からの集計 |
| `BudgetRaiseService` | 増額提案値の算出と適用 | 現予算 × 1.5 などのロジック |

### 3.5 Adapters（外部ポートへのゲートウェイ）

| アダプタ | 責務 | 本 MVP 実装 |
|---|---|---|
| `DeliveryAdapter` (interface) | デリバリー注文の抽象化 | `MockDeliveryAdapter` が固定応答（CoCo壱カレー ¥1,200）を返す |
| `BedrockAdapter` (interface) | Bedrock Claude への推論呼び出し | `ClaudeBedrockAdapter` が `bedrock-runtime:Converse` を呼ぶ |
| `FallbackSuggestProvider` | Bedrock 失敗時の固定サジェスト | 履歴最頻カテゴリ + 固定店舗 |

### 3.6 Repositories（DynamoDB アクセス層）

| リポジトリ | 対応テーブル | 責務 |
|---|---|---|
| `WalletRepository` | `GoroPay_Wallet` | 残高の取得・条件付き減算・リセット |
| `BudgetSettingsRepository` | `GoroPay_BudgetSettings` | 予算額の取得・設定 |
| `OrderHistoryRepository` | `GoroPay_OrderHistory` | 履歴の追加・期間検索（TTL 3 ヶ月） |
| `IdempotencyRepository` | `GoroPay_IdempotencyKeys` | 冪等性キー記録・二重送信判定（TTL 短め） |
| `BudgetResetLogRepository` | `GoroPay_BudgetResetLog` | 月初リセットの実行履歴 |

---

## 4. Infrastructure Layer（コンポーネント視点）

### 4.1 DynamoDB テーブル（論理ビュー、Infrastructure Design で詳細化）

| テーブル | PK | SK | TTL | 用途 |
|---|---|---|---|---|
| `GoroPay_Wallet` | `userId` | - | - | 残高 |
| `GoroPay_BudgetSettings` | `userId` | - | - | 月間予算、増額履歴 |
| `GoroPay_OrderHistory` | `userId` | `orderedAt#orderId` | 90 日 | 注文履歴 |
| `GoroPay_IdempotencyKeys` | `idempotencyKey` | - | 24 時間 | 冪等性判定 |
| `GoroPay_BudgetResetLog` | `resetDate` (YYYY-MM) | `userId` | 12 ヶ月 | 月初リセット監査 |

### 4.2 Amazon Bedrock
- Claude 系モデル（モデル ID は Infrastructure Design で確定、Claude 4.7 Sonnet / Haiku を候補とする）
- Converse API を直接呼び出し

### 4.3 Amazon Cognito
- User Pool（メール + パスワード）
- App Client (Next.js から利用)
- API Gateway Cognito Authorizer

### 4.4 EventBridge Scheduler
- 月初リセットのスケジュール
- ターゲット: `monthlyResetLambda`

### 4.5 Amplify Hosting
- Next.js App Router のホスティング
- GitHub リポジトリ連携で自動ビルド・デプロイ
- Terraform の `aws_amplify_app` で管理

---

## 5. コンポーネントと Unit 候補のマッピング

stories.md 末尾で示唆した 6 Unit 候補との対応：

| Unit | 対応コンポーネント（主要） |
|---|---|
| Unit A: 認証・ユーザー管理 | `AuthScreens`, `Cognito User Pool`, `AuthContextService` |
| Unit B: ダメ予算・仮想ウォレット | `BudgetSetupScreen`, `WalletService`, `WalletRepository`, `BudgetSettingsRepository`, `SchedulerLambda` |
| Unit C: 代行手配コア | `MainScreen`, `OrderCompletionScreen`, `OrderHandler`, `OrderService`, `OrderHistoryRepository`, `IdempotencyRepository`, `DeliveryAdapter`, `BedrockAdapter` |
| Unit D: 行動学習・先回り提案 | `SuggestHandler`, `SuggestService`, `OrderHistoryRepository`（参照）, `BedrockAdapter`, `FallbackSuggestProvider` |
| Unit E: ダメ化メトリクス | `MainScreen`（表示）, `BudgetEmptyScreen`, `MetricsHandler`, `MetricsService`, `BudgetRaiseHandler`, `BudgetRaiseService` |
| Unit F: ダメ化UX 体験（横串） | 全コンポーネントに通底する UX 要件、独立コンポーネントはなし（NFR-DEG に帰属） |

---

## 6. 審査観点へのトレーサビリティ

| 審査観点 | 本ドキュメントでの対応 |
|---|---|
| ビジネス意図（Intent）の明確さ | 全コンポーネントの「利用ストーリー」列で US-*-* へ紐付け、Intent がコンポーネントまで連続 |
| 創造性とテーマ適合性 | `MainScreen` / `BudgetEmptyScreen` / `BudgetRaiseHandler` 等、ダメ化UX に直結するコンポーネントを明示 |
| Unit 分解の適切さ | §5 で 6 Unit 候補とコンポーネントのマッピングを整理 |
| ドキュメント品質 | 階層構成図・責務記述・実装方針・トレーサビリティを体系化 |
