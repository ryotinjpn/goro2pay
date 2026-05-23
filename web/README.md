# web — ゴロゴロPay フロントエンド (Next.js App Router)

PWA フロントエンド (Next.js 15 + AWS Amplify Auth v6)。Amplify Hosting で SSR 配信。

## ローカル開発

```bash
cd web
npm install
npm run dev
```

`.env.local` にローカル開発用 env を配置:

```
NEXT_PUBLIC_USER_POOL_ID=ap-northeast-1_xxxxxxxxx
NEXT_PUBLIC_USER_POOL_CLIENT_ID=xxxxxxxxxxxxxxxxxxxx
NEXT_PUBLIC_AWS_REGION=ap-northeast-1
API_ENDPOINT=https://xxxxxxxx.execute-api.ap-northeast-1.amazonaws.com
```

`API_ENDPOINT` は **server-only** (`NEXT_PUBLIC_` なし)。catch-all Route Handler (`app/api/[...path]/route.ts`) のみが利用。

## ビルド

```bash
npm run build
```

## テスト

```bash
npm test            # Vitest 単体テスト
npm run e2e         # Playwright E2E
```

## ディレクトリ

- `app/`              — Next.js App Router
- `components/auth/`  — Auth Unit の UI コンポーネント
- `hooks/`            — useAuth 等
- `state/`            — Jotai atoms
- `lib/`              — apiClient / authMessages / Amplify 設定 / Hub Listener

## BFF パターン

- ブラウザ → `/api/*` → `app/api/[...path]/route.ts` (catch-all proxy)
  → 上流 API Gateway に Authorization ヘッダ透過で転送
- API Gateway URL (`API_ENDPOINT`) はブラウザに露出しない
