# Auth Unit — Business Logic Model

**Document Version**: 1.2
**Created**: 2026-05-21
**Updated**: 2026-05-22 (BFF パターン採用: F-3 / F-4 を Next.js Server 経由に書き換え、ブラウザ → Amplify SSR → API Gateway の経路に統一)
**Updated**: 2026-05-22 (API 認証 Token を IdToken → AccessToken に統一、ヘッダを Authorization: Bearer に統一、OAuth2 ベストプラクティス整合)
**Unit**: A (`auth`)
**Stage**: Functional Design / Construction

本ドキュメントは Unit A の **業務フロー** をシーケンス図とアルゴリズム記述で示す。技術スタックの参照は最小限とし、フローのロジックそのものに焦点を当てる。

参照: [domain-entities.md](./domain-entities.md), [business-rules.md](./business-rules.md)

---

## 1. 全体フロー俯瞰

```
            ┌─────────────────┐
            │  LandingScreen  │ (未認証 / "/")
            └────────┬────────┘
                     │
        ┌────────────┴────────────┐
        ▼                         ▼
 ┌─────────────┐          ┌─────────────┐
 │ SignupScreen│          │ LoginScreen │
 └──────┬──────┘          └──────┬──────┘
        │                        │
        │  Signup → 自動 SignIn   │  SignIn
        ▼                        ▼
        └────────┬───────────────┘
                 ▼
        ┌────────────────┐
        │  AuthSession   │ 確立
        │  router.push("/")
        └────────┬───────┘
                 ▼
       ┌──────────────────────────┐
       │  / (MainScreen, Unit B/C)│
       │  予算未設定なら /budget へ │
       └──────────────────────────┘

            （セッション中）
                 │
                 │ 401 検出 or
                 │ tokenRefresh_failure
                 ▼
       ┌──────────────────┐
       │  Session Expired │
       │  Modal → /login  │
       └──────────────────┘

       「ログアウト」 ボタン押下
                 │
                 ▼
       ┌──────────────────────────┐
       │  GlobalSignOut →         │
       │  POST /api/auth/logout → │
       │  router.push("/")        │
       └──────────────────────────┘
```

---

## 2. Flow F-1: 新規登録（Signup）

### 2.1 シーケンス図

```mermaid
sequenceDiagram
    autonumber
    actor User as 太郎
    participant UI as SignupScreen<br/>(Next.js)
    participant Auth as Amplify Auth<br/>(@aws-amplify/auth)
    participant Cog as Cognito User Pool
    participant PreSU as Pre Sign-up<br/>Lambda Trigger

    User->>UI: メール + パスワード入力
    UI->>UI: クライアント側検証<br/>(R-Email-1, R-Pwd-1)
    User->>UI: 「登録」ボタン押下
    UI->>Auth: signUp({ username, password, options.userAttributes.email })
    Auth->>Cog: SignUp API
    Cog->>PreSU: PreSignUp_SignUp イベント
    PreSU-->>Cog: autoConfirmUser=true,<br/>autoVerifyEmail=true
    Cog-->>Auth: {userConfirmed: true, userId: <sub>}
    Auth-->>UI: signUp 成功

    Note over UI,Cog: そのまま自動ログインへ

    UI->>Auth: signIn({ username, password })
    Auth->>Cog: InitiateAuth (USER_SRP_AUTH)
    Cog-->>Auth: {idToken, accessToken, refreshToken}
    Auth-->>UI: AuthSession 確立
    UI->>UI: router.push("/")

    Note over UI: パスワード state を即クリア (R-Pwd-3-c)
```

### 2.2 アルゴリズム（疑似コード）

```pseudo
function handleSignup(email: string, password: string):
    # クライアント側事前検証
    assert isValidEmailFormat(email)        # R-Email-1
    assert meetsPasswordPolicy(password)    # R-Pwd-1

    normalizedEmail = email.toLowerCase()   # R-Email-3

    try:
        signUpResult = AmplifyAuth.signUp({
            username: normalizedEmail,
            password: password,
            options: { userAttributes: { email: normalizedEmail } }
        })
        # signUpResult.userConfirmed === true (auto-confirm 効いている)

        # 自動ログイン
        AmplifyAuth.signIn({ username: normalizedEmail, password: password })
        # Amplify が AuthSession を確立し localStorage に保存

        # メモリからパスワードを除去
        clearPasswordState()                # R-Pwd-3-c

        router.push("/")                    # R-Signup-3
    except UsernameExistsException:
        showError(EMAIL_ALREADY_EXISTS)
    except InvalidPasswordException:
        showError(WEAK_PASSWORD)
    except NetworkError:
        showError(NETWORK_ERROR)
    except others:
        showError(UNKNOWN)
        logErrorWithoutPassword(error)      # R-Pwd-3-b
```

