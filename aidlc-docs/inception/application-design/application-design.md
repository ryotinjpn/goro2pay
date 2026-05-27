# Application Design — ゴロゴロPay（集約ドキュメント）

**Document Version**: 1.0
**Created**: 2026-05-07
**Depth**: Comprehensive
**Persona**: 佐藤陽介（[personas.md](../user-stories/personas.md) / requirements.md §2.2.1）
**Related Docs**: [components.md](./components.md), [component-methods.md](./component-methods.md), [services.md](./services.md), [component-dependency.md](./component-dependency.md)

本ドキュメントは、ゴロゴロPay MVP の **Application Design** ステージ全成果物を集約する。詳細は各分冊に委ね、ここでは技術選択・全体構造・Intent との対応を俯瞰的に記述する。

---

## 1. Executive Summary

「ゴロゴロPay」は、AWS サーバレスをバックエンドとし、Next.js (App Router) on Amplify Hosting をフロントエンドとするモバイル向け PWA である。「人をダメにする」Intent を最優先とし、**ダメ化UX** という独自 NFR のもとに、以下を実装する：

- ボタン 1 タップで代行手配が完了する即時解決体験（US-1-01 コア）
- 行動学習と起動時サジェストによる思考削減（US-2-01）
- ダメ化メトリクスと増額誘導による退化ループ（US-3-01〜US-3-04）

技術スタックは **Go + Gin + Lambda Web Adapter（モノリシック API Lambda）**、**DynamoDB 5 テーブル**、**Amazon Bedrock (Claude 系)**、**Amazon Cognito**、**EventBridge Scheduler**、**Terraform IaC** の組み合わせ。

---

## 2. 技術スタック確定事項

| レイヤ | 決定事項 | 根拠（Application Design Plan Answer） |
|---|---|---|
| Backend Runtime | Go + Gin + Lambda Web Adapter（コンテナイメージ） | Q-A=C |
| Frontend Framework | Next.js (App Router) フル機能 | Q-B=B |
| Frontend Hosting | AWS Amplify Hosting（Terraform `aws_amplify_app` 管理） | Q-B=B |
| API Role Split | API は全て Lambda + Gin、Next.js は UI 専用 | Q-B-ext=A |
| Frontend State | Jotai + TanStack Query | Q-C=X (Jotai) |
| API Protocol | REST (JSON over HTTPS) | Q-D=A |
| Lambda Granularity | 単一モノリシック API Lambda + Scheduler Lambda（計 2 物理 Lambda） | Q-E=B' |
| Adapter DI | Go interface + struct + 手動注入（DI フレームワーク不使用） | Q-F=A |
| Bedrock Error Handling | 指数バックオフ 1 回リトライ → 固定フォールバック | Q-G=C |
| Auth Protection | `/health` 以外の全 API に Cognito Authorizer | Q-H=A |
| IaC | Terraform（Requirements Q-12=C） | Requirements 準拠 |
| Region | `ap-northeast-1` | Requirements §5.5 |
| Security Extension | 強制しない（Q-16=B）ただし AWS デフォルト遵守 | Requirements §4.5 |
| PBT Extension | 純粋関数 + シリアライゼーションに適用（Q-17=B, gopter/rapid） | Requirements §4.9 |

---

## 3. 階層アーキテクチャ（要約）

```
┌──────────────────────────────────────────────────────────┐
│ Presentation Layer  (Next.js on Amplify Hosting)          │
│  Screens + Jotai Atoms + TanStack Query Hooks             │
└──────────────────────────────────────────────────────────┘
                         │ REST JSON
                         ▼
┌──────────────────────────────────────────────────────────┐
│ API Gateway (REST) + Cognito User Pool Authorizer         │
└──────────────────────────────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────┐
│ ApiLambda (Go + Gin + LWA, コンテナイメージ)              │
│   Gin middleware (AuthContextService)                     │
│   ├ Handlers: Health / Wallet / Order / Suggest /         │
│   │           Metrics / BudgetRaise                       │
│   ├ Services: WalletService / OrderService /              │
│   │           SuggestService / MetricsService /           │
│   │           BudgetRaiseService                          │
│   ├ Adapters: DeliveryAdapter(Mock) / BedrockAdapter /    │
│   │           FallbackSuggestProvider                     │
│   └ Repositories: Wallet / BudgetSettings /               │
│                   OrderHistory / Idempotency /            │
│                   BudgetResetLog                          │
└──────────────────────────────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────┐
│ SchedulerLambda (Go, Cron 月初 00:00 JST)                 │
│   MonthlyResetHandler → WalletService.ResetAll            │
└──────────────────────────────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────┐
│ Infrastructure Layer                                       │
│  Cognito User Pool                                        │
│  DynamoDB: Wallet, BudgetSettings, OrderHistory,          │
│            IdempotencyKeys, BudgetResetLog                │
│  Amazon Bedrock (Claude) — Converse API                   │
│  EventBridge Scheduler — cron trigger                     │
│  CloudWatch Logs — 構造化ログ                             │
└──────────────────────────────────────────────────────────┘
```

