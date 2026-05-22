# Auth Unit — Frontend Components

**Document Version**: 1.0
**Created**: 2026-05-21
**Unit**: A (`auth`)
**Stage**: Functional Design / Construction

本ドキュメントは Unit A の **フロントエンドコンポーネント構造、Props/State、ユーザインタラクション、API 連携** を定義する。スタイリング詳細・ピクセル指定は Code Generation で確定する。

参照: [business-logic-model.md](./business-logic-model.md), [business-rules.md](./business-rules.md), [domain-entities.md](./domain-entities.md)

---

## 1. コンポーネント階層

```
app/
├── layout.tsx                       # 横串 (AmplifyConfig + Providers)
├── page.tsx                         # / : 認証状態で UI 切替 (Public)
│   ├── <LandingScreen>              # 未認証時に描画
│   └── <MainScreen>                 # 認証時に描画 (Unit C 主体)
├── login/page.tsx                   # /login (Public)
│   └── <LoginScreen>
├── (auth)/signup/page.tsx           # /signup (Public)
│   └── <SignupScreen>
└── (authenticated)/                 # 認証必須グループ
    ├── layout.tsx                   # <AuthGuard> 適用
    ├── budget/page.tsx              # Unit B
    ├── budget-empty/page.tsx        # Unit E
    └── order/[id]/complete/page.tsx # Unit C

hooks/
└── useAuth.ts                       # Unit A の中核フック

components/auth/
├── AuthGuard.tsx                    # 認証必須レイアウト
├── LandingScreen.tsx                # ランディング (US-0-01)
├── LoginScreen.tsx                  # ログイン画面 (US-0-02)
├── SignupScreen.tsx                 # 登録画面 (US-0-01)
├── LogoutButton.tsx                 # ヘッダ/設定で利用
├── LogoutConfirmModal.tsx           # ログアウト確認
└── SessionExpiredModal.tsx          # セッション失効時のモーダル

lib/
├── amplifyConfig.ts                 # Amplify 設定 (env から読込)
├── apiClient.ts                     # 401 検出 interceptor を含む
└── authMessages.ts                  # ダメ化トーン軽のメッセージリソース
```

---

## 2. AuthProvider と global 初期化

### 2.1 `app/layout.tsx` の責務

| 責務 | 実装手段 |
|---|---|
| Amplify を 1 度だけ設定 | `Amplify.configure(amplifyConfig)` を Client Component で実行 |
| Auth Hub のリスナー登録 | `Hub.listen("auth", handler)` で `tokenRefresh_failure` 等を購読 |
| TanStack Query の Provider | `QueryClientProvider` |
| Jotai の Provider | `<JotaiProvider>` |
| グローバルモーダル領域 | `<SessionExpiredModalHost />` を mount |

```tsx
// 概念コード
"use client";
import { Amplify } from "aws-amplify";
import { Hub } from "aws-amplify/utils";
import { amplifyConfig } from "@/lib/amplifyConfig";

let configured = false;

export function AppProviders({ children }: { children: React.ReactNode }) {
  if (!configured) {
    Amplify.configure(amplifyConfig, { ssr: true });
    configured = true;
  }
  // Auth Hub listener は useEffect 内で 1 回登録
  return (
    <QueryClientProvider client={queryClient}>
      <JotaiProvider>
        <SessionExpiredModalHost />
        {children}
      </JotaiProvider>
    </QueryClientProvider>
  );
}
```

---

## 3. `useAuth` フック仕様

### 3.1 シグネチャ

```ts
type AuthStatus = "loading" | "authenticated" | "unauthenticated";

export interface UseAuthReturn {
  status: AuthStatus;
  user: { userId: string; email: string } | null;
  isAuthenticated: boolean;
  signup(email: string, password: string): Promise<void>;
  login(email: string, password: string): Promise<void>;
  logout(): Promise<void>;
}

export function useAuth(): UseAuthReturn;
```

