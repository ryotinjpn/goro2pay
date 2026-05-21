# Auth Unit — Functional Design Plan

**Document Version**: 0.1 (Draft, awaiting user answers)
**Created**: 2026-05-21
**Unit**: A (`auth` / 認証)
**Construction Depth**: Standard
**Stage**: Functional Design (Construction Phase)
**Prerequisite**: Application Design 承認済み、Units Generation 承認済み

---

## 1. Plan の目的と範囲

本 Plan は、Unit A（認証）の **Functional Design ステージ** を遂行するための作業計画と、ユーザへの確認質問を定義する。Plan 承認後、回答内容を反映した Functional Design 成果物を生成する。

### 1.1 Unit A のスコープ（再掲）

- **責務**:
  - Amazon Cognito User Pool による認証機構（メール+パスワード）
  - API Gateway Cognito Authorizer + Lambda 内 middleware による JWT 検証
  - `userId` の Gin Context への注入（後続 Unit が利用）
  - Frontend の Landing / Login / Signup 画面と認証状態管理
- **対象ストーリー**: US-0-01（新規登録）、US-0-02（ログイン）
- **依存先**: なし（最下位レイヤ）
- **依存元**: Unit B / C / D / E（全 Unit が認証済状態を前提）

### 1.2 含まれるコンポーネント（Application Design より）

**Backend (Go)**
- `internal/auth/` パッケージ
  - `AuthContextService`（Gin middleware: `AttachUserID`, `UserIDFromContext`）

**Frontend (Next.js / App Router)**
- `app/(auth)/login/page.tsx` → `LoginScreen`
- `app/(auth)/signup/page.tsx` → `SignupScreen`
- `app/page.tsx`（未認証時の `Landing`、認証済み遷移）
- `hooks/useAuth.ts`

**Infrastructure（参照のみ。詳細は Infrastructure Design ステージで確定）**
- Cognito User Pool / App Client
- API Gateway Cognito Authorizer

### 1.3 Functional Design で扱うこと / 扱わないこと

| 扱う | 扱わない |
|---|---|
| ドメインエンティティ（User Identity, AuthSession） | Cognito の AWS リソース構成 |
| 業務ルール（パスワード強度、検証ロジック、JWT クレーム解釈） | Lambda リソース・IAM・VPC 等の物理構成 |
| 認証データフロー（Signup / Login / 認証付き API 呼出） | パフォーマンス目標値の設計（NFR Requirements で扱う） |
| エラーシナリオ（認証失敗、期限切れ、レート制限到達） | Cognito Identity Pool（本 MVP では不使用） |
| Frontend コンポーネント階層と状態管理 | UI のスタイリング詳細 |

---

## 2. 作業手順（Checkboxes）

ユーザ承認後、以下の順序で実施する。

- [x] §3 の質問にユーザから回答を得る（対話形式・1問ずつ）
- [x] 回答の曖昧さを点検し、必要なら追加質問を挟む（Q-A4 / Q-A10 で追加説明実施）
- [x] `aidlc-docs/construction/auth/functional-design/business-logic-model.md` を作成
- [x] `aidlc-docs/construction/auth/functional-design/business-rules.md` を作成
- [x] `aidlc-docs/construction/auth/functional-design/domain-entities.md` を作成
- [x] `aidlc-docs/construction/auth/functional-design/frontend-components.md` を作成（Unit A は UI を含むため）
- [ ] aidlc-state.md と audit.md を更新（承認後）
- [ ] 完了メッセージを提示し、承認ゲートに進む

---

## 3. 確認質問（対話ヒアリング対象）

各質問は対話形式（1 問ずつ提示）でヒアリングする。回答は本ファイルの `[Answer]:` タグに反映する。

### Q-A1: 新規登録時のメール検証フロー

US-0-01 では「登録から最初のボタン操作までの総タップ数 5 回以下」が制約。Cognito のデフォルトはメール検証コード（confirmation code）が必須だが、これを採用すると登録 → メール確認 → コード入力 → ログインで 5 回を超えやすい。

| 案 | 内容 | トレードオフ |
|---|---|---|
| A | **メール検証あり**（Cognito 標準フロー、Auto-Verified Email） | 標準的・本人確認確実だが摩擦増 |
| B | **メール検証なし**（Signup 直後に CONFIRMED 状態にする = `AdminConfirmSignUp` または Pre Sign-up Lambda Trigger で auto-confirm） | 摩擦最小、ダメ化UXに整合。本人確認は将来対応 |
| C | **メール検証あり、ただしマジックリンク方式**（Custom Message + Custom Auth Challenge） | 体験は良いが MVP には実装重い |

[Answer]: **B**（メール検証なし、auto-confirm）。Pre Sign-up Lambda Trigger で即 CONFIRMED 状態にし、ダメ化UXの「摩擦最小」を優先。本人確認は将来対応として設計書に明記。