詳細階層図・Mermaid 依存関係図は [component-dependency.md](./component-dependency.md) §1 を参照。

---

## 4. コンポーネント一覧

### 4.1 Presentation Layer（Next.js / Client Components）

| コンポーネント | 主な責務 | 利用ストーリー |
|---|---|---|
| `PresentationApp` | PWA ルート、ルーティング、Provider | 全ストーリー |
| `AuthScreens` | ランディング / 新規登録 / ログイン | US-0-01, US-0-02 |
| `BudgetSetupScreen` | ダメ予算設定 | US-0-03 |
| `MainScreen` | 残高・メトリクス・ボタン・サジェストカード | US-0-04, US-1-01, US-1-02, US-2-01, US-2-03, US-3-01, US-3-02 |
| `OrderCompletionScreen` | 注文完了表示・自動遷移 | US-1-01, US-1-07 |
| `BudgetEmptyScreen` | 「ダメになれません」画面 + 増額誘導 | US-3-03, US-3-04 |
| State atoms + Query hooks | Jotai atoms と TanStack Query | 横串 |

詳細: [components.md §2](./components.md)

### 4.2 Application Layer（Go + Gin）

| コンポーネント | 責務 | Unit |
|---|---|---|
| ApiLambda (goroPayApi) | 全 REST API のハンドラ・ルーティング | 全 Unit に関わる |
| SchedulerLambda (monthlyResetLambda) | 月初リセット | Unit B |
| Gin Middleware (AuthContextService) | JWT claims → userId 注入 | Unit A |
| 6 Handlers | HTTP リクエスト受理とサービス呼び出し | C/D/E に対応 |
| 5 Services | ビジネスロジック（下記 §5） | B/C/D/E |
| 3 Adapters | 外部連携のポート（Mock/Bedrock/Fallback） | C/D |
| 5 Repositories | DynamoDB アクセス層 | B/C |

---

## 5. サービス定義（要約）

| サービス | 責務 | 対応ストーリー |
|---|---|---|
| WalletService | 残高取得・条件付き減算・予算設定・月初リセット | US-0-03, US-1-02, US-1-04, US-1-05, US-1-06, US-3-05 |
| OrderService | 代行手配ユースケースのオーケストレータ（本 MVP のコア） | US-1-01, US-1-03, US-1-07, US-2-03 |
| SuggestService | 起動時サジェストの生成・保存 | US-2-01, US-2-02, US-2-03 |
| MetricsService | ダメ化回数・消化率の集計 | US-3-01, US-3-02 |
| BudgetRaiseService | 増額推奨値算出・翌月予算適用 | US-3-04 |
| AuthContextService (middleware) | JWT claims から userId を Context 注入 | 全認証済 API |

詳細: [services.md](./services.md)

### 5.1 代表シーケンス（再掲）

- **コア**: 「ご飯めんどくさい」押下 → OrderService → BedrockAdapter → WalletService（冪等） → DeliveryAdapter(Mock) → OrderHistoryRepository → 完了応答
- **Bedrock Fallback**: Bedrock 失敗 → 指数バックオフリトライ 1 回 → FallbackSuggestProvider の固定プランで継続
- **サジェスト**: 起動時 → OrderHistoryRepository → BedrockAdapter.InferSuggestion → 一時保存
- **1 タップ注文**: サジェストタップ → OrderService.PlaceOrder(suggestionId) → SuggestService.ResolveSuggestion → Bedrock スキップ → 通常フロー
- **月初リセット**: EventBridge Scheduler → SchedulerLambda → WalletService.ResetAll → 全ユーザ WalletRepository.ResetTo + BudgetResetLogRepository.Insert

詳細シーケンス図: [services.md §4](./services.md)

---

## 6. データモデル（DynamoDB 論理設計）