> **メソッド名規約**: 公開メソッド名は [unit-interfaces.md §9](../../interfaces/unit-interfaces.md) の凍結契約に合わせ `signup` / `login` / `logout` を採用する。Amplify SDK 内部関数（`Auth.signUp` / `Auth.signIn` / `Auth.signOut`）はあくまで実装ディテールとして区別する。

### 3.2 内部実装方針

- `status` は内部 state、初期値 `"loading"`
- `useEffect` で `getCurrentUser()` を呼んで現在のセッションを判定
- Auth Hub の `signedIn` / `signedOut` / `tokenRefresh_failure` を購読し state を更新
- `signup` は内部で `Auth.signUp` → `Auth.signIn` の連鎖で実装（[business-logic-model.md F-1](./business-logic-model.md)）
- `logout` は `Auth.signOut({ global: true })` → `apiClient.post("/api/auth/logout")` → `router.push("/")`
- メソッドは Promise を返し、呼び出し側でエラーハンドリング可能

### 3.3 エラーハンドリング契約

`signup` / `login` は失敗時に `AuthErrorWithCode` 型を throw する:

```ts
export class AuthErrorWithCode extends Error {
  code: AuthErrorCode;  // domain-entities.md §6.2 と一対一対応
  constructor(code: AuthErrorCode, originalError?: unknown) { ... }
}

export type AuthErrorCode =
  | "INVALID_CREDENTIALS"
  | "EMAIL_ALREADY_EXISTS"
  | "WEAK_PASSWORD"
  | "INVALID_EMAIL_FORMAT"
  | "RATE_LIMIT_EXCEEDED"
  | "NETWORK_ERROR"
  | "UNKNOWN";
```

呼出側（LoginScreen / SignupScreen）はこの code を `authMessages.ts` に渡してメッセージを取得する。

---

## 4. `LandingScreen`

### 4.1 役割
未認証ユーザがアプリにアクセスした際の入口。「ゴロゴロPay」のブランディングと、登録/ログインへの導線を提供。

### 4.2 Props
なし（`/` ルート直下で `useAuth().status` を見て描画される）。

### 4.3 State
ローカル State なし（純粋な表示）。

### 4.4 UI 要素

| 要素 | 説明 |
|---|---|
| ロゴ + タグライン | 「ゴロゴロPay — めんどくさいを丸投げ」 |
| 「はじめる」ボタン | クリックで `/signup` へ |
| 「ログイン」リンク | クリックで `/login` へ |

### 4.5 タップ数（NFR-DEG-01）
ランディング画面のボタンタップ自体は US-0-01 の総タップ 5 回には含まないが、ボタン位置を画面中央 + 大きめにし「はじめる」一発で次画面へ誘導。

---

## 5. `SignupScreen`

### 5.1 役割
US-0-01 の主役。メールアドレス + パスワードを入力させ、Cognito へ登録 + 自動ログイン。

### 5.2 Props
なし（`/signup` ページから直接マウント）。

### 5.3 State

```ts
type SignupState = {
  email: string;
  password: string;
  pwdHints: {
    lengthOk: boolean;
    upperOk: boolean;
    lowerOk: boolean;
    digitOk: boolean;
  };
  isSubmitting: boolean;
  errorMessage: string | null;
};
```

### 5.4 検証ロジック（クライアント側）

| 項目 | ルール | 表示 |
|---|---|---|
| メール | `/^[^\s@]+@[^\s@]+\.[^\s@]+$/` | 不正時に「メール形式を確認してください」 |
| パスワード長 | `>= 8` | チェックリスト 4 項目をリアルタイムで切替 |
| 英大文字 | `/[A-Z]/` | 同上 |
| 英小文字 | `/[a-z]/` | 同上 |
| 数字 | `/\d/` | 同上 |

「登録」ボタンは全項目 OK + `!isSubmitting` の場合のみ enabled。

### 5.5 API 連携

| 操作 | 呼出先 |
|---|---|
| 登録ボタン | `useAuth().signup(email, password)` |

成功時:
1. `clearPasswordState()` を実行
2. `router.push("/")` で `/` へ遷移（MainScreen が予算未設定なら `/budget` へ replace）

失敗時:
- `error.code` を `authMessages` から引いて `errorMessage` に表示

