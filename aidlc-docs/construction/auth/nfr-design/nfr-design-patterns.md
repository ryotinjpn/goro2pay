# Auth Unit — NFR Design Patterns

**Document Version**: 1.0
**Created**: 2026-05-22
**Unit**: A (`auth`)
**Construction Depth**: Standard
**Stage**: NFR Design / Construction
**Predecessors**: NFR Requirements 完了 (PR #67)

本ドキュメントは Unit A の **設計パターン集** を定義する。NFR Requirements で確定した数値・しきい値を実現するための論理レベルのパターンを記述し、実装コードは Code Generation で書く。

参照: [auth-nfr-design-plan.md](../../plans/auth-nfr-design-plan.md), [nfr-requirements.md](../nfr-requirements/nfr-requirements.md), [unit-interfaces.md](../../interfaces/unit-interfaces.md), [Functional Design](../functional-design/)

---

## 1. Resilience Patterns（耐障害性）

### P-RES-01: Frontend HTTP Client Interceptor (Q-D1)

**目的**: 認証経路と業務 API の双方で 401 / 429 / NetworkError を一元検出する。

**パターン**: 手書き fetch ラッパ（ファクトリ関数）。`@aws-amplify/api` や TanStack `onError` には依存させない。

**設計責務**:
- すべての API 呼び出しは `apiClient.request(...)` 経由とする（直接 `fetch` を呼ばない）
- リクエスト前に Amplify Auth `fetchAuthSession()` で IdToken を取得し `Authorization: Bearer` を付与
- レスポンス受信時:
  - `401` → `triggerSessionExpired()` を呼び、`AuthErrorWithCode('SESSION_EXPIRED')` を throw
  - `429` → `AuthErrorWithCode('RATE_LIMIT_EXCEEDED')` を throw（P-RES-04）
  - `5xx` / NetworkError → `AuthErrorWithCode('NETWORK_ERROR')` を throw

**論理シグネチャ**:
```ts
type ApiClient = {
  request(input: RequestInit & { url: string }): Promise<Response>;
};
```

### P-RES-02: 401 / Token Refresh 失敗の二重発火防止 (Q-D2)

**目的**: 複数の API コールが同時に 401 を返したり、Auth Hub `tokenRefresh_failure` と 401 が両方発火しても、`SessionExpiredModal` を 1 度だけ表示する。

**パターン**: Jotai atom の atomic write による once-only 動作。

**設計責務**:
- `sessionExpiredAtom`（`atom<{ openedAt: number } | null>(null)`）を 1 つ用意
- `triggerSessionExpired()` は atom を read → null なら `{ openedAt: Date.now() }` を write、すでに値があれば no-op
- `SessionExpiredModalHost` がこの atom を購読、null 以外なら表示
- 1.5 秒後に atom を null に戻し、`router.push('/login?from=session_expired')`
- OK ボタン即時押下時も同じ atom 操作

**他の発火経路との接続**:
- 401 検出（P-RES-01）→ `triggerSessionExpired()`
- Auth Hub `tokenRefresh_failure` イベント → `triggerSessionExpired()`
- どちらが先でも atom が null でなければ二重発火しない

### P-RES-03: Pre Sign-up Lambda Trigger 失敗時の伝播 (Q-N9 と整合)

**目的**: auto-confirm を担う Pre Sign-up Lambda の失敗を Cognito 標準フローに任せ、ユーザにエラーを返す。

**パターン**: 例外をそのまま伝播する（捕捉せず）。

**設計責務**:
- Trigger 関数は `event.response.autoConfirmUser = true` / `autoVerifyEmail = true` をセットして event を返す
- 内部で例外が発生した場合、Lambda runtime が CloudWatch Logs に自動出力 → Cognito が SignUp 全体を失敗扱い
- Frontend は Cognito の `SignUpException` 系を `AuthErrorWithCode` にマップ
- DLQ・カスタムログ追加・自動リトライは行わない

### P-RES-04: クライアント側 429 ハンドリング (Q-D11)

**目的**: API Gateway Stage Throttling（A-NFR-SEC-04: 100 req/s, Burst 200）で 429 が返った際、ユーザにダメ化トーン軽の文言で告知する。

**パターン**: 既存の `AuthError.RATE_LIMIT_EXCEEDED` に合流させる。

**設計責務**:
- `apiClient` interceptor が `res.status === 429` を検知 → `AuthErrorWithCode('RATE_LIMIT_EXCEEDED')` を throw（P-RES-01）
- `authMessages.ts` の既存マッピングをそのまま再利用: 「ちょっと頑張りすぎです。少し待ってから試してください」
- 自動リトライ・ボタン disabled 等の追加 UX は本 MVP では実装しない
- 本番化時に `Retry-After` ヘッダ尊重・指数バックオフ・ボタン disabled タイマーを再評価する旨を後続 Code Generation の備考に明記

### P-RES-05: Cognito 障害時のフォールバック (A-NFR-AVAIL-01 と整合)

**目的**: Cognito 自体の障害時、Unit A は独自リトライをせず Cognito SLA に追従する。

**パターン**: フェイルファスト + 静的エラーメッセージ。

**設計責務**:
- Amplify Auth が内部で標準リトライを試みた後の結果を尊重
- 最終的に失敗した場合は `AuthErrorWithCode('NETWORK_ERROR')` を throw（P-RES-01 と同じ経路）
- 専用障害告知ページ・サーバ側追加対応は実装しない

---

## 2. Performance Patterns（性能）

### P-PERF-01: AuthGuard の Loading 表示遅延 (Q-D3)

**目的**: NFR-DEG-01「認知負荷ゼロ」を最優先しつつ、認証判定が長時間かかる場合だけフィードバックを返す。

**パターン**: 300ms までは空白、以降は `<FullScreenLoader />` 表示。

**設計責務**:
- `useAuth().status === "loading"` の間、内部 state `showLoader` は初期値 `false`
- `useEffect` で `setTimeout(() => setShowLoader(true), 300)` をセットし、unmount or status 変化時に `clearTimeout`
- `<AuthGuard>` は `status === "loading" && !showLoader` のとき null を返す
- `status === "loading" && showLoader` のとき `<FullScreenLoader />` を返す
- `status === "authenticated"` で children を表示
- `status === "unauthenticated"` で `router.replace('/login')`

### P-PERF-02: 認証付き API のキャッシュなしポリシー (Functional Design 整合)

**目的**: A-NFR-OBS-01 のログ整合・セキュリティ観点から、認証付き API レスポンスをブラウザ・CDN にキャッシュしない。

**パターン**: レスポンスヘッダ `Cache-Control: no-store` を Lambda 側で付与。

**設計責務**:
- 認証必須エンドポイント全てが `Cache-Control: no-store` を返す
- Login / Logout は特に明示的に no-store
- API Gateway / Amplify Hosting の Cache 設定は本 MVP では追加カスタマイズしない（Lambda レスポンスヘッダのみで担保）

### P-PERF-03: ログイン応答時間目標達成のための制約 (A-NFR-PERF-01)

**目的**: P50 500ms / P95 1500ms を達成する。

**設計上の制約**:
- Cognito API は Tokyo リージョン (`ap-northeast-1`) 内で完結
- Lambda コールドスタート対策は本 MVP では実装しない（Provisioned Concurrency 等は使わない、A-NFR-SCALE-02）
- 同期処理のみ、追加の外部 I/O (DB アクセス等) を Auth フローに挟まない

---

## 3. Security Patterns（セキュリティ）

### P-SEC-01: パスワード状態の即時クリア (Q-D4, R-Pwd-3-c)

**目的**: パスワードを送信完了次第メモリから速やかに参照を断つ。

**パターン**: `try-finally` 句で `setPassword('')` を実行。

**設計責務**:
- Login / Signup フォームの `handleSubmit` 内で:
  ```pseudo
  try:
    AmplifyAuth.signIn(...) (or signUp)
  finally:
    setPassword('')
  ```
- React の制御コンポーネント (controlled input) として state とフォーム表示を 1:1 にしておくことで、state クリアと同時に DOM input value も空になる
- `autoComplete="new-password"` (Signup) / `autoComplete="current-password"` (Login) で password manager 連携は維持
- パスワードは sessionStorage / localStorage / Cookie / IndexedDB 等の永続層に決して書かない

### P-SEC-02: email_hash 生成タイミング (Q-D6)

**目的**: ログ用 `email_hash` (SHA256 hex) を 1 リクエスト 1 回だけ計算し、`gin.Context` 経由でログ Handler に渡す。

**パターン**: middleware / handler 入口で hash 化 → context に保存 → context-based slog Handler が自動抽出（P-OBS-01 と連携）。

**設計責務**:
- **認証必須エンドポイント**: `AttachUserID` middleware 内で `claims["email"]` から hash 生成 → `c.Set("emailHash", hash)`
- **認証前エンドポイント** (Signup / Login / Logout): handler 内で request body の email を `normalize` → hash 生成 → `c.Set("emailHash", hash)`
- ハッシュアルゴリズム: SHA256、出力は hex 文字列（64 桁）
- Salt は使わない（A-NFR-SEC-08 の目的はトレーサビリティ目的、暗号学的耐性は本 MVP で要求していない）

### P-SEC-03: Token の取扱い (A-NFR-SEC-03 / Functional Design 整合)

**目的**: IdToken / AccessToken / RefreshToken をクライアント側のみで管理し、サーバ側で永続化しない。

**パターン**: Amplify Auth の標準ストレージに委譲。

**設計責務**:
- Amplify Auth はデフォルトで localStorage に Token を保存
- Lambda 側は受信した IdToken の検証を Cognito Authorizer に任せ、claims を Context 経由で読むのみ
- Lambda・DynamoDB に Token 値を保存しない
- ログ出力時 Token 値を含めない（A-NFR-OBS-01 の 8 項目に Token 値は含まれていない）

### P-SEC-04: API Gateway Stage Throttling (A-NFR-SEC-04)

**目的**: ピーク 10 req/s 想定に対して上限 100 req/s / Burst 200 を全 Stage に適用、ブルートフォース攻撃の影響を緩和する。

**パターン**: API Gateway Stage Throttling（Stage 全体に一括適用）。

**設計責務**:
- Stage Throttling のパラメータ:
  - `RateLimit`: 100 req/s
  - `BurstLimit`: 200 req
- IP 単位の throttling・WAF・Cognito Advanced Security Features は本 MVP では導入しない
- Terraform 実装は Infrastructure Design ステージで確定

---

## 4. Observability Patterns（観測性）

### P-OBS-01: 構造化ログ Handler 構成 (Q-D5, A-NFR-OBS-01)

**目的**: 8 項目（level / timestamp / userId / action / traceId / requestId / email_hash / userAgent）の構造化ログを、ハンドラ呼び出し側で項目を意識せずに自動付与する。

**パターン**: middleware が context に attrs を仕込む + custom slog Handler が context から attrs を抽出。

**設計責務**:
- **slog logger 初期化**: アプリ起動時に custom Handler を `slog.SetDefault(logger)` で登録
- **custom Handler の実装責務**:
  - `Handle(ctx context.Context, record slog.Record)` で context から指定キー (requestId / traceId / userId / userAgent / emailHash) を抽出
  - 抽出した attrs を record にマージし、内部の `slog.NewJSONHandler` に委譲
  - 出力先は `os.Stdout`（CloudWatch Logs にそのまま流れる）
- **middleware の責務**:
  - リクエスト受信時に新しい context を生成し、API Gateway / Lambda が付与する `requestId` / `traceId` を読み取り context へ書き込む
  - `userAgent` をリクエストヘッダから取得して context へ書き込む
  - `userId` は `AttachUserID` middleware が claims から書き込む（既存 Functional Design）
  - `emailHash` は P-SEC-02 で書き込む
- **ログ呼び出し側**: `slog.InfoContext(ctx, "user logout", "action", "logout")` のように呼ぶだけ
- **平文 email を絶対に出力しない**: A-NFR-SEC-08 / A-NFR-OBS-01 の制約。レビューで grep `\"email\":` を検査ルールに含めることを推奨

### P-OBS-02: ログレベルの環境別制御

**目的**: 本番デモ時はノイズを抑え、開発時は詳細情報を取得できるようにする。

**パターン**: 環境変数 `LOG_LEVEL`（`info` / `debug`）を slog の `Level` に渡す。

**設計責務**:
- `LOG_LEVEL=info`（デフォルト、本番デプロイ）
- `LOG_LEVEL=debug`（開発・トラブルシュート時）
- middleware ログは `info`、内部 trace は `debug` に切り分け

### P-OBS-03: メトリクス・X-Ray 不採用 (A-NFR-OBS-02 / A-NFR-OBS-03)

CloudWatch Custom Metrics / X-Ray は本 MVP では導入しない。後続 Code Generation で `slog.With(...)` 等の API は使うが EMF 出力はしない。本番化時の再評価事項として nfr-design-patterns.md の備考に明記する。

---

## 5. Testability Patterns（テスト容易性）

### P-TEST-01: PBT の CI 組込 (Q-D8, A-NFR-TEST-02)

**目的**: `normalize` / `emailHash` の純粋関数 PBT を全 PR で軽量に実行する。

**パターン**: 通常の `go test` の一部として実行（Build tag・nightly 分離なし）。

**設計責務**:
- `internal/auth/email_test.go`（または同等のファイル）に `gopter` ベースのテストを配置
- gopter の Generation 数はデフォルト（100）を使用、5 プロパティで合計 1-3 秒
- `seed` を必要なら `MinSuccessfulTests` と組合せて固定可能（flaky 抑制）
- CI ワークフロー側で `go test ./...` を呼ぶだけで実行される

**5 プロパティ** (再掲、A-NFR-TEST-02):
- `normalize` のべき等性
- `normalize` の大文字小文字不問
- `emailHash` の長さ不変（64 hex chars）
- `emailHash` の正規化整合
- `emailHash` の衝突なし（高確率）

### P-TEST-02: middleware と context-based slog の Spy パターン

**目的**: P-OBS-01 で導入した context-based ログ Handler のテストを書きやすくする。

**パターン**: テスト時は custom Handler の `os.Stdout` を `bytes.Buffer` に差し替えて、出力 JSON を assert する。

**設計責務**:
- Handler 生成時に `io.Writer` を引数で受け取れる構造にしておく
- テストでは `&bytes.Buffer{}` を渡し、ログ出力後にバッファを JSON parse して項目を assert

### P-TEST-03: Frontend のテスト戦略

| レイヤ | パターン | ライブラリ |
|---|---|---|
| `useAuth` ロジック単体 | Mock Amplify Auth、各メソッドの happy / error path | Vitest + @testing-library/react |
| 401 interceptor (P-RES-01) | `fetch` をスタブ、401 受信 → atom 書き込み確認 | Vitest |
| `sessionExpiredAtom` 二重発火防止 (P-RES-02) | atom の atomic write を assert | Vitest + Jotai test util |
| `<AuthGuard>` の Loading 切替 (P-PERF-01) | jest fake timers で 300ms 経過を assert | Vitest |
| E2E (US-0-01 / US-0-02) | LocalStack Cognito + Playwright | Playwright |

---

## 6. Logical Components 連携

各パターンが利用する論理コンポーネントの一覧と相互関係は [logical-components.md](./logical-components.md) を参照。

---

## 7. パターン → NFR トレーサビリティ

| パターン | 対応 A-NFR / Q-D 質問 |
|---|---|
| P-RES-01 (interceptor) | A-NFR-AVAIL-01 / Q-D1 |
| P-RES-02 (二重発火防止) | A-NFR-AVAIL-01 / Q-D2 |
| P-RES-03 (Trigger 失敗) | A-NFR-REL-01 / Q-N9 |
| P-RES-04 (429) | A-NFR-SEC-04 / Q-D11 |
| P-RES-05 (Cognito 障害) | A-NFR-AVAIL-01 / Q-N5 |
| P-PERF-01 (Loading 遅延) | A-NFR-PERF-04, NFR-DEG-01 / Q-D3 |
| P-PERF-02 (no-store) | A-NFR-OBS-01 / 暗黙の Best Practice |
| P-PERF-03 (応答時間) | A-NFR-PERF-01 |
| P-SEC-01 (PW クリア) | A-NFR-SEC-07 / Q-D4 |
| P-SEC-02 (email_hash) | A-NFR-SEC-08 / Q-D6 |
| P-SEC-03 (Token 取扱) | A-NFR-SEC-03 |
| P-SEC-04 (Stage Throttle) | A-NFR-SEC-04 / Q-N6 |
| P-OBS-01 (slog Handler) | A-NFR-OBS-01 / Q-D5 |
| P-OBS-02 (LOG_LEVEL) | A-NFR-OBS-01 |
| P-OBS-03 (Metrics 不採用) | A-NFR-OBS-02, A-NFR-OBS-03 |
| P-TEST-01 (PBT CI) | A-NFR-TEST-02 / Q-D8 |
| P-TEST-02 (slog Spy) | A-NFR-TEST-01 |
| P-TEST-03 (Frontend Test) | A-NFR-TEST-01, A-NFR-TEST-04 |

---

## 8. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| **Infrastructure Design** | API Gateway Stage Throttling の Terraform 値、Cognito の TokenValidityUnits・PasswordPolicy・LambdaConfig の HCL 表現、Pre Sign-up Lambda の zip パッケージング |
| **Code Generation** | 各パターンの実装本体（`apiClient` / `useAuth` / `sessionExpiredAtom` / `<AuthGuard>` / `<SessionExpiredModalHost>` / Pre Sign-up Trigger Node.js コード / `normalize` / `emailHash` / custom slog Handler / PBT テスト） |