| テーブル | PK | SK | 主要属性 | TTL |
|---|---|---|---|---|
| `GoroPay_Wallet` | `userId` | - | `balance`, `updatedAt` | - |
| `GoroPay_BudgetSettings` | `userId` | - | `monthlyBudget`, `effectiveFrom`, `raiseHistory` | - |
| `GoroPay_OrderHistory` | `userId` | `orderedAt#orderId` | `category`, `storeName`, `menuName`, `amount`, `orderedAt` | 90 日 |
| `GoroPay_IdempotencyKeys` | `idempotencyKey` | - | `userId`, `requestHash`, `resultPayload` | 24 時間 |
| `GoroPay_BudgetResetLog` | `resetDate` (YYYY-MM) | `userId` | `prevBalance`, `newBalance`, `at` | 12 ヶ月 |

物理設計（キャパシティモード・GSI 等）は Construction フェーズの Infrastructure Design（per-unit）で詳細化する。

---

## 7. API エンドポイント一覧（リファレンス）

| メソッド | パス | ハンドラ | 認証 | 主な入出力 |
|---|---|---|---|---|
| GET | `/health` | HealthHandler.Check | 不要 | `200 {status:"ok"}` |
| GET | `/wallet` | WalletHandler.GetBalance | 必須 | → `{balance, monthlyBudget}` |
| POST | `/wallet/budget` | WalletHandler.SetBudget | 必須 | `{monthlyBudget}` → `{monthlyBudget}` |
| POST | `/orders` | OrderHandler.PlaceOrder | 必須 | `{category, idempotencyKey, suggestionId?}` → `{orderId, storeName, menuName, amount, remainingBalance}` |
| GET | `/orders` | OrderHandler.GetHistory | 必須 | `?limit=20` → `{items[]}` |
| GET | `/metrics` | MetricsHandler.GetMetrics | 必須 | → `{damageCount, consumptionRate, monthlyBudget, remainingBalance, thresholdExceeded}` |
| GET | `/suggest` | SuggestHandler.GetSuggestion | 必須 | → `{hasSuggestion, suggestionId?, title?, plan?}` |
| POST | `/budget/raise` | BudgetRaiseHandler.Accept | 必須 | `{newMonthlyBudget}` → `{newMonthlyBudget, appliedFrom}` |

エラー応答体系は [component-methods.md §7](./component-methods.md) を参照。

---

## 8. Go パッケージ構成（案）

```
apps/api/                               # ApiLambda (Gin + LWA, コンテナ)
  Dockerfile                            # LWA を含むコンテナ定義
  main.go                               # 依存注入と Gin の起動
  internal/
    auth/                               # AuthContextService middleware
    wallet/                             # WalletService
    order/                              # OrderService
    suggest/                            # SuggestService
    metrics/                            # MetricsService
    budget_raise/                       # BudgetRaiseService
    handlers/                           # Gin handlers
    adapters/
      delivery/                         # DeliveryAdapter + MockDeliveryAdapter
      bedrock/                          # BedrockAdapter + ClaudeBedrockAdapter
      fallback/                         # FallbackSuggestProvider
    repo/
      wallet_repo/
      budget_settings/
      order_history/
      idempotency/
      budget_reset_log/
    apperrors/                          # sentinel errors

apps/scheduler/                          # SchedulerLambda (月初リセット)
  main.go
  internal/
    (wallet repo を api から共有するため後で検討。初期は api 側と shared-module で抽出)

pkg/                                     # 共有ドメイン型 (必要になれば抽出)

web/                                     # Next.js App Router (Amplify Hosting デプロイ対象)
  app/
    layout.tsx
    page.tsx                             # MainScreen
    (auth)/
      login/page.tsx
      signup/page.tsx
    budget/page.tsx
    budget-empty/page.tsx
    order/[id]/complete/page.tsx
  hooks/
    useAuth.ts
    useWallet.ts
    useOrder.ts
    useSuggestion.ts
    useMetrics.ts
    useBudgetRaise.ts
  state/
    atoms.ts
  lib/
    api.ts                               # REST client
    queryClient.ts                       # TanStack Query setup

infra/                                   # Terraform
  modules/
    lambda_api/
    lambda_scheduler/
    cognito/
    dynamodb/
    amplify/
    api_gateway/
    bedrock/
  envs/
    dev/
    prd/                                 # 将来用
```

この構成は Construction フェーズの Code Generation で確定。**Terraform モジュール構造**はプロジェクトの `terraform-plugin:terraform-module-design` 規約に従う。

---

## 9. ダメ化UX NFR と Application Design の対応