### 5.6 アクセシビリティ
- セマンティック HTML: `<form>`, `<label htmlFor>`, `<input type="email">`, `<input type="password" autoComplete="new-password">`
- エラーメッセージは `aria-live="polite"` の領域にレンダリング
- パスワード強度ヒントは `<ul>` でリスト化

### 5.7 ダメ化UX 文言例

| 場面 | 文言 |
|---|---|
| ヘッダ | 「ようこそ、ダメへの第一歩へ」 |
| 「はじめる」キャプション | 「30 秒でダメ化体験スタート」 |

---

## 6. `LoginScreen`

### 6.1 役割
US-0-02 の主役。メール + パスワードでログイン。

### 6.2 Props
なし（`/login` ページから直接マウント）。

### 6.3 State

```ts
type LoginState = {
  email: string;
  password: string;
  isSubmitting: boolean;
  errorMessage: string | null;
  fromSessionExpired: boolean;  // ?from=session_expired のとき true
};
```

### 6.4 検証ロジック
- メール形式チェックのみ（パスワード強度はログイン時はチェックしない）

### 6.5 API 連携

| 操作 | 呼出先 |
|---|---|
| 「ログイン」ボタン | `useAuth().login(email, password)` |

成功時: `clearPasswordState()` → `router.push("/")` 。

失敗時:
- `INVALID_CREDENTIALS` → 「ログインすらめんどくさいですよね…もう一度お試しください」
- `RATE_LIMIT_EXCEEDED` → 「ちょっと頑張りすぎです。少し待ってから試してください」
- `NETWORK_ERROR` → 「通信が…ちょっと待ってもう一度」

### 6.6 セッション失効時の表示
`fromSessionExpired === true` の場合、画面上部に「お疲れ様でした。もう一度ログインしてください」のヒント領域を表示（フォーム自体は通常ログイン画面と同一）。

### 6.7 アクセシビリティ
- `<input type="password" autoComplete="current-password">`
- 「パスワードを忘れた」リンクは MVP 範囲外（リンク自体を出さない）

---

## 7. `AuthGuard`

### 7.1 役割
`(authenticated)` レイアウトに適用し、未認証時は `/login` にリダイレクト。

### 7.2 Props

```ts
type AuthGuardProps = { children: React.ReactNode };
```

### 7.3 実装方針

```tsx
"use client";
export function AuthGuard({ children }: AuthGuardProps) {
  const { status } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (status === "unauthenticated") {
      router.replace("/login");
    }
  }, [status]);

  if (status === "loading") return <FullScreenLoader />;
  if (status === "unauthenticated") return null;
  return <>{children}</>;
}
```

### 7.4 SSR / SEO 考慮
本 MVP はクライアント側ガードのみ（Cognito の Authorizer がサーバ側ガードを担うため、SEO や CSR 制約は問題にならない）。

---

## 8. `LogoutButton` + `LogoutConfirmModal`

### 8.1 構造

```
[ヘッダ右上 LogoutButton] ── クリック ──► [LogoutConfirmModal] ── 確定 ──► useAuth().logout()
                                              │
                                              └── キャンセル ──► モーダル閉じる
```

### 8.2 LogoutButton
- 表示: アイコン（⏏︎）+ 「ログアウト」
- 配置: アプリのヘッダ（認証済みレイアウト内のみ）

### 8.3 LogoutConfirmModal
- 文言（ダメ化トーン軽）:
  - タイトル: 「もうダメ化を終わらせますか？」
  - 本文: 「もう一度入るには再ログインが必要です」
  - キャンセル / ログアウト

### 8.4 アクセシビリティ
- モーダルは `<dialog>` または `role="dialog" aria-modal="true"`
- 確定ボタンに `aria-describedby` でタイトルを参照

---

## 9. `SessionExpiredModal`

### 9.1 役割
JWT 失効を検出した際に表示するシステムモーダル。`<SessionExpiredModalHost />` で常時マウントされ、トリガー時に表示される。

