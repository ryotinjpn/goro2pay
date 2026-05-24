# Amplify Hosting (Q-I14=A) — CodeStar Connection は envs 側で別途作成
#
# GitHub 接続について:
# `aws_amplify_app` は repository を指定すると CreateApp 時に oauth_token /
# access_token のいずれかを必須とするため、token 無しで apply すると
# "You should at least provide one valid token" で失敗する。
# 本 module は GitHub PAT (classic, scopes: admin:repo_hook + repo) を
# var.github_oauth_token で受け取り、envs 側で SSM SecureString
# `/goro2pay/${env}/amplify/github_oauth_token` から data source 経由で
# 注入する。AWS 側でハッシュ化保管後は ignore_changes で diff を抑える。

resource "aws_amplify_app" "web" {
  name                 = "${local.prefix}-web"
  repository           = "https://github.com/${var.github_owner}/${var.github_repo_name}"
  iam_service_role_arn = aws_iam_role.amplify_ssr.arn
  platform             = "WEB_COMPUTE"

  enable_branch_auto_build = true

  build_spec = file(var.amplify_yml_path)

  # GitHub Personal Access Token (classic, admin:repo_hook + repo)。
  # CreateApp 時に Amplify が GitHub Webhook を登録するため初回 apply で必要。
  # AWS 側で受け取った後はマスクされ、後続の更新では使われないため、
  # tf 側 (tfvars / TF_VAR_) からも消す運用にしている。
  oauth_token = var.github_oauth_token

  # NOTE: SPA 用の custom_rule (`/<*>` → `/index.html` 404-200) は WEB_COMPUTE
  # (Next.js SSR) では設定しない。Next.js Server がリクエストを受けて自前で
  # ルーティングするため、Amplify 側で 404 を /index.html に書き換えると
  # Server 経路と干渉して二重処理 / 想定外 fallback の原因になる。

  lifecycle {
    # oauth_token は AWS 側でハッシュ化保管されるため、tf state とは差分が
    # 出続ける。再 apply の度に oauth_token が更新されないよう ignore する。
    # access_token も将来的な切り替え用に残す。
    ignore_changes = [oauth_token, access_token]
  }
}

resource "aws_amplify_branch" "develop" {
  app_id      = aws_amplify_app.web.id
  branch_name = var.github_branch
  framework   = "Next.js - SSR"
  stage       = "DEVELOPMENT"

  enable_auto_build = true

  environment_variables = {
    AMPLIFY_MONOREPO_APP_ROOT = "web"

    # Browser に露出する公開 env (NEXT_PUBLIC_*)
    NEXT_PUBLIC_USER_POOL_ID        = var.cognito_user_pool_id
    NEXT_PUBLIC_USER_POOL_CLIENT_ID = var.cognito_user_pool_client_id
    NEXT_PUBLIC_AWS_REGION          = var.region

    # server-only env (NEXT_PUBLIC_ なし、BFF Route Handler 用)
    API_ENDPOINT = var.api_endpoint
  }
}