### 2.3 例外分岐

| 例外 | 表示エラー | 後続動作 |
|---|---|---|
| `UsernameExistsException` | EMAIL_ALREADY_EXISTS | フォームに留まる、ログインへの導線提示 |
| `InvalidPasswordException` | WEAK_PASSWORD | フォームに留まる、ヒント表示 |
| `InvalidParameterException` | INVALID_EMAIL_FORMAT | フォームに留まる |
| Network error | NETWORK_ERROR | リトライボタン表示 |
| その他 | UNKNOWN | フォームに留まる、エラー報告ヒント |

---

## 3. Flow F-2: ログイン（Login）

### 3.1 シーケンス図

```mermaid
sequenceDiagram
    autonumber
    actor User as 太郎
    participant UI as LoginScreen
    participant Auth as Amplify Auth
    participant Cog as Cognito User Pool

    User->>UI: メール + パスワード入力
    UI->>UI: クライアント側検証
    User->>UI: 「ログイン」ボタン押下
    UI->>Auth: signIn({ username, password })
    Auth->>Cog: InitiateAuth (USER_SRP_AUTH)

    alt 認証成功
        Cog-->>Auth: {idToken, accessToken, refreshToken}
        Auth-->>UI: AuthSession 確立
        UI->>UI: clearPasswordState()
        UI->>UI: router.push("/")
    else NotAuthorizedException
        Cog-->>Auth: 401
        Auth-->>UI: error
        UI->>UI: showError(INVALID_CREDENTIALS)
    else UserNotFoundException
        Cog-->>Auth: 404
        Auth-->>UI: error
        UI->>UI: showError(INVALID_CREDENTIALS)<br/>(同一文言、R-Err-1)
    else TooManyRequestsException
        Cog-->>Auth: 429
        Auth-->>UI: error
        UI->>UI: showError(RATE_LIMIT_EXCEEDED)
    end
```

### 3.2 アルゴリズム

```pseudo
function handleLogin(email: string, password: string):
    assert isValidEmailFormat(email)
    normalizedEmail = email.toLowerCase()

    try:
        AmplifyAuth.signIn({ username: normalizedEmail, password: password })
        clearPasswordState()
        router.push("/")
    except NotAuthorizedException, UserNotFoundException:
        showError(INVALID_CREDENTIALS)      # R-Err-1, R-Login-2
    except TooManyRequestsException:
        showError(RATE_LIMIT_EXCEEDED)
    except NetworkError:
        showError(NETWORK_ERROR)
```

---

## 4. Flow F-3: 認証付き API 呼出（BFF 経由）

### 4.1 BFF パターンの基本方針

Amplify SSR (Next.js App Router) を採用しているため、**ブラウザは API Gateway を直接呼ばず、Next.js Server 経由（BFF）で API Gateway を呼び出す**。

主なメリット:
- API Gateway URL を **server-only env (`API_ENDPOINT`)** で秘匿化（ブラウザバンドルに露出しない）
- CORS の `allow_origins` を Amplify ドメインだけに絞れる（より厳格）
- Server Side Rendering の旨味を活かせる
- 認証ヘッダ付与を Server で一元化

### 4.2 シーケンス図

```mermaid
sequenceDiagram
    autonumber
    actor User as 太郎
    participant Browser as Browser<br/>(Client Component)
    participant NextSrv as Next.js Server<br/>(Route Handler /<br/> Server Action)
    participant Auth as Amplify Auth<br/>(@aws-amplify/auth)
    participant APIGW as API Gateway<br/>+ JWT Authorizer
    participant LMD as API Lambda<br/>(Gin + LWA)
    participant MW as AuthContext<br/>middleware
    participant H as Handler

    User->>Browser: アクション (例: ログアウト確定)
    Browser->>Auth: fetchAuthSession()
    Auth-->>Browser: accessToken (有効期限内なら同じ、切れていれば自動 refresh)

    Browser->>NextSrv: POST /api/auth/logout<br/>Authorization: Bearer <accessToken>
    Note over NextSrv: catch-all Route Handler<br/>(/api/[...path]/route.ts)<br/>server-only env API_ENDPOINT<br/>を読み出し、Authorization は透過

    NextSrv->>APIGW: POST {API_ENDPOINT}/api/auth/logout<br/>Authorization: Bearer <accessToken>
    APIGW->>APIGW: 1. 署名検証 (JWKS)<br/>2. exp / aud / iss 検証

    alt Authorizer 成功
        APIGW->>LMD: invoke<br/>event.requestContext.authorizer.claims
        LMD->>MW: AttachUserID()
        MW->>MW: claims["sub"] を取り出す
        MW->>MW: c.Set("userId", sub)
        MW->>H: c.Next()
        H->>H: UserIDFromContext(c) → sub
        H-->>LMD: 200 / レスポンス
        LMD-->>APIGW: response
        APIGW-->>NextSrv: 200
        NextSrv-->>Browser: 200 (透過 or 必要なら整形)
    else Authorizer 失敗 (期限切れ / 改竄)
        APIGW-->>NextSrv: 401 Unauthorized
        NextSrv-->>Browser: 401 Unauthorized<br/>(透過、Client が SESSION_EXPIRED 検出)
        Browser->>Browser: triggerSessionExpired()<br/>(Client 側 P-RES-02 atom)
    end
```