### Q-A2: パスワードポリシーの強度

Cognito のパスワードポリシーをどのレベルにするか。

| 案 | 内容 |
|---|---|
| A | **緩め**: 最小8文字、英数字のみ要求（記号・大文字・小文字混在は不要）→ MVP のダメ化UX に整合 |
| B | **標準**: 最小8文字、英大小・数字混在を要求（記号は不要） |
| C | **厳格**: 最小12文字、大小英・数字・記号すべて要求 |

[Answer]: **B**（標準: 最小8文字、英大小+数字混在を要求、記号は不要）。Cognito デフォルト近い設定でセキュリティと利便性のバランスを取る。

### Q-A3: ログアウトの扱い（FR-AUTH-04）

US-0-01/0-02 は MUST だが、FR-AUTH-04（ログアウト）は要件には MUST だがストーリーは未作成。Functional Design でログアウト機能を扱うか。

| 案 | 内容 |
|---|---|
| A | Functional Design に含める（API: `POST /auth/logout` を仕様化、UI: 設定画面 or ヘッダの簡易ログアウト） |
| B | Functional Design に含めない（クライアント側のトークン削除のみ。API 不要、UI は最小限） |

[Answer]: **A**（API + UI 仕様化）。`POST /auth/logout` を仕様化し、UI はヘッダ or 設定メニューに簡易ログアウトを配置。Cognito の `GlobalSignOut` でリフレッシュトークン取り消し。

### Q-A4: JWT の Lambda 内検証戦略

API Gateway Cognito Authorizer が一次検証を行うが、Lambda（Gin middleware の `AuthContextService`）は何をするか。

| 案 | 内容 |
|---|---|
| A | **Authorizer のみに任せる**: Lambda は `requestContext.authorizer.claims` から `sub` を読み出すだけ（再検証なし） |
| B | **Lambda でも JWT を再検証**: Cognito JWKS を取得して Lambda 内で `Authorization` ヘッダを検証（Authorizer + 二段階） |

A は Cognito Authorizer の標準パターン。B は Authorizer をスキップするユースケース（例: Lambda Function URL 直結）でも使える。

[Answer]: **A**（Authorizer のみ）。Lambda は `event.requestContext.authorizer.claims` から `sub` を読み出すだけ。再検証はしない。本 MVP は API Gateway 経由のみのため二重検証は不要。なおユーザから「JWT をフロントから AccessToken 投げて API GW が Cognito 認証して Lambda に行く」フローの確認あり、本 AI 応答で図解説明済み。

### Q-A5: `userId` として何を採用するか

JWT の `sub`（Cognito ユーザの UUID）と `cognito:username` が候補。後続 Unit のテーブル PK にも影響する。

| 案 | 内容 |
|---|---|
| A | **`sub`（UUID v4）を採用**: 不変・グローバル一意。Cognito 標準。 |
| B | **`cognito:username` を採用**: メールアドレスベース（要件 5.1.2 の例）だが、ユーザ名変更があると不変性が崩れる可能性 |

[Answer]: **A**（`sub` UUID v4 採用）。後続 Unit B/C/D/E のテーブル PK は全て `sub` 文字列とする。メール変更等での不変性破綻を回避。

### Q-A6: 未認証時のメイン画面アクセス

未ログインで `/`（メイン画面）にアクセスした場合の挙動。

| 案 | 内容 |
|---|---|
| A | **Landing 画面を表示**（同一 URL 内で UI 切替） — US-0-04 と整合性あり、初回タップ削減 |
| B | **`/login` にリダイレクト**（Next.js middleware で強制リダイレクト） — 標準的だがタップ増 |

[Answer]: **A**（同一 URL に Landing UI を表示し UI 切替）。`/` で認証状態を判定し、未認証時は LandingScreen、認証済み時は MainScreen を描画。タップ削減と US-0-04 のダメ化UX に整合。

### Q-A7: 認証状態管理ライブラリ

Frontend の認証状態管理に何を使うか。Application Design では Cognito 連携手段に「amplify-js または amazon-cognito-identity-js」を挙げている。

| 案 | 内容 |
|---|---|
| A | **AWS Amplify (`@aws-amplify/auth`)**: Amplify Hosting と統合、Auth/Storage/API がワンセット。学習コスト低 |
| B | **`amazon-cognito-identity-js`**: 軽量、Amplify を入れずに Cognito 直接操作 |
| C | **Cognito Hosted UI（OAuth2 リダイレクト）**: Cognito のホスト型ログイン画面に飛ばす。UI 自作不要だが体験が外に出る |

ダメ化UXとブランディング（自虐コピー）を考慮すると C は不向きと推測。

