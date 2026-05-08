# Application Design Plan — ゴロゴロPay

**Created**: 2026-05-07
**Role**: Software Architect
**Prerequisite**: Requirements Analysis / User Stories / Workflow Planning すべて承認済み
**Depth**: Comprehensive（execution-plan.md §6 に従う）

---

## 0. 目的

Requirements Analysis と User Stories（23 ストーリー、6 Unit 候補）を基に、**コンポーネント・サービス層・コンポーネント依存関係** を高レベル設計として固める。

**スコープ外**:
- ビジネスロジックの詳細（Construction フェーズの Functional Design で per-unit に実施）
- NFR 実装パターン詳細（Construction フェーズの NFR Design）
- 具体的インフラ（Construction フェーズの Infrastructure Design）

---

## 1. 事前確認質問 — 未決定事項のヒアリング

以下の質問にお答えください。各 `[Answer]:` タグに **A / B / C / D / X** の文字で回答してください。

### Question A: Lambda バックエンドの実装言語

どの言語で Lambda ハンドラを書きますか？（Workflow Planning で保留とした項目）

A) **TypeScript / Node.js** — フロントと言語統一、型安全、エコシステム厚い、Bedrock/DynamoDB SDK v3 標準
B) **Python** — AI/ML 文脈と親和性、Bedrock SDK (boto3) シンプル、実装量少なめ
C) **Go** — パフォーマンス重視、コールドスタート速い、型安全、goroutine の恩恵はこのスコープでは限定的
D) **Java / Kotlin** — 金融系での実績、ただしコールドスタート最大
X) Other (please describe after [Answer]: tag below)

**推奨**: **A (TypeScript)**  
理由: (1) PWA フロントと言語統一でコードシェア（型定義・ドメインモデル）が可能、(2) DynamoDB / Bedrock / Cognito SDK v3 が TypeScript ファーストで最も型が充実、(3) ハッカソンの開発スピード観点で学習・実装コストが低い、(4) Property-Based Testing（Q17=B）は fast-check が TypeScript で最も成熟している。

[Answer]: C — Go + Gin + Lambda Web Adapter (LWA) の構成を採用。Lambda はコンテナイメージとしてデプロイし、内部で Gin ベースの HTTP サーバが起動。LWA が API Gateway イベントを localhost:8080 への HTTP リクエストに変換する。1 Lambda に複数ルートを束ねるルーティング主体の構成となる（Q-E の粒度選択に影響）。Property-Based Testing は Go の gopter を使用する方針（元々の Q17=B「fast-check 想定」は Go の gopter / rapid に変更）。

---

### Question B: PWA フロントのフレームワーク

どのフレームワーク / ライブラリで PWA を作りますか？

A) **React + Vite**（SPA、PWA プラグインで manifest/Service Worker 生成）
B) **Next.js (App Router)**（SSR/SSG 対応、AWS Amplify Hosting でフル機能デプロイ可能）
C) **Vue 3 + Vite / Nuxt**（React より記述量少、日本で事例多い）
D) **Svelte / SvelteKit**（バンドル最小、体感速度◎）
X) Other (please describe after [Answer]: tag below)

**推奨**: **A (React + Vite)**  
理由: (1) 本 MVP は SSR/SEO 不要のシンプル SPA、Next.js のランタイムは過剰、(2) Vite のビルド速度は DX 良好、(3) AWS Amplify Hosting での静的配信と相性良い、(4) TypeScript 回答 A と揃えた場合、React エコシステムが最も成熟。

[Answer]: B — Next.js (App Router) を採用。ホスティングは AWS Amplify Hosting に統一する。これにより Next.js のフル機能（Server Components / Server Actions / Route Handlers を含む SSR/ISR）が利用可能。Amplify Gen2 の defineBackend 機能は使わず、Amplify は Hosting 機能のみ利用。Amplify App 自体は Terraform の `aws_amplify_app` リソースで管理する。

**Question B-ext: Next.js と Lambda + Gin の役割分担**

A) パターン 1: API はすべて Lambda + Gin に集約、Next.js は画面・UI 専用（Route Handlers 未使用）
B) パターン 2: 参照系 API は Next.js Route Handlers、トランザクション系は Lambda + Gin
C) パターン 3: すべて Next.js Route Handlers
X) Other