| NFR (Requirements §4.1) | Application Design での具現化 |
|---|---|
| NFR-DEG-01: アクション数 ≤ 2 | MainScreen の単一ボタン設計、OrderHandler の 1 リクエスト完結 |
| NFR-DEG-02: 起動時サジェスト | SuggestService + GET /suggest（App 初期表示で必ず呼ぶ） |
| NFR-DEG-03: 常時可視化 | MainScreen の残高 + メトリクス常駐表示 |
| NFR-DEG-04: 増額誘導 | BudgetRaiseService の推奨値算出 + BudgetEmptyScreen の増額誘導 |
| NFR-DEG-05: 自虐的 UI コピー | フロント側の文言は UX コピー集として Construction で整備 |

---

## 10. Unit 分解への示唆（Units Generation ステージ入力）

本ドキュメント群は、Units Generation ステージでの 6 Unit 分解の入力となる。

| Unit 候補 | 境界内の主要コンポーネント | 備考 |
|---|---|---|
| Unit A: 認証・ユーザー管理 | AuthScreens, Cognito, AuthContextService | Infrastructure Design で Cognito User Pool 詳細化 |
| Unit B: ダメ予算・仮想ウォレット | BudgetSetupScreen, WalletHandler, WalletService, 各 Repo（Wallet/BudgetSettings/Idempotency/BudgetResetLog）, SchedulerLambda | 他 Unit から呼ばれる基盤 |
| Unit C: 代行手配コア | MainScreen（ボタン）, OrderCompletionScreen, OrderHandler, OrderService, OrderHistoryRepository, DeliveryAdapter, BedrockAdapter | 本 MVP のコア、Unit B/D に依存 |
| Unit D: 行動学習・先回り提案 | SuggestHandler, SuggestService, 共有 BedrockAdapter, 共有 FallbackSuggestProvider | OrderHistoryRepository を参照 |
| Unit E: ダメ化メトリクス | MainScreen（メトリクス部）, BudgetEmptyScreen, MetricsHandler, MetricsService, BudgetRaiseHandler, BudgetRaiseService | 読取中心、Unit B/C を参照 |
| Unit F: ダメ化UX 体験（横串） | なし（NFR-DEG に帰属） | 全 Unit に UX 制約として適用 |

**推奨実装順序**: A（認証前提） → B（残高基盤） → C（コアユースケース） → D（学習・先回り） → E（メトリクス）。F は横串で各 Unit の受入基準に織り込む。

---

## 11. 審査観点への総合対応

| 審査観点 | 本ドキュメント群での対応 |
|---|---|
| ビジネス意図（Intent）の明確さ | §1 / §9 / §10 でダメ化UX と各コンポーネントの対応を明示。Intent は Presentation 層からサービス層まで名前空間レベルで保持 |
| 課題と解決策の整合（Problem-Solution Fit） | requirements.md §2.3 Problem Statement で定義した「意思決定リソースの委任」課題に対し、§5 SuggestService（先回り提案）/ OrderService（1タップ実行）/ MetricsService（消費可視化）が解決手段として一貫して接続 |
| 創造性とテーマ適合性 | §5 で SuggestService / BudgetRaiseService / FallbackSuggestProvider / MetricsService をダメ化UX 専用コンポーネントとして明記 |
| Unit 分解の適切さ | §10 で 6 Unit 境界と依存関係を明示、Units Generation ステージの直接の入力 |
| ドキュメント品質 | components.md / component-methods.md / services.md / component-dependency.md の 4 分冊 + 本集約ドキュメントで立体構造を提供 |

---

## 12. 次ステージへの引き継ぎ

**Units Generation ステージ**で確定する項目:
1. Unit の正式名称・粒度・境界
2. Unit 間の正式な依存グラフ
3. Unit 単位の実装優先順位（推奨: A → B → C → D → E）
4. Unit 単位の Construction フェーズでの depth 設定

Construction フェーズで per-unit に展開する項目:
- Functional Design: ビジネスロジックの詳細（Bedrock プロンプト文面、残高計算の境界値、TTL の具体値 等）
- NFR Requirements: Unit ごとのダメ化UX 要件の数値化
- NFR Design: 冪等性キーの具体実装、条件付き書き込みの CAS パターン等
- Infrastructure Design: Terraform モジュール定義、DynamoDB 物理スキーマ、IAM ポリシー、Cognito User Pool 詳細、Amplify Hosting 設定、EventBridge Scheduler cron 式、パラメータシート、コスト見積
- Code Generation: Go コード + Next.js コード + Terraform コードの生成計画と実装
