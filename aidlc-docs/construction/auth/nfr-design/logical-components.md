# Auth Unit — Logical Components

**Document Version**: 1.2
**Created**: 2026-05-22
**Updated**: 2026-05-22 (BFF パターン採用: LC-AUTH-09 を Browser/Server 二段化、新規 LC-AUTH-18 BffProxyRouteHandler 追加)
**Updated**: 2026-05-22 (API 認証 Token を IdToken → AccessToken に統一、Authorization: Bearer ヘッダで透過、X-Id-Token ヘッダ廃止、LC-AUTH-09 / LC-AUTH-18 / LC-AUTH-01 の責務を更新)
**Unit**: A (`auth`)
**Construction Depth**: Standard
**Stage**: NFR Design / Construction
**Predecessors**: NFR Requirements / [nfr-design-patterns.md](./nfr-design-patterns.md)

本ドキュメントは Unit A の **論理コンポーネント** を定義する。論理レベルでの責務・入出力・他コンポーネントとの関係を明示し、実装本体は Code Generation で書く。

参照: [nfr-design-patterns.md](./nfr-design-patterns.md), [unit-interfaces.md](../../interfaces/unit-interfaces.md), [Functional Design](../functional-design/)

---

## 1. コンポーネント一覧

| ID | 名前 | 配置 | 主要責務 |
|---|---|---|---|
| **LC-AUTH-01** | `AttachUserIDMiddleware` | Backend / Go (`internal/auth/`) | JWT claims から `userId` 抽出 + context attrs 注入 |
| **LC-AUTH-02** | `EmailNormalizer` | Backend / Go (`internal/auth/`) | メールアドレスの正規化（純粋関数） |
| **LC-AUTH-03** | `EmailHasher` | Backend / Go (`internal/auth/`) | SHA256 hex hash 生成（純粋関数） |
| **LC-AUTH-04** | `RequestContextMiddleware` | Backend / Go (`internal/logging/`) | requestId / traceId / userAgent / emailHash を context に注入 |
| **LC-AUTH-05** | `ContextAwareSlogHandler` | Backend / Go (`internal/logging/`) | context から attrs を抽出して JSON 出力 |
| **LC-AUTH-06** | `LogoutHandler` | Backend / Go (`internal/handlers/`) | `POST /api/auth/logout` の最小ハンドラ |
| **LC-AUTH-07** | `PreSignUpTriggerLambda` | AWS Lambda / Node.js (独立) | auto-confirm + auto-verify-email |
| **LC-AUTH-08** | `useAuthHook` | Frontend / TypeScript (`hooks/`) | 認証状態管理 + signup/login/logout の公開 IF |
| **LC-AUTH-09** | `apiClient` | Frontend / TypeScript (`web/lib/`) | **Browser 側** fetch ラッパ + 401/429 interceptor + Authorization: Bearer <accessToken> ヘッダ付与 |
| **LC-AUTH-10** | `sessionExpiredAtom` | Frontend / Jotai atom | セッション失効状態の atomic 管理 |
| **LC-AUTH-11** | `SessionExpiredModalHost` | Frontend / React | atom 購読 + Modal 表示 + 1.5s 後遷移 |
| **LC-AUTH-12** | `AuthGuard` | Frontend / React | 認証必須レイアウトのガード + 300ms Loading 遅延 |
| **LC-AUTH-13** | `AuthMessagesResource` | Frontend / TypeScript (`web/lib/`) | AuthErrorCode → 日本語ダメ化トーン軽メッセージ |
| **LC-AUTH-14** | `AuthHubListener` | Frontend / TypeScript (`web/lib/`) | Amplify Auth Hub `tokenRefresh_failure` 等を監視 |
| **LC-AUTH-15** | `CognitoUserPoolConfig` | AWS / 論理表現 | User Pool の論理パラメータ群 |
| **LC-AUTH-16** | `CognitoAppClientConfig` | AWS / 論理表現 | App Client の論理パラメータ群 |
| **LC-AUTH-17** | `ApiGatewayStageThrottlingConfig` | AWS / 論理表現 | Stage Throttling のレート上限定義 |
| **LC-AUTH-18** | `BffProxyRouteHandler` | Frontend / Next.js (`web/app/api/[...path]/route.ts`) | **Server 側** catch-all proxy: `/api/*` を受けて API Gateway に転送、`Authorization` ヘッダを透過 (Browser 側で AccessToken を付与済み)、`API_ENDPOINT` server-only env 利用 |