[Answer (B-ext)]: A — すべての API エンドポイントを Lambda + Gin (Go) に集約する。Next.js は UI レンダリングとクライアントサイドロジック（Cognito 連携、TanStack Query、画面遷移等）に専念し、Route Handlers は未使用。Question A (Go + Gin + LWA) の意図を最大限活かす。

---

### Question C: フロントエンドの状態管理

クライアント状態管理の方針は？

A) **React 標準のみ**（useState / useReducer / useContext で十分、追加ライブラリなし）
B) **Zustand**（軽量、ボイラープレート少、人気急上昇）
C) **TanStack Query + 最小 Context**（サーバ状態は Query、UI 状態は Context）
D) **Redux Toolkit**（実績豊富、ボイラープレート多め）
X) Other

**推奨**: **C (TanStack Query + Context)**  
理由: 本 MVP は Bedrock / API 呼び出しの非同期 UI が中心。TanStack Query は loading / error / cache を宣言的に扱える。ローカル UI 状態は Context だけで十分。

[Answer]: X — Jotai を採用。UI 状態（モーダル開閉、サジェスト表示状態、メイン画面の描画フラグ等）は Jotai の atom ベースで管理する。サーバ状態（残高取得・履歴取得・サジェスト取得・注文実行等）は TanStack Query と併用する（Jotai atom と Query の組み合わせは `jotai-tanstack-query` 等を利用、もしくは各 Query の結果を atom に反映させる薄いラッパで統合）。

---

### Question D: API 通信プロトコル

バックエンドとの通信は？

A) **REST (JSON over HTTPS)** — API Gateway + Lambda プロキシ統合の基本形
B) **GraphQL** — AppSync 経由、単一エンドポイント、オーバーフェッチ回避
C) **REST + WebSocket 併用** — 先回り提案のリアルタイム配信を WebSocket で
X) Other

**推奨**: **A (REST)**  
理由: エンドポイント数が少ない（注文・サジェスト・ウォレット・メトリクス）、WebSocket は不要（先回り提案はアプリ起動時のみ: FR-SUGGEST-05）、GraphQL は本スコープではオーバーキル。

[Answer]: A — REST (JSON over HTTPS)。API Gateway (HTTP API もしくは REST API) + Lambda プロキシ統合 → LWA → Gin の構成。

---

### Question E: コンポーネント境界の粒度

Unit 候補（6 つ）に対するコンポーネント境界の粒度は？

A) **1 Unit = 1 Lambda + 1 DynamoDB テーブル**（モノリス風、Unit 境界が素直）
B) **1 Unit = 複数 Lambda（ユースケース単位）+ 共有テーブル**（細粒度、テスタビリティ高い）
C) **Hexagonal / Clean Architecture**（Domain / Application / Adapter 三層を厳密に）
X) Other

**推奨**: **B（ユースケース単位 Lambda）**  
理由: stories.md の Acceptance Criteria に対応する粒度で Lambda を分離すると、冪等性キー・残高不変条件のテストが書きやすい。1 Unit 内でも `placeOrderHandler` / `getSuggestHandler` / `resetBudgetHandler` のように分割。

**注記（Go + Gin + LWA 採用後の再提示）**:
- A'. 1 Unit = 1 Lambda（Gin 内部ルーティング、5 Lambda + スケジューラ 1）
- B'. 単一モノリシック Lambda（全 API を 1 Lambda に、スケジューラのみ別）
- C'. ドメイン境界で粗く分割（3 Lambda 程度）

[Answer]: B' — 単一モノリシック Lambda を採用。全 API エンドポイント（認証・ウォレット・注文・サジェスト・メトリクス）を 1 つの Lambda（Go + Gin + LWA）に集約する。月初の予算リセットは EventBridge Scheduler から起動する別の Lambda（`monthlyResetHandler`）として分離する。つまり実質 **API Lambda × 1 + Scheduler Lambda × 1 の計 2 Lambda** 構成。Unit 分解（Units Generation ステージ）は「コード上のモジュール境界」として表現し、Lambda 物理分割とは独立させる。

---

### Question F: アダプタ層の実装パターン

`DeliveryAdapter`（Q9 で決定済み）の実装パターンは？

A) **インターフェース + 複数実装**（`DeliveryAdapter` interface + `MockDeliveryAdapter` 実装、DI コンテナは使わず手動注入）
B) **tsyringe / awilix などの DI コンテナ**を導入（将来の実装追加を想定）
C) **Function factory パターン**（`createDeliveryAdapter()` を渡して差し替え、クラス不要）
X) Other