[Answer]: **A**（AWS Amplify Auth `@aws-amplify/auth`）。Amplify Hosting と統合され、`signIn` / `signUp` / `signOut` / `fetchAuthSession` 等の API が揃う。Auth Hub イベントで状態変化を購読し、`useAuth` フックから配布。

### Q-A8: 認証エラーのユーザ表示文言（ダメ化UXコピー）

NFR-DEG-05 に沿って、認証エラーのコピーをどうするか。

| 案 | 内容 |
|---|---|
| A | **標準的で淡白**:「メールアドレスまたはパスワードが違います」「ネットワークエラー」 |
| B | **ダメ化トーン軽**: 「ログインすらめんどくさいですよね…もう一度試してください」 |
| C | **ダメ化トーン強**: 「ダメすぎて入れません。深呼吸してから再挑戦を…」 |

[Answer]: **B**（ダメ化トーン軽）。誤入力 → 「ログインすらめんどくさいですよね…もう一度お試しください」、ネットワーク → 「通信が…ちょっと待ってもう一度」のように、ノリを残しつつ可読性は維持。

### Q-A9: セッション失効時の動線

JWT 期限切れ（デフォルト 1 時間）でリフレッシュトークンも失効した場合の UX。

| 案 | 内容 |
|---|---|
| A | **再ログイン画面に自動遷移**（標準的） |
| B | **モーダルで「お疲れ様でした、もう一度ログインしてください」表示 → ログイン画面**（ダメ化UX 風味） |

[Answer]: **B**（ダメ化風モーダル → ログイン画面）。401 を検知したら一瞬モーダル「お疲れ様でした、もう一度ログインしてください」を表示してから `/login` に遷移。Amplify Auth Hub の `tokenRefresh_failure` イベントをフックして実装。

### Q-A10: 登録直後のリダイレクト先

US-0-01 の Acceptance Criteria は「ログイン済み状態でアプリのメイン画面へ遷移」とある。一方、Unit B（予算）は「予算未設定の場合は BudgetSetup 画面へ」を匂わせる（unit-of-work.md §3.1 のダメ化UX 織り込み）。Functional Design の段階でこの分岐を仕様化するか。

| 案 | 内容 |
|---|---|
| A | **Unit A 内で分岐ロジックを定義**: Signup 完了 → BudgetSettings 取得 → 未設定なら `/budget`、設定済みなら `/`。`useAuth` が後続 Unit B と連携する責務を持つ |
| B | **Unit A は単純に `/` に遷移、分岐は `/`（MainScreen, Unit C/B 横断）側の責務とする** |

A の方が unit-of-work.md の記述に整合的だが、Unit 境界が曖昧になる。

[Answer]: **B**（`/` 側で分岐）。Unit A は認証完了後 `router.push("/")` するだけ。`/` の MainScreen が `useWallet` の結果を見て予算未設定なら `router.replace("/budget")` する。Unit 境界（unit-of-work.md §5 の依存方向ルール）を維持。チラつきは Next.js のサーバコンポーネントまたは `useEffect` 内 replace で抑制。なおユーザから「メリットデメリットは？」の確認あり、本 AI 応答で表形式比較を提示済み。

---

## 4. 想定成果物（Plan 承認後に生成）

| ファイル | 内容概要 |
|---|---|
| `business-logic-model.md` | Signup / Login / 認証付き API 呼出の業務フロー（Mermaid シーケンス図含む）、フォールバック分岐 |
| `business-rules.md` | パスワード強度、メール検証ポリシー、JWT クレーム解釈、認証エラー分類 |
| `domain-entities.md` | UserIdentity / AuthSession / AuthError 等のドメインエンティティ定義（技術中立） |
| `frontend-components.md` | Landing / SignupScreen / LoginScreen / `useAuth` フックの構造、Props/State、API 連携、検証ルール |

---

## 5. 想定外の論点（後続ステージへの引き継ぎ）

- Cognito User Pool の物理リソース構成（password policy パラメータの最終値、MFA 有無、Token Validity の数値、Email Configuration の SES 利用可否）→ **Infrastructure Design**
- API Gateway Authorizer の TTL、レート制限値 → **Infrastructure Design**
- 性能・可用性目標値（メイン画面 5 秒以内表示の認証フェーズ寄与分）→ **NFR Requirements**
- セキュリティ最低限要件（NFR-SEC-01 〜 04）の実装手段 → **NFR Design**

---

## 6. 承認ゲート

本 Plan の構造（質問項目・成果物範囲・作業手順）について以下のいずれかを選択してください:

- 🔧 **Request Changes** — 質問の追加削除や成果物範囲の修正
- ✅ **Approve & Start Q&A** — 上記の質問 Q-A1 〜 Q-A10 を対話形式で順にヒアリング開始