### 4.3 middleware 疑似コード（API Lambda 側、変更なし）

API Gateway Authorizer の検証は変わらないため、middleware ロジックは BFF 採用前と同一。

```pseudo
function AttachUserID() handler:
    return func(c *gin.Context):
        claims = c.Request.Context["authorizer.claims"]
        if claims == nil or claims["sub"] == "":
            log.error("missing claims, possible misconfiguration")
            c.AbortWithStatus(500)               # R-JWT-4
            return
        c.Set("userId", claims["sub"])
        c.Set("userEmail", claims["email"])      # 監査用、永続化禁止
        c.Next()

function UserIDFromContext(c) -> (userId, error):
    v = c.Get("userId")
    if v == nil:
        return "", ErrUnauthorized
    return v.(string), nil
```

### 4.4 Next.js catch-all Route Handler 疑似コード

ブラウザのリクエストパスは **既存の `/api/*` のまま**（unit-interfaces.md §3.3 と整合）。Next.js の `app/api/[...path]/route.ts` で全 `/api/*` を受けて上流 API Gateway に転送する。

```pseudo
# web/app/api/[...path]/route.ts (catch-all proxy)
export async function ANY_METHOD(request, { params: { path } }):
    const apiEndpoint = process.env.API_ENDPOINT  # server-only
    const authHeader = request.headers.get("Authorization")
    if not authHeader or not authHeader.startsWith("Bearer "):
        return new Response("Missing Authorization header", { status: 401 })

    const upstreamUrl = `${apiEndpoint}/api/${path.join('/')}`
    const upstream = await fetch(upstreamUrl, {
        method: request.method,
        headers: { Authorization: authHeader },  # 透過、Server 側で変換しない
        body: request.method != "GET" ? await request.text() : undefined,
    })

    return new Response(upstream.body, {
        status: upstream.status,
        headers: upstream.headers,
    })
```

このパターンは Logout 専用ではなく、**全ての認証付き API** で同じ catch-all が動作する。`Authorization: Bearer <accessToken>` は Browser → Server → API Gateway を **透過する**（Server で変換しない）。個別整形が必要な API は `app/api/<feature>/route.ts` を個別作成して上書き可能。

---

## 5. Flow F-4: ログアウト（Logout）BFF 経由

### 5.1 シーケンス図

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant UI as Header / Settings<br/>(Client Component)
    participant Auth as Amplify Auth
    participant Cog as Cognito
    participant NextSrv as Next.js Server<br/>(Route Handler)
    participant APIGW as API Gateway
    participant LMD as API Lambda

    User->>UI: 「ログアウト」ボタン押下
    UI->>UI: 確認モーダル表示<br/>(R-Logout-4)
    User->>UI: 「ログアウト」確定

    UI->>Auth: fetchAuthSession()
    Auth-->>UI: accessToken (監査ログ送信用、signOut 前に取得)

    Note over UI: 監査ログ目的（任意、UI が成立すれば router.push へ進む）
    UI->>NextSrv: POST /api/auth/logout<br/>Authorization: Bearer <accessToken>
    Note over NextSrv: catch-all Route Handler が proxy<br/>(Authorization は透過)
    NextSrv->>APIGW: POST {API_ENDPOINT}/api/auth/logout<br/>Authorization: Bearer <accessToken>
    APIGW->>LMD: invoke
    LMD->>LMD: 構造化ログに userId, action=logout 記録
    LMD-->>APIGW: 204
    APIGW-->>NextSrv: 204
    NextSrv-->>UI: 204

    Note over UI: バックエンド監査ログ完了後、Cognito GlobalSignOut
    UI->>Auth: signOut({ global: true })
    Auth->>Cog: GlobalSignOut(accessToken)
    Cog-->>Auth: 200
    Auth->>Auth: localStorage からトークン除去
    Auth-->>UI: 完了

    UI->>UI: router.push("/")<br/>(LandingScreen が描画)