**推奨**: **A または C**（小規模 MVP では DI コンテナはオーバーキル）。
- A: OOP 寄りで可読性高い
- C: 関数型寄りでテストが簡潔

**Go 採用後の再提示**:
- A'. Go interface + struct + 手動注入（関数引数 or Dependencies 構造体経由）
- B'. `google/wire` や `uber-go/fx` などの DI フレームワーク
- C'. Function type（`type PlaceOrder func(ctx, req) (res, err)`）

[Answer]: A' — Go の interface + struct パターンを採用。`DeliveryAdapter` interface を定義し、`MockDeliveryAdapter` struct が実装する。main.go で Dependencies 構造体に束ねて手動注入し、Gin ハンドラには依存オブジェクトを渡す。DI フレームワークは導入しない。

---

### Question G: Bedrock 呼び出しのエラーハンドリング戦略

Bedrock が失敗・タイムアウトしたときの挙動は？

A) **フォールバック: 履歴最頻パターンをコード内で返す**（サジェスト品質は落ちるが体験は継続）
B) **エラー応答 + UI はフォールバック表示**（「デフォルトカテゴリで進めますか？」）
C) **リトライ（指数バックオフ 1回）→ フォールバック**（A と B の中間）
X) Other

**推奨**: **C**  
理由: Bedrock の一時的スロットルで全体が止まらないようにしつつ、過度なリトライでユーザを待たせない。MVP として現実的。

[Answer]: C — 指数バックオフで 1 回リトライし、それでも失敗したら履歴最頻パターン（もしくは固定のデフォルトプラン。例: CoCo壱番屋 カレーライス ¥1,200）を返すフォールバックを実装する。フォールバック発動時も US-1-01 の注文フローは正常完了させ、ユーザ体験（「即時解決」）は維持する。内部的に fallback 発動フラグを構造化ログに記録し、Bedrock 呼び出し失敗率を観測できるようにする。

---

### Question H: 認証の保護範囲

どの API を Cognito 認証必須にしますか？

A) **全 API** — ランディング・ヘルスチェック以外は JWT 必須
B) **コアのみ** — 注文・ウォレット・履歴系は JWT 必須、サジェストは任意
C) **緩く** — デモ用途で認証は最小限
X) Other

**推奨**: **A**  
理由: 金融アプリらしさと実装シンプルさの両立。API Gateway の Cognito Authorizer を全保護対象エンドポイントに適用。

[Answer]: A — ヘルスチェック（`/health`）以外の全 API Gateway エンドポイントに Cognito User Pool Authorizer を適用する。JWT 検証はゲートウェイ層で完結し、Lambda (Gin) には JWT の claims（userId 等）が event.requestContext.authorizer 経由で渡される。Gin ミドルウェアで userId を context に載せてハンドラから参照できるようにする。

---

## 2. ストーリー生成タスク一覧（Part 2 で実行）

ユーザ回答後、以下の成果物を生成します：

### 成果物ファイル
- [ ] `aidlc-docs/inception/application-design/components.md` — コンポーネント定義と高レベル責務
- [ ] `aidlc-docs/inception/application-design/component-methods.md` — メソッドシグネチャ
- [ ] `aidlc-docs/inception/application-design/services.md` — サービス定義とオーケストレーション
- [ ] `aidlc-docs/inception/application-design/component-dependency.md` — 依存関係図（Mermaid）とデータフロー
- [ ] `aidlc-docs/inception/application-design/application-design.md` — 上記を統合した集約ドキュメント

### 検証項目
- [ ] 各コンポーネントが少なくとも 1 ストーリー（US-*-*）に紐付く
- [ ] 各コンポーネントが少なくとも 1 機能要件（FR-*-*）に紐付く
- [ ] ダメ化UX NFR との対応を明示（体感 3 秒・1 タップ原則など）
- [ ] アダプタ層の差し替え可能性を依存関係図に明示
- [ ] 6 Unit 候補（stories.md 末尾）への分解示唆を保ちつつ、より具体的な依存関係として可視化

### 状態更新
- [ ] `aidlc-docs/aidlc-state.md` の Stage Progress で Application Design を [x]
- [ ] `aidlc-docs/audit.md` に完了記録を追加

---

## 3. 完了条件

- 全 `[Answer]:` タグに回答あり
- 上記チェックリストが全て [x]
- 5 つの成果物ドキュメントがユーザ承認済み
