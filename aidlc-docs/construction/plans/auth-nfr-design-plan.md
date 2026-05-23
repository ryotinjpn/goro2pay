# Auth Unit — NFR Design Plan

**Document Version**: 0.1 (Draft, awaiting user answers)
**Created**: 2026-05-22
**Unit**: A (`auth` / 認証)
**Construction Depth**: Standard
**Stage**: NFR Design (Construction Phase)
**Prerequisite**: NFR Requirements 承認済み (PR #67 マージ済み)

---

## 1. Plan の目的と範囲

本 Plan は、Unit A（認証）の **NFR Design ステージ** を遂行するための作業計画と、ユーザへの確認質問を定義する。NFR Requirements で確定した数値・しきい値を **どう実現するか（パターン・論理コンポーネント）** を設計する。

### 1.1 NFR Requirements からの引き継ぎ事項

| 項目 | 値 | NFR Design で扱う論点 |
|---|---|---|
| Token Validity | 8h / 8h / 30d | Cognito App Client の TokenValidityUnits の表現方針、Frontend での expiration 監視パターン |
| ログイン応答 P95 | 1.5s | Frontend での Loading 状態管理パターン |
| Stage Throttling | 100 req/s, Burst 200 | API Gateway 設定の表現、429 応答のクライアント側ハンドリング |
| 構造化ログ 8 項目 | level/ts/userId/action/traceId/requestId/email_hash/userAgent | slog Handler 設計、Lambda runtime context との連携、email_hash 生成タイミング |
| Auth エラー処理 | INVALID_CREDENTIALS / NETWORK_ERROR 等 9 種別 | Frontend authMessages 注入パターン、HTTP interceptor の二重発火防止 |
| Pre Sign-up Trigger 失敗 | Cognito 標準（DLQ なし） | Trigger Lambda 実装の例外伝播パターン |
| Token Refresh 失敗 | tokenRefresh_failure → 1.5s モーダル → /login | Auth Hub リスナーパターン、モーダルの二重発火防止 |
| PBT 対象 | normalize / emailHash の 5 プロパティ | gopter / fast-check のジェネレータ設計、CI 組込パターン |
| AuthGuard | (authenticated) レイアウトで囲む | クライアント側 SSR/CSR タイミング、Loading state パターン |
| ブラウザ対応 | モダン最新 2 バージョン | polyfill 不要の確認、autoComplete 属性方針 |

### 1.2 NFR Design で扱うこと / 扱わないこと

| 扱う | 扱わない |
|---|---|
| 設計パターン（Loading state / Retry / Circuit / Hub Listener / Interceptor） | パターンの具体実装コード（Code Generation で書く） |
| 論理コンポーネント（slog Handler / authMessages Resource / SessionExpiredModalHost / AuthGuard / apiClient interceptor） | 論理コンポーネントの内部実装ロジック |
| クライアント側 / サーバ側の責務分担 | リソース ARN / 環境変数の具体値（→ Infrastructure Design） |
| Cognito 設定値の論理表現 | Terraform HCL の具体記述（→ Infrastructure Design） |
| ログ・テストのフレームワーク統合パターン | テストケース実装本体（→ Code Generation） |

---

## 2. 作業手順（Checkboxes）

ユーザ承認後、以下の順序で実施する。

- [x] §3 の質問にユーザから回答を得る（対話形式・1問ずつ）
- [x] 回答の曖昧さを点検し、必要なら追加質問を挟む（Q-D1〜D8/D10/D11 で AI 比較表 + 推奨提示を実施）
- [x] `aidlc-docs/construction/auth/nfr-design/nfr-design-patterns.md` を作成
- [x] `aidlc-docs/construction/auth/nfr-design/logical-components.md` を作成
- [ ] aidlc-state.md と audit.md を更新（承認後）
- [ ] 完了メッセージを提示し、承認ゲートに進む

---

## 3. 確認質問（対話ヒアリング対象）

各質問は対話形式（1 問ずつ提示）でヒアリングする。回答は本ファイルの `[Answer]:` タグに反映する。

### Q-D1: Frontend 401 検出の interceptor 実装パターン

`apiClient` での 401 検出 → `triggerSessionExpired()` のフローをどう実装するか。

| 案 | パターン | 特徴 |
|---|---|---|
| A | **手書き fetch ラッパ**（ファクトリ関数） | 依存ゼロ。`fetch()` を一段階ラップする関数を共通モジュールに置く。Functional Design `frontend-components.md §10` の方針 |
| B | **`@aws-amplify/api` の REST API クライアント** | Amplify が提供するクライアント。Auth と統合された signing 機能あり |
| C | **TanStack Query の `onError` グローバル設定** | API 呼び出しが Query/Mutation 経由なら一元管理可能 |

A は明示的でテストしやすい。C はライフサイクルが Query に紐づく。

[Answer]: **A**（手書き fetch ラッパ）。Functional Design `frontend-components.md §10` の方針と一致、依存ゼロ、テスト容易、認証経路と業務 API の双方を共通 interceptor で扱える。Amplify Auth (Q-A7=A) と API クライアントの責務を分離。なおユーザから「推奨はどれ？」の確認あり、本 AI 応答で 3 案を「Functional Design 整合 / 依存追加 / テスト容易性 / Amplify 連携 / コード行数」の観点で比較表提示済み。

### Q-D2: Token Refresh 失敗 / 401 検出の二重発火防止

複数の API コールが同時に 401 を返したり、Auth Hub `tokenRefresh_failure` と 401 が両方発火したりした際、Modal が複数開かないよう抑止する仕組み。

| 案 | パターン |
|---|---|
| A | **モジュールスコープのフラグ**（`let isHandling = false`）で once-only 動作 |
| B | **Jotai atom + atom write の atomic 性**で状態管理（`sessionExpiredAtom`） |
| C | **イベントバス + first-event-wins**（`Hub.dispatch` で抑止） |

ダメ化UX 的には「1 度だけモーダル → /login」が最重要。

[Answer]: **B**（Jotai atom）。`sessionExpiredAtom` で modal open + `/login` 遷移トリガ + リセットを一元管理。Application Design で採用済み Jotai に乗せ、Provider 隔離でテスト容易、DevTools で可視化、モーダル閉じた後のリセット忘れリスク回避。なおユーザから「推奨はどれ？」の確認あり、本 AI 応答で 3 案を「既存スタック整合 / テスト容易性 / レース安全 / 連動表現 / デバッグ / リセット忘れ」の観点で比較表提示済み。

### Q-D3: AuthGuard の Loading 状態の見せ方

`useAuth().status === "loading"` の間、`(authenticated)` 配下のページで何を見せるか。

| 案 | 内容 |
|---|---|
| A | **フルスクリーンスピナー**（`<FullScreenLoader />`）。シンプルでシステム的 |
| B | **空白画面**（何も表示しない）。チラつきは消えるが反応がないと感じやすい |
| C | **スケルトン UI**（コンテンツの形だけ薄く描画）。体感速度が良い |
| D | **ダメ化トーン軽の文言**（「ちょっと待って…」）と簡易スピナー |

[Answer]: **B+α**（300ms まで空白、それ以降スピナー）。NFR-DEG-01「認知負荷ゼロ」を最優先、認証判定の通常時間 (100-300ms) なら何も見せず低摩擦、長時間 (300ms 超) のみ FullScreenLoader でフィードバック。`setTimeout(() => setShowLoader(true), 300)` パターンで実装。なおユーザから「ダメ化UXに従うとどうなる？」の確認あり、本 AI 応答で各案を NFR-DEG-01/02/03/05 適合度で評価し折衷案 B+α を提示済み。

### Q-D4: パスワード送信前の状態クリア戦略

R-Pwd-3-c「送信完了次第メモリから速やかに参照を断つ」の実装パターン。

| 案 | パターン |
|---|---|
| A | **送信成功・失敗どちらも setState("") でクリア**（finally 句） |
| B | **送信ボタン押下直後に input value を即時クリア + 内部変数に退避**（DOM 上にも残さない） |
| C | **Auth API の結果に関わらず `useEffect` でフォーム unmount で確実にクリア** |

[Answer]: **A**（finally 句で setState("")）。成功時は画面遷移で自然に消え、失敗時もパスワード欄だけクリアされてユーザは即再入力可能。R-Pwd-3-c「速やかに参照を断つ」要件と整合、最もシンプル。`setIsSubmitting(false)` と一緒に `setPassword("")` を呼ぶ。なおユーザから「どれが UX 的にいいの？」の確認あり、本 AI 応答で 3 案を成功時/失敗時/ネットワークエラー時のシナリオ別 + セキュリティ vs UX のトレードオフで比較表提示済み。

### Q-D5: 構造化ログ Handler の構成

NFR-OBS-01 (8 項目構造化ログ) を Go 標準 `log/slog` でどう構成するか。

| 案 | パターン |
|---|---|
| A | **`slog.NewJSONHandler` をそのまま使い、各ログ呼び出しで個別に項目を渡す** |
| B | **カスタム Handler でデフォルト属性を自動付与**（`handler.WithAttrs([...])` で `requestId` / `traceId` を request スコープで束ねる） |
| C | **`slog.SetDefault` + middleware が context に attrs を仕込み、ログ呼び出しが context から拾う** |

middleware で context に詰めるパターンが Gin と相性良い。

[Answer]: **C**（middleware + context-based）。AttachUserID と同じ middleware で `requestId` / `traceId` / `userId` / `userAgent` を context に注入、custom slog Handler が context から属性を自動抽出。`slog.InfoContext(ctx, ...)` で 8 項目自動出力、他 Unit (B/C/D/E) でも同じパターン適用可能。Go 1.22+ 推奨パターン。なおユーザから「推奨は？」の確認あり、本 AI 応答で 3 案を「request スコープ属性 / 遅延注入 / Unit 横断 / Go 慣習 / テスト容易性」の観点で比較表 + 実装イメージ提示済み。

### Q-D6: email_hash 生成のタイミング

ログ用 `email_hash` (SHA256) をいつ計算するか。

| 案 | 内容 |
|---|---|
| A | **ログ呼び出しの直前に都度計算**（パフォーマンス影響微少、シンプル） |
| B | **Auth Trigger / handler 入口で 1 度計算 → context に保存** → ログから引く（重複計算回避） |
| C | **クライアント側で計算 → リクエストヘッダ `X-Email-Hash` で送る**（Server は計算しない） |

[Answer]: **B**（middleware / handler 入口で 1 度計算 → context）。Q-D5=C の context-based slog パターンと一貫性。認証必須エンドポイント (Q-D5 の AttachUserID middleware で claims から email 取得) と認証前エンドポイント (Signup/Login handler で request body から email 取得) で同じ context 注入パターン。重複計算回避、改ざん耐性あり、他 Unit にも展開可能。なおユーザから「推奨は？」の確認あり、本 AI 応答で 3 案を「Q-D5 整合 / 重複計算 / email 取得経路統一 / 改ざん耐性 / 他 Unit 展開」の観点で比較表 + B のサブパターン B-1/B-2 提示済み。

### Q-D7: Pre Sign-up Lambda Trigger の実装言語

Trigger Lambda は API Lambda と独立。実装言語を統一するかどうか。

| 案 | 言語 | 特徴 |
|---|---|---|
| A | **Node.js**（AWS 公式サンプル多数、5 行程度の最小実装） | 起動時間最速、保守容易 |
| B | **Go**（API Lambda と統一） | リポジトリ内で言語統一、ただし Trigger 用に独立ビルド必要 |
| C | **Python** | AWS デフォルト・公式ドキュメント豊富、本プロジェクトでは異物 |

A が一般的。B は統一性のメリットだがビルドオーバーヘッド。

[Answer]: **A**（Node.js）。Pre Sign-up Trigger は 5 行の最小実装、Cold Start 最短 (50-100ms)、AWS 公式サンプル・デファクトスタンダード、Console でも直接編集可能、ビルド不要 (zip 直)。リポジトリ言語統一を多少犠牲にしても利得が大きい。なおユーザから「推奨は？」の確認あり、本 AI 応答で 3 案を「実装サイズ / Cold Start / 公式サンプル / ビルド / 統一性 / 保守 / Trigger 慣習」の観点で比較表 + Node.js / Go の典型コード提示済み。

### Q-D8: PBT の CI 組込パターン

A-NFR-TEST-02 の PBT を CI でどう走らせるか。

| 案 | パターン |
|---|---|
| A | **通常の `go test` の一部として実行**（`gopter` を `_test.go` に含める）。タグも分離も不要 |
| B | **Build tag `pbt` で分離**（`go test -tags=pbt`）、CI で nightly 実行 |
| C | **PR 単位で必ず実行**、ただし時間予算を 30 秒以内に抑える設定 |

[Answer]: **A**（通常の `go test` の一部として実行）。5 プロパティの実行時間は 1-3 秒で軽量、build tag や nightly 分離は YAGNI。コミットごとに反例検出する PBT のメリットを最大化。flaky リスクは低い（決定的関数 + seed 固定可）。重い PBT が将来追加された時点で分離検討。なおユーザから「どれが推奨？」の確認あり、本 AI 応答で 3 案を「実行時間 / フィードバック速度 / 設定複雑さ / PBT メリット最大化 / 将来拡張 / flaky リスク」の観点で比較表 + 実行時間見積もり提示済み。

### Q-D9: Cognito 設定の論理表現粒度

NFR Design では Terraform HCL を書かないが、Cognito の **論理パラメータ** をどこまで明文化するか。

| 案 | 内容 |
|---|---|
| A | **必須項目のみ**（UserPoolName, PasswordPolicy, AutoVerifiedAttributes, TokenValidityUnits, MfaConfiguration, LambdaConfig） |
| B | **A + 推奨項目**（UsernameAttributes, AccountRecoverySetting, AdminCreateUserConfig） |
| C | **全項目網羅**（DeletionProtection, EmailConfiguration, UserPoolAddOns 等） |

Infrastructure Design 側で Terraform を書く前提なので、論理層では概念だけ押さえれば十分。

[Answer]: **B**（A + 推奨項目）。UserPoolName / PasswordPolicy / AutoVerifiedAttributes / TokenValidityUnits / MfaConfiguration / LambdaConfig + UsernameAttributes / AccountRecoverySetting / AdminCreateUserConfig。設定意図が読めるバランス、Infrastructure Design への引き継ぎが明確。

### Q-D10: ダメ化風モーダル (SessionExpiredModal) の表示パターン

「お疲れ様でした」モーダルを 1.5 秒表示してから遷移する実装パターン。

| 案 | パターン |
|---|---|
| A | **`useEffect` + `setTimeout`** で時間管理、`router.push` を発火 |
| B | **Promise.race + UI animation 終了イベント** で時間と動作を同期 |
| C | **CSS animation の onAnimationEnd で確実に遷移**（時間ハードコード回避） |

A はシンプル。C はビジュアルとの同期取れる。

[Answer]: **A**（useEffect + setTimeout）。Functional Design `frontend-components.md §9` の方針と一致、Q-D2=B Jotai atom と自然連携、OK ボタン即時遷移は `clearTimeout` で実装、テスト容易（jest fake timers）。Animation 同期は不要（NFR-DEG-01 低摩擦が最優先、最小演出で十分）。なおユーザから「どれが推奨？」の確認あり、本 AI 応答で 3 案を「実装コスト / 既存 Animation 設計 / テスト / atom 連携 / OK 即時遷移 / ビジュアル一貫性」の観点で比較表 + 実装イメージ提示済み。

### Q-D11: クライアント Throttling (429) ハンドリング

Stage Throttling で 429 が返った場合の Frontend の振る舞い。

| 案 | パターン |
|---|---|
| A | **AuthError.RATE_LIMIT_EXCEEDED にマップして同じダメ化メッセージ表示**（既存パスに合流、追加実装ゼロ） |
| B | **Retry-After ヘッダを尊重して自動リトライ**（指数バックオフ 1 回） |
| C | **A + 一定時間ボタン disabled にする UX 配慮** |

ピーク 10 req/s 想定なので滅多に発生しないが、起きた時の体験を決める。

[Answer]: **A**（AuthError.RATE_LIMIT_EXCEEDED にマップ）。Functional Design `domain-entities.md §6.2` で既に定義済みのエラーコード・文言「ちょっと頑張りすぎです。少し待ってから試してください」を再利用、追加実装ゼロ、YAGNI 原則。本番化時に B / C を再検討する旨を NFR Design ドキュメントに明記。なおユーザから「推奨は？」の確認あり、本 AI 応答で 3 案を「発生頻度 / 追加実装コスト / authMessages 整合 / ユーザ混乱回避 / 観測性 / NFR-DEG-01 / YAGNI」の観点で比較表 + 実装イメージ提示済み。

### Q-D12: Tech Stack バージョン pin の責務移譲

NFR Design の `tech-stack-decisions.md` (NFR Requirements で生成済み) に書いた version 目安を、NFR Design でさらに踏み込むか。

| 案 | 内容 |
|---|---|
| A | **NFR Design では追加で何もしない**（version pin は Code Generation で go.mod / package.json 確定時に行う） |
| B | **NFR Design で minor version まで明示**（例: gin v1.10.0、Amplify 6.5.x） |

通常 A が標準。

[Answer]: **A**（NFR Design では追加で何もしない）。version pin は Code Generation で `go.mod` / `package.json` 確定時に行う。NFR Design は設計パターンと論理コンポーネントに集中、version 固定は実装ステージの責務。

---

## 4. 想定成果物（Plan 承認後に生成）

| ファイル | 内容概要 |
|---|---|
| `nfr-design-patterns.md` | 設計パターン集: Resilience（401 interceptor / Token Refresh / Pre Sign-up failure）/ Performance（Loading state / 認証付き API のキャッシュなしポリシー）/ Security（Token 取扱・パスワード即時クリア・email_hash）/ Observability（slog Handler / context-based attrs）/ Testability（PBT 統合）の各パターンを本 Unit 向けに記述 |
| `logical-components.md` | 論理コンポーネント定義: slog Handler / authMessages Resource / apiClient interceptor / Auth Hub Listener / SessionExpiredModalHost / AuthGuard / Pre Sign-up Lambda（論理） / Cognito 設定の論理パラメータ群 |

---

## 5. 想定外の論点（後続ステージへの引き継ぎ）

- Cognito User Pool の Terraform HCL → **Infrastructure Design**
- API Gateway Stage Throttling の Terraform HCL → **Infrastructure Design**
- Pre Sign-up Lambda の Node.js コード本体（5 行程度）→ **Code Generation**
- `apiClient` / `useAuth` / `AuthGuard` / `SessionExpiredModal` の実装 → **Code Generation**
- PBT テストケース本体 → **Code Generation**

---

## 6. 承認ゲート

本 Plan の構造（質問項目・成果物範囲・作業手順）について以下のいずれかを選択してください:

- 🔧 **Request Changes** — 質問の追加削除や成果物範囲の修正
- ✅ **Approve & Start Q&A** — 上記の質問 Q-D1 〜 Q-D12 を対話形式で順にヒアリング開始