### 9.2 表示条件
以下のいずれかで Open:
- `apiClient` の HTTP interceptor が 401 を検出
- Auth Hub `tokenRefresh_failure` イベント発火

### 9.3 文言
- タイトル: 「お疲れ様でした」
- 本文: 「もう一度ログインしてください」
- 自動遷移: モーダル表示後 1.5 秒で `router.push("/login?from=session_expired")`
- ユーザが OK ボタンを押せば即時遷移

### 9.4 二重発火抑止
内部 ref で「既に handling 中」フラグを保持し、複数の 401 が同時に来ても 1 回しか表示されないようにする。

---

## 10. `apiClient` の 401 interceptor

```ts
import { fetchAuthSession } from "aws-amplify/auth";

export const apiClient = {
  async request(input: RequestInit & { url: string }): Promise<Response> {
    const session = await fetchAuthSession();
    const idToken = session.tokens?.idToken?.toString();
    const res = await fetch(input.url, {
      ...input,
      headers: {
        ...input.headers,
        Authorization: idToken ? `Bearer ${idToken}` : "",
      },
    });
    if (res.status === 401) {
      triggerSessionExpired();  // §9 と接続
    }
    return res;
  },
};
```

---

## 11. メッセージリソース `lib/authMessages.ts`

```ts
export const authMessages: Record<AuthErrorCode, string> = {
  INVALID_CREDENTIALS: "ログインすらめんどくさいですよね…もう一度お試しください",
  EMAIL_ALREADY_EXISTS: "このメールはもう使われています。ログインしますか？",
  WEAK_PASSWORD: "パスワードは8文字以上、英大文字・小文字・数字を含めてください",
  INVALID_EMAIL_FORMAT: "メールアドレスの形式が正しくないようです",
  RATE_LIMIT_EXCEEDED: "ちょっと頑張りすぎです。少し待ってから試してください",
  NETWORK_ERROR: "通信が…ちょっと待ってもう一度",
  UNKNOWN: "うまくいきませんでした。もう一度お試しください",
};
```

---

## 12. ステート遷移図（クライアント認証状態）

```
                      ┌──────────┐
            初期化──►│ loading  │
                      └────┬─────┘
                           │ getCurrentUser()
                           │
              ┌────────────┴────────────┐
              ▼                         ▼
     ┌────────────────┐        ┌─────────────────┐
     │ unauthenticated│        │  authenticated  │
     └────┬───────────┘        └────┬────────────┘
          │  login / signup 成功    │ logout
          └──────────────────────────┤
                                     │
                                     │ tokenRefresh_failure
                                     │  または 401 検出
                                     ▼
                            ┌─────────────────┐
                            │ unauthenticated │
                            │ + Modal 1.5s    │
                            │ → /login        │
                            └─────────────────┘
```

---

## 13. テスト観点（Code Generation 時のテスト計画への引き継ぎ）

| テスト種別 | 対象 | 観点 |
|---|---|---|
| 単体 | `useAuth` のメソッド | signup/login/logout の各 happy path / error path |
| 単体 | パスワード強度判定 | R-Pwd-1 の境界値 |
| 単体 | `apiClient` 401 interceptor | `triggerSessionExpired` が呼ばれること |
| 統合 | `<SignupScreen>` | フォーム検証、ボタン disabled 制御、API 呼出 |
| 統合 | `<AuthGuard>` | 未認証時に `router.replace("/login")` が呼ばれること |
| E2E | ハッピーパス Signup → Login → Logout | US-0-01 / US-0-02 受入基準 |
| E2E | 総タップ数 | US-0-01 受入基準（5 回以下） |

---

## 14. 後続ステージへの引き継ぎ

| 引き継ぎ先 | 内容 |
|---|---|
| NFR Requirements | フォーム入力の応答時間、`useAuth` 状態確定時間 |
| NFR Design | パスワード入力フィールドの autoComplete 属性のセキュリティ考慮 |
| Infrastructure Design | Amplify 設定値 (UserPoolId, ClientId, Region) の env 注入経路 |
| Code Generation | 上記コンポーネント / フック / 設定ファイル一式の実装、テスト |