```

### 5.2 サーバ側の最小実装（API Lambda 側）

```pseudo
# API Lambda: POST /api/auth/logout (BFF 経由で受信)
function handleLogout(c *gin.Context):
    userId, _ = UserIDFromContext(c)
    log.info("user logout", "userId", userId, "at", time.Now())
    c.Status(204)
    return
```

サーバ側で GlobalSignOut を再実行しない（クライアント側で既に呼出済）。

---

## 6. Flow F-5: セッション失効（Session Expiry）

### 6.1 シーケンス図

```mermaid
sequenceDiagram
    autonumber
    participant UI as Browser<br/>(Client Component)
    participant Auth as Amplify Auth
    participant NextSrv as Next.js Server<br/>(Route Handler)
    participant APIGW as API Gateway

    Note over UI,APIGW: パターン 1: BFF (Next.js catch-all proxy) 経由 API 呼出時の 401

    UI->>NextSrv: POST /api/xxx<br/>Authorization: Bearer <expired-accessToken>
    NextSrv->>APIGW: 上流に透過転送 (Authorization: Bearer <expired-accessToken>)
    APIGW-->>NextSrv: 401 Unauthorized
    NextSrv-->>UI: 401 Unauthorized (透過)
    UI->>UI: 401 検出 (Client interceptor)
    UI->>UI: triggerSessionExpired()

    Note over UI,Auth: パターン 2: 自動 refresh 失敗

    UI->>Auth: fetchAuthSession()
    Auth->>Auth: refreshToken も期限切れ
    Auth-->>UI: tokenRefresh_failure イベント
    UI->>UI: triggerSessionExpired()

    Note over UI: 共通動線
    UI->>UI: モーダル表示「お疲れ様でした…」
    UI->>UI: 1.5 秒後 / OK タップ
    UI->>UI: router.push("/login?from=session_expired")
```

### 6.2 アルゴリズム

#### 6.2.1 クライアント側 HTTP interceptor（apiClient）

ブラウザ → Next.js BFF へのリクエスト返却を監視。

```pseudo
function apiClient.onResponse(response):
    if response.status == 401:
        triggerSessionExpired()
    return response

function setupAmplifyAuthHub():
    Amplify.Hub.listen("auth", (event) => {
        if event.payload.event == "tokenRefresh_failure":
            triggerSessionExpired()
    })

function triggerSessionExpired():
    if alreadyHandling: return                 # 二重発火防止 (P-RES-02)
    alreadyHandling = true
    closeAllOpenModals()
    showModal(SESSION_EXPIRED_MESSAGE)
    setTimeout(() => router.push("/login?from=session_expired"), 1500)
```

#### 6.2.2 サーバ側 BFF Route Handler

Next.js Server から API Gateway への転送。401 を透過して Client 側で処理させる。

```pseudo
function bffProxy(request, upstreamPath):
    const authHeader = request.headers.get("Authorization")
    if not authHeader or not authHeader.startsWith("Bearer "):
        return new Response("Missing Authorization header", { status: 401 })

    const apiEndpoint = process.env.API_ENDPOINT  # server-only
    const upstream = await fetch(`${apiEndpoint}${upstreamPath}`, {
        method: request.method,
        headers: { Authorization: authHeader, "Content-Type": "application/json" },  # 透過
        body: request.method !== "GET" ? await request.text() : undefined,
    })

    # 401 はそのまま透過、Client 側 P-RES-02 atom が拾う
    return new Response(upstream.body, {
        status: upstream.status,
        headers: upstream.headers,
    })
```

---

## 7. Flow F-6: 未認証アクセスのガード

### 7.1 シーケンス図

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant URL as Browser URL
    participant Layout as AuthGuard Layout
    participant Auth as Amplify Auth

    User->>URL: 直接 /budget に GET
    URL->>Layout: render
    Layout->>Auth: getCurrentUser() / fetchAuthSession()
    alt 認証済み
        Auth-->>Layout: AuthSession
        Layout->>Layout: children を描画
    else 未認証
        Auth-->>Layout: NotAuthenticatedError
        Layout->>URL: router.replace("/login")
    end
```

