# Auth Unit — Frontend Components Generation Summary

**Date**: 2026-05-22
**Stage**: Code Generation Step 9-11 完了

## 生成ファイル

### Lib (横串)
| ファイル | 役割 | 関連 LC |
|---|---|---|
| `web/lib/amplifyConfig.ts` | Amplify.configure 設定 | (横串) |
| `web/lib/authMessages.ts` | AuthErrorCode → 日本語メッセージ + AuthErrorWithCode | LC-AUTH-13 |
| `web/lib/apiClient.ts` | Browser fetch ラッパ + 401/429/5xx interceptor | LC-AUTH-09 |
| `web/lib/authHubListener.ts` | Amplify Hub `tokenRefresh_failure` 購読 | LC-AUTH-14 |

### State
| ファイル | 役割 |
|---|---|
| `web/state/auth.ts` | `sessionExpiredAtom` (Jotai) | LC-AUTH-10 |

### Hooks
| ファイル | 役割 |
|---|---|
| `web/hooks/useAuth.ts` | useAuth フック (signup / login / logout、AccessToken 採用) | LC-AUTH-08 |

### Components
| ファイル | 役割 |
|---|---|
| `web/components/auth/LandingScreen.tsx` | 未認証時の Landing |
| `web/components/auth/LoginScreen.tsx` | /login (form + 401 エラー表示 + autoComplete) |
| `web/components/auth/SignupScreen.tsx` | /signup (PW 強度ヒント + form) |
| `web/components/auth/LogoutButton.tsx` | ヘッダ / 設定の Logout ボタン |
| `web/components/auth/LogoutConfirmModal.tsx` | ログアウト確認モーダル |
| `web/components/auth/SessionExpiredModalHost.tsx` | セッション失効 modal + 1.5s 自動遷移 |
| `web/components/auth/AuthGuard.tsx` | 認証必須レイアウト + 300ms Loading 遅延 |

### App Router
| ファイル | 役割 |
|---|---|
| `web/app/layout.tsx` | RootLayout (lang=ja) |
| `web/app/providers.tsx` | AppProviders (Amplify config + Jotai + QueryClient + SessionExpiredModalHost) |
| `web/app/page.tsx` | / (LandingScreen / MainScreen 切替、MainScreen は placeholder) |
| `web/app/login/page.tsx` | /login (LoginScreen) |
| `web/app/(auth)/signup/page.tsx` | /signup (SignupScreen) |
| `web/app/(authenticated)/layout.tsx` | 認証必須レイアウト (AuthGuard) |
| `web/app/api/[...path]/route.ts` | BFF catch-all proxy (LC-AUTH-18) |

### Build / Config
| ファイル | 役割 |
|---|---|
| `web/amplify.yml` | Amplify monorepo build (appRoot: web、AMPLIFY_MONOREPO_APP_ROOT=web 環境変数) |
| `web/vitest.config.ts` | Vitest 設定 (jsdom, alias) |
| `web/tests/setup.ts` | Vitest setup |

## 単体テスト

| テストファイル | カバレッジ |
|---|---|
| `tests/apiClient.test.ts` | Authorization ヘッダ付与 / 401 detection / 二重発火防止 / 429 / 5xx |
| `tests/authMessages.test.ts` | 9 種別メッセージ定義 / AuthErrorWithCode の PII 漏洩防止 |
| `tests/sessionExpired.test.tsx` | atom null / atom 起動 + 1.5s タイマー → router.push |

E2E (Playwright) は Build & Test ステージで実行。

## BFF パターン整合

- ブラウザ: `apiClient.request({ path: "/api/auth/logout" })` → 同一オリジン `/api/auth/logout`
- Next.js Route Handler: `app/api/[...path]/route.ts` (catch-all) が `Authorization` 透過、`API_ENDPOINT` (server-only env) で API Gateway に proxy
- `NEXT_PUBLIC_API_ENDPOINT` は使用しない（API URL は秘匿化）

## トレーサビリティ

- US-0-01 (Signup): SignupScreen + useAuth.signup (auto-confirm 後の自動 signIn)
- US-0-02 (Login): LoginScreen + useAuth.login (Amplify Auth + JWT セッション維持)
- FR-AUTH-04 (Logout): LogoutButton + useAuth.logout (API 監査ログ → GlobalSignOut の順)
- NFR-DEG-01 (低摩擦): AuthGuard 300ms Loading 遅延、必要最小限の UI
- NFR-DEG-05 (コピー): authMessages のダメ化トーン軽
- A-NFR-SEC-07 (PW 即時クリア): finally で setPassword("")
- A-NFR-SEC-08 (PII 漏洩防止): AuthErrorWithCode が originalError を捨てる
- LC-AUTH-08〜14, LC-AUTH-18: 全コンポーネント実装済み
