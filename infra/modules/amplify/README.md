# `infra/modules/amplify/` — Amplify Hosting Module (Unit 横串)

AWS Amplify Hosting (Next.js App Router SSR、`platform = WEB_COMPUTE`)。
unit-of-work.md §4.1 の `infra/modules/amplify/` (PWA 配信、Unit A/B/C/D/E 共通) 定義に準拠。

## リソース

- `aws_amplify_app.web` (build_spec = web/amplify.yml、Repository = GitHub)
- `aws_amplify_branch.develop` (auto_build、env: NEXT_PUBLIC_* + server-only API_ENDPOINT + AMPLIFY_MONOREPO_APP_ROOT=web)
- `aws_iam_role.amplify_ssr` + inline policy (CloudWatch Logs `/aws/amplify/*` 最小権限。旧 AWSAmplifyServerSideRendering managed policy は AWS から削除済)

CodeStar Connection は本 module 内では作成せず、envs/ 側で作成して URL 経由で repository を指定する (Amplify は GitHub App / OAuth Token なしの直接 URL でも動作)。

## 利用例

```hcl
module "amplify" {
  source                      = "../../modules/amplify"
  env                         = "dev"
  region                      = "ap-northeast-1"
  github_owner                = local.github_owner
  github_repo_name            = local.github_repo
  github_branch               = local.github_branch
  cognito_user_pool_id        = module.cognito.user_pool_id
  cognito_user_pool_client_id = module.cognito.user_pool_client_id
  api_endpoint                = module.api_gateway.api_endpoint
}
```

## env vars (BFF パターン整合)

| 変数 | スコープ | 用途 |
|---|---|---|
| `AMPLIFY_MONOREPO_APP_ROOT` | Build | Amplify monorepo build (公式仕様、必須) |
| `NEXT_PUBLIC_USER_POOL_ID` | Browser | Amplify Auth がブラウザで利用 |
| `NEXT_PUBLIC_USER_POOL_CLIENT_ID` | Browser | 同上 |
| `NEXT_PUBLIC_AWS_REGION` | Browser | 同上 |
| `API_ENDPOINT` | server-only | catch-all Route Handler (BFF) が API Gateway を呼ぶ際に利用 |

`NEXT_PUBLIC_API_ENDPOINT` は **存在させない** (BFF パターン: API URL 秘匿化)。