---

## 2. Backend 論理コンポーネント

### LC-AUTH-01: `AttachUserIDMiddleware`

| 項目 | 内容 |
|---|---|
| 配置 | `back/api/internal/auth/middleware.go`（仮） |
| 公開 IF | [unit-interfaces.md §2.1](../../interfaces/unit-interfaces.md) — `AttachUserID() gin.HandlerFunc` / `UserIDFromContext(c) (string, error)` |
| 入力 | `gin.Context`（API Gateway → Lambda 経由、`requestContext.authorizer.claims` を保持、AccessToken の claims） |
| 出力 | `c.Set("userId", sub)`<br>※ AccessToken の claims に `email` がないため、emailHash は middleware では生成しない（P-SEC-02、認証前 handler 内でのみ生成） |
| 失敗時 | claims 欠落 → `c.AbortWithStatus(500)` + `slog.ErrorContext(...)`（A-NFR-REL-02） |
| 連携 | [LC-AUTH-04 RequestContextMiddleware](#lc-auth-04-requestcontextmiddleware) |

### LC-AUTH-02: `EmailNormalizer`

| 項目 | 内容 |
|---|---|
| 配置 | `back/api/internal/auth/email.go`（仮） |
| 公開 IF | `Normalize(email string) string` |
| 振る舞い | 前後 trim → `strings.ToLower(...)` |
| テスト | PBT 対象（A-NFR-TEST-02 / P-TEST-01）: べき等性 / 大文字小文字不問 |

### LC-AUTH-03: `EmailHasher`

| 項目 | 内容 |
|---|---|
| 配置 | `back/api/internal/auth/email.go`（仮、Normalizer と同パッケージ） |
| 公開 IF | `Hash(email string) string` |
| 振る舞い | `hex.EncodeToString(sha256.Sum256([]byte(Normalize(email))))` |
| テスト | PBT 対象: 長さ 64 / Normalize 整合 / 衝突なし |

### LC-AUTH-04: `RequestContextMiddleware`

| 項目 | 内容 |
|---|---|
| 配置 | `back/api/internal/logging/middleware.go`（仮） |
| 公開 IF | `RequestContext() gin.HandlerFunc` |
| 振る舞い | リクエスト受信時に下記を `gin.Context` 経由で内部 context へ載せる:<br>- `requestId`（API Gateway / Lambda が付与した値、なければ ULID 生成）<br>- `traceId`（X-Ray header `X-Amzn-Trace-Id` から、なければ null）<br>- `userAgent`（リクエストヘッダ） |
| 注意 | `userId` / `emailHash` は別 middleware (LC-AUTH-01) が後で書き込む |
| 連携 | [LC-AUTH-05 ContextAwareSlogHandler](#lc-auth-05-contextawareslogahandler) |

### LC-AUTH-05: `ContextAwareSlogHandler`

| 項目 | 内容 |
|---|---|
| 配置 | `back/api/internal/logging/handler.go`（仮） |
| 公開 IF | `New(w io.Writer, level slog.Level) slog.Handler` |
| 振る舞い | `Handle(ctx, record)` で context から下記キーを抽出 → record に attrs 追加:<br>- `requestId`, `traceId`, `userId`, `userAgent`, `emailHash`, `action`<br>内部実装は `slog.NewJSONHandler` に委譲 |
| 出力先 | `os.Stdout`（CloudWatch Logs に流れる） |
| テスト | P-TEST-02 で `bytes.Buffer` 差し替えにより JSON 内容を assert |

### LC-AUTH-06: `LogoutHandler`

| 項目 | 内容 |
|---|---|
| 配置 | `back/api/internal/handlers/logout.go`（仮） |
| 公開 IF | `Logout(c *gin.Context)` |
| 振る舞い | `userId` を context から取得 → `slog.InfoContext(ctx, "user logout", "action", "logout")` → `c.Status(204)` |
| 認証 | 必須（middleware で済んでいる前提） |
| Cognito 連携 | サーバ側では GlobalSignOut を呼ばない（クライアント側で Amplify Auth.signOut が呼出済） |

### LC-AUTH-07: `PreSignUpTriggerLambda`

| 項目 | 内容 |
|---|---|
| 配置 | `infra/lambdas/pre-signup/index.js`（仮） |
| ランタイム | Node.js 20.x |
| 振る舞い | `event.response.autoConfirmUser = true` / `autoVerifyEmail = true` を設定して event を返す |
| 失敗時 | 例外をそのまま伝播（DLQ なし、追加ログなし、A-NFR-REL-01 / P-RES-03） |
| デプロイ | Terraform `aws_lambda_function` + `archive_file` で zip 化、Cognito User Pool LambdaConfig.PreSignUp に紐付け |

---

## 3. Frontend 論理コンポーネント

### LC-AUTH-08: `useAuthHook`

| 項目 | 内容 |
|---|---|
| 配置 | `web/hooks/useAuth.ts`（仮） |
| 公開 IF | [unit-interfaces.md §9](../../interfaces/unit-interfaces.md) と一致:<br>`{ user, login(), signup(), logout(), isAuthenticated }` |
| 内部状態 | `status: "loading" \| "authenticated" \| "unauthenticated"` |
| 振る舞い | `login()`/`signup()`/`logout()` 内で Amplify SDK 関数 (`Auth.signIn` / `signUp` / `signOut`) を呼ぶ。エラーは `AuthErrorWithCode` に正規化して throw |
| 連携 | [LC-AUTH-13 AuthMessagesResource](#lc-auth-13-authmessagesresource), [LC-AUTH-14 AuthHubListener](#lc-auth-14-authhublistener) |

### LC-AUTH-09: `apiClient`

| 項目 | 内容 |
|---|---|
| 配置 | `web/lib/apiClient.ts`（仮、Browser 側のみ） |
| 公開 IF | `request(input: RequestInit & { path: string }): Promise<Response>` |
| 振る舞い (BFF パターン) | 1. `fetchAuthSession()` で **AccessToken** を取得（IdToken ではない、OAuth2 ベストプラクティス）<br>2. `fetch(input.path, { headers: { Authorization: 'Bearer <accessToken>', ...input.headers } })` で同一オリジンの `/api/*` を叩く<br>3. レスポンス 401 → `triggerSessionExpired()`（[LC-AUTH-10](#lc-auth-10-sessionexpiredatom)）+ `AuthErrorWithCode('SESSION_EXPIRED')` throw<br>4. 429 → `AuthErrorWithCode('RATE_LIMIT_EXCEEDED')` throw（P-RES-04）<br>5. 5xx / NetworkError → `AuthErrorWithCode('NETWORK_ERROR')` throw |
| 連携 | [LC-AUTH-10 sessionExpiredAtom](#lc-auth-10-sessionexpiredatom), [LC-AUTH-13 AuthMessagesResource](#lc-auth-13-authmessagesresource), [LC-AUTH-18 BffProxyRouteHandler](#lc-auth-18-bffproxyroutehandler) |
| 注意 | `Authorization: Bearer <accessToken>` ヘッダを付与（Server 側 LC-18 はこれを透過するだけ）。同一オリジンのため CORS 不要 |

### LC-AUTH-10: `sessionExpiredAtom`

| 項目 | 内容 |
|---|---|
| 配置 | `web/state/auth.ts`（仮） |
| 型 | `atom<{ openedAt: number } | null>(null)` |
| 振る舞い | `triggerSessionExpired()` ヘルパが `getSnapshot() === null` のときだけ書き込み → 二重発火防止（P-RES-02） |
| Reset | `setSessionExpired(null)` で明示クリア（モーダル閉じた後 / 遷移完了後） |
| Subscribe | [LC-AUTH-11 SessionExpiredModalHost](#lc-auth-11-sessionexpiredmodalhost) |

### LC-AUTH-11: `SessionExpiredModalHost`

| 項目 | 内容 |
|---|---|
| 配置 | `web/components/auth/SessionExpiredModalHost.tsx`（仮） |
| 振る舞い | `useAtomValue(sessionExpiredAtom)` で監視<br>非 null のとき Modal 表示 + `useEffect` で `setTimeout(1500ms)` → atom クリア + `router.push('/login?from=session_expired')`<br>OK ボタン押下時は即時実行（`clearTimeout` + 同処理） |
| ライフサイクル | アプリ全体に 1 つだけマウント (`app/layout.tsx`) |

### LC-AUTH-12: `AuthGuard`

| 項目 | 内容 |
|---|---|
| 配置 | `web/components/auth/AuthGuard.tsx`（仮） |
| 適用範囲 | `app/(authenticated)/layout.tsx` |
| 振る舞い | `useAuth().status` を監視<br>- `"loading"` で `showLoader = false` (300ms 経過後 `true`、P-PERF-01)<br>- `"loading" && !showLoader` → null<br>- `"loading" && showLoader` → `<FullScreenLoader />`<br>- `"authenticated"` → children<br>- `"unauthenticated"` → `useEffect` で `router.replace('/login')`、本体は null |

### LC-AUTH-13: `AuthMessagesResource`

| 項目 | 内容 |
|---|---|
| 配置 | `web/lib/authMessages.ts`（仮） |
| 公開 IF | `authMessages: Record<AuthErrorCode, string>` |
| 内容 | Functional Design [frontend-components.md §11](../functional-design/frontend-components.md) に既定義の 7 メッセージ + `SESSION_EXPIRED` エントリ追加 |
| 注意 | Q-A8=B（ダメ化トーン軽）の文言を保持 |

### LC-AUTH-14: `AuthHubListener`

| 項目 | 内容 |
|---|---|
| 配置 | `web/lib/authHubListener.ts`（仮） |
| 振る舞い | アプリ起動時に `Hub.listen('auth', handler)` で購読<br>`tokenRefresh_failure` イベント発火時 → `triggerSessionExpired()`（[LC-AUTH-10](#lc-auth-10-sessionexpiredatom)）<br>`signedIn` / `signedOut` イベントは `useAuth` 側の状態同期に使用 |
| ライフサイクル | アプリ全体に 1 つだけセットアップ (`app/layout.tsx` 内の AppProviders) |

---

## 4. AWS 論理コンポーネント（パラメータ表現）

実 Terraform HCL は Infrastructure Design で確定。本書では論理パラメータのみ。

### LC-AUTH-15: `CognitoUserPoolConfig`

Q-D9=B「A + 推奨項目」に従う。

| カテゴリ | パラメータ | 値 |
|---|---|---|
| 基本 | `UserPoolName` | `goro2pay-dev` 等（Infrastructure Design で確定） |
| 認証属性 | `UsernameAttributes` | `["email"]` |
| 自動検証 | `AutoVerifiedAttributes` | `["email"]`（Pre Sign-up Trigger との合わせ技） |
| パスワード | `PasswordPolicy.MinimumLength` | 8 |
| パスワード | `PasswordPolicy.RequireUppercase` | true |
| パスワード | `PasswordPolicy.RequireLowercase` | true |
| パスワード | `PasswordPolicy.RequireNumbers` | true |
| パスワード | `PasswordPolicy.RequireSymbols` | false |
| MFA | `MfaConfiguration` | `OFF` |
| アカウント回復 | `AccountRecoverySetting` | 本 MVP ではパスワードリセット未実装、最小設定 |
| 管理者作成 | `AdminCreateUserConfig.AllowAdminCreateUserOnly` | false（ユーザ自身が SignUp 可能） |
| Lambda Trigger | `LambdaConfig.PreSignUp` | LC-AUTH-07 の Lambda ARN |

### LC-AUTH-16: `CognitoAppClientConfig`

| カテゴリ | パラメータ | 値 |
|---|---|---|
| 基本 | `AppClientName` | `goro2pay-web` 等 |
| Auth Flow | `ExplicitAuthFlows` | `["ALLOW_USER_PASSWORD_AUTH", "ALLOW_REFRESH_TOKEN_AUTH"]` |
| Token 有効期限 | `IdTokenValidity` | **8** （単位: hours） |
| Token 有効期限 | `AccessTokenValidity` | **8** （単位: hours） |
| Token 有効期限 | `RefreshTokenValidity` | **30** （単位: days） |
| Token Validity Units | `TokenValidityUnits` | `{ AccessToken: "hours", IdToken: "hours", RefreshToken: "days" }` |
| Secret | `GenerateSecret` | false（PWA Public Client） |

### LC-AUTH-17: `ApiGatewayStageThrottlingConfig`

| パラメータ | 値 |
|---|---|
| `RateLimit` | 100 (req/s) |
| `BurstLimit` | 200 (req) |
| 適用範囲 | Stage 全体 |

### LC-AUTH-18: `BffProxyRouteHandler`

BFF パターン採用 (Q-I14/I15) に伴う新規論理コンポーネント。Next.js Server (Amplify SSR Compute) で動作。

| 項目 | 内容 |
|---|---|
| 配置 | `web/app/api/[...path]/route.ts`（catch-all proxy） |
| 公開 IF | Next.js Route Handler の規約に従い GET / POST / PUT / DELETE / PATCH の各エクスポート |
| 振る舞い | 1. `request.headers.get('Authorization')` で `Bearer ` プレフィックス確認（欠落時は 401 即返）<br>2. `process.env.API_ENDPOINT` (server-only) を読み出し<br>3. `params.path` から上流 URL を組み立て (`${API_ENDPOINT}/api/${path.join('/')}`)<br>4. **`Authorization` ヘッダを透過** して上流に fetch（変換しない、Browser からの AccessToken をそのまま渡す）<br>5. 上流レスポンスの status / body / headers をそのまま透過 (401 を含む全 status) |
| 環境変数 | `API_ENDPOINT` (server-only、`NEXT_PUBLIC_` プレフィックスなし) |
| 連携 | [LC-AUTH-09 apiClient](#lc-auth-09-apiclient) からの呼び出しを受ける、上流 [LC-AUTH-15 Cognito](#lc-auth-15-cognitouserpoolconfig) で認証された API Gateway / API Lambda へ proxy |
| 注意 | 個別 API ルート (`web/app/api/<feature>/route.ts`) を新設すると catch-all より優先される。本 Unit A では catch-all のみ実装し、各 Unit (B/C/D/E) も catch-all を再利用 |

**設計上のポイント**:
- `API_ENDPOINT` を Browser に露出しない（NEXT_PUBLIC_ なし）
- ブラウザ ↔ Next.js Server は同一オリジンなので CORS preflight 不要
- **Authorization ヘッダ透過**: Browser → Server → API Gateway の経路で `Bearer <accessToken>` をそのまま転送、Server 側で IdToken 等への変換責務を持たない（OAuth2 ベストプラクティス）
- 401 は透過する（Server 側で SESSION_EXPIRED 検出はしない、責務は Client 側 LC-AUTH-09）

---

## 5. コンポーネント関係図

```
┌──── Frontend (Next.js) ────────────────────────────────┐
│                                                          │
│  ┌──────────┐    ┌──────────────────┐                   │
│  │ useAuth  │ ── │ AuthMessages     │                   │
│  │ (LC-08)  │    │ (LC-13)          │                   │
│  └─┬────┬───┘    └──────────────────┘                   │
│    │    │                                                │
│    │    │  signup/login/logout                          │
│    │    ▼                                                │
│    │  ┌──────────┐    ┌─────────────────────────┐       │
│    │  │ apiClient│ ── │ Authorization:          │       │
│    │  │ (LC-09)  │    │ Bearer <accessToken>    │       │
│    │  │ Browser  │    │ (Browser 側付与)        │       │
│    │  └─┬────────┘    └─────────────────────────┘       │
│    │    │ fetch /api/* (同一オリジン、CORS 不要)         │
│    │    │ 401/429                                       │
│    │    ▼                                                │
│    │  ┌──────────────────┐                              │
│    │  │ sessionExpired   │ ◄─── triggerSessionExpired() │
│    │  │ Atom (LC-10)     │      from AuthHubListener    │
│    │  └─┬────────────────┘      (LC-14)                 │
│    │    │ subscribe                                     │
│    │    ▼                                                │
│    │  ┌──────────────────────┐                          │
│    │  │ SessionExpiredModal  │                          │
│    │  │ Host (LC-11)         │                          │
│    │  └──────────────────────┘                          │
│    │                                                     │
│    │  ┌──────────────────┐                              │
│    └▶ │ AuthGuard (LC-12)│                              │
│       │ (300ms Loading)  │                              │
│       └──────────────────┘                              │
└──────────────────────────────────────────────────────────┘
       │ Browser → /api/* (同一オリジン)
       ▼
┌── Next.js Server (Amplify Hosting SSR) ──────────────────┐
│                                                            │
│  ┌─────────────────────────────────────────────────────┐  │
│  │ BffProxyRouteHandler (LC-18)                        │  │
│  │  /api/[...path]/route.ts (catch-all)                │  │
│  │  - Authorization: Bearer <accessToken> を透過       │  │
│  │  - server-only env API_ENDPOINT を読み出し          │  │
│  │  - 上流 API Gateway に proxy                        │  │
│  └─────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────┘
                       │ HTTPS (server-to-server)
                       ▼
┌── API Gateway (Cognito Authorizer + Stage Throttling LC-17)
└──┬──────────────────────────────────────────────────────┘
   │ event.requestContext.authorizer.claims
   ▼
┌── Lambda (Go + Gin + LWA) ───────────────────────────────┐
│                                                            │
│  ┌─────────────────────────┐                              │
│  │ RequestContextMiddleware│ ── set requestId/traceId/UA │
│  │ (LC-04)                 │                              │
│  └────────┬────────────────┘                              │
│           ▼                                                │
│  ┌─────────────────────────┐    ┌──────────────────┐     │
│  │ AttachUserIDMiddleware  │ ── │ EmailHasher      │     │
│  │ (LC-01)                 │ ── │ (LC-03)          │     │
│  └────────┬────────────────┘    └──────────────────┘     │
│           │                       ▲                       │
│           │                       │ uses Normalize        │
│           │                       │                       │
│           │                       ┌──────────────────┐    │
│           │                       │ EmailNormalizer  │    │
│           │                       │ (LC-02)          │    │
│           │                       └──────────────────┘    │
│           ▼                                                │
│  ┌─────────────────────────┐                              │
│  │ LogoutHandler (LC-06)   │                              │
│  │ + 他 Handler (Unit B/C/.)│                              │
│  └────────┬────────────────┘                              │
│           │ slog.InfoContext(...)                         │
│           ▼                                                │
│  ┌─────────────────────────┐                              │
│  │ ContextAwareSlogHandler │ ── stdout → CloudWatch       │
│  │ (LC-05)                 │                              │
│  └─────────────────────────┘                              │
└────────────────────────────────────────────────────────────┘

┌── Cognito User Pool (LC-15) ─────────────────────────────┐
│                                                            │
│  ┌─────────────────────────┐                              │
│  │ Pre Sign-up Trigger     │ ◄── invoked on SignUp        │
│  │ Lambda (LC-07, Node.js) │     auto-confirm + verify    │
│  └─────────────────────────┘                              │
│                                                            │
│  ┌─────────────────────────┐                              │
│  │ App Client (LC-16)      │ ◄── used by Frontend Amplify │
│  │ Token Validity 8h/30d   │                              │
│  └─────────────────────────┘                              │
└────────────────────────────────────────────────────────────┘
```

---

## 6. コンポーネント → パターン対応

| LC ID | 関連パターン |
|---|---|
| LC-AUTH-01 | P-OBS-01, P-SEC-02 |
| LC-AUTH-02 | P-SEC-02, P-TEST-01 |
| LC-AUTH-03 | P-SEC-02, P-TEST-01 |
| LC-AUTH-04 | P-OBS-01 |
| LC-AUTH-05 | P-OBS-01, P-TEST-02 |
| LC-AUTH-06 | P-OBS-01 |
| LC-AUTH-07 | P-RES-03 |
| LC-AUTH-08 | P-RES-01, P-SEC-01 |
| LC-AUTH-09 | P-RES-01, P-RES-04 |
| LC-AUTH-10 | P-RES-02 |
| LC-AUTH-11 | P-RES-02, P-PERF-01 (派生) |
| LC-AUTH-12 | P-PERF-01 |
| LC-AUTH-13 | P-RES-01, P-RES-04 |
| LC-AUTH-14 | P-RES-02 |
| LC-AUTH-15 | P-SEC-04 (パラメータ整合性のみ) |
| LC-AUTH-16 | P-SEC-03 |
| LC-AUTH-17 | P-SEC-04 |
| LC-AUTH-18 | P-RES-01 (BFF パターン Server 側), P-SEC-04 (API URL 秘匿化) |

---

## 7. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| **Infrastructure Design** | LC-15/16/17 を Terraform モジュール (`infra/modules/cognito/`, `infra/modules/api_gateway/`) に展開、LC-07 Lambda の `archive_file` 設定、Cognito Authorizer の API Gateway 統合 |
| **Code Generation** | LC-01〜LC-14 の実装本体、`go.mod` / `package.json` の version pin、PBT テストコード、E2E テストコード |