### 7.2 ルートポリシー

[business-rules.md §8 R-Unauth-1](./business-rules.md) に基づき、ルートを 2 グループに分ける:

- **Public**: `/`, `/login`, `/(auth)/signup` — `<AuthGuard>` で囲まない
- **Authenticated**: `/budget`, `/budget-empty`, `/order/*` 等 — Next.js の `app/(authenticated)/` レイアウトで `<AuthGuard>` を適用

`/` だけは特殊で、Public だが認証状態に応じて `<LandingScreen>` か `<MainScreen>` を切り替える。

---

## 8. データフロー（BFF パターン、Unit 横断）

```
┌──────────┐  /api/* リクエスト   ┌─────────────────┐    Authorization     ┌──────────┐  claims抽出   ┌────────┐
│ Browser  ├────────────────────►│ Next.js Server  ├────────────────────►│API Gateway├─────────────►│ Lambda │
│ (Client) │   Authorization:    │ (Amplify SSR)   │   Bearer 透過        │+JWT Auth  │              │        │
│          │   Bearer <accessTok>│                 │                      └──────────┘              │ Gin    │
│ Amplify  │                     │ env.API_ENDPOINT│                                                 │+Auth MW│
│ Auth     │                     │ (server-only)   │                                                 │        │
│          │   AuthSession       │                 │                                                 │ ┌────┐ │
│          │   ┌─────────────┐   │  catch-all      │                                                 │ │user│ │
│  user    │   │ idToken /   │   │  Route Handler  │                                                 │ │Id  │ │
│  state   │   │ access /    │   │  /api/[...path] │                                                 │ │ctx │ │
│          │   │ refresh tk  │   │  /route.ts      │                                                 │ └─┬──┘ │
│          │   └─────────────┘   │                 │                                                 │   │    │
│          │   localStorage      │ Authorization 透過                                                │   ▼    │
│          │                     │ (Server で変換しない)                                              │ Unit B/│
│          │                     │                 │                                                 │ C/D/E  │
└──────────┘                     └─────────────────┘                                                 │ ハンド │
                                                                                                     │ ラへ   │
                                                                                                     └────────┘

ポイント:
1. ブラウザは API Gateway URL を一切知らない (NEXT_PUBLIC_* に API_ENDPOINT を含めない)
2. `API_ENDPOINT` は Next.js Server (Amplify SSR Compute) のみが知る server-only env
3. CORS は Amplify ドメインのみに絞れる (allow_origins = "*" でなくよい)、本 MVP は CORS 設定そのものを未設定
4. **API 認証は AccessToken を使用** (OAuth2 ベストプラクティス、PII を含まない)
5. **`Authorization: Bearer <accessToken>`** ヘッダは Browser → Server → API Gateway を透過 (Server で変換しない)
6. UserIdentity の永続化は Cognito User Pool のみ。アプリ DB には userId のみ保存される
```

---

## 9. 受入基準（US-0-01 / US-0-02）の検証マッピング

| 受入基準 | 検証ポイント | フロー |
|---|---|---|
| US-0-01: Cognito にユーザが登録される | Pre Sign-up Trigger 経由で `userConfirmed=true` になる | F-1 |
| US-0-01: ログイン済み状態でメイン画面へ遷移 | F-1 末尾の自動 SignIn → `router.push("/")` | F-1 |
| US-0-01: 総タップ数 5 回以下 | R-Signup-4 で定義、実装は UI 設計と E2E テストで検証 | F-1 |
| US-0-02: 認証され JWT を受け取る | F-2 シーケンス図 step 8 の Cognito レスポンス | F-2 |
| US-0-02: メイン画面へ遷移 | F-2 末尾 `router.push("/")` | F-2 |
| US-0-02: セッション 1 時間以上維持 | Token Validity (Infrastructure Design で確定、目標 1 時間以上) | NFR Design 領域 |
| US-0-02: パスワード誤りでエラー表示 | F-2 alt 分岐 `INVALID_CREDENTIALS` | F-2 |

---

## 10. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| NFR Requirements | F-2 の応答時間目標、F-3 middleware の処理オーバーヘッド許容値 |
| NFR Design | リトライ戦略（NetworkError 時）、HTTP interceptor の実装パターン |
| Infrastructure Design | Pre Sign-up Lambda Trigger の関数定義、Cognito User Pool 全パラメータ |
| Code Generation | F-1〜F-6 の実装、`AuthGuard` コンポーネント、`useAuth` フック、Gin middleware |
