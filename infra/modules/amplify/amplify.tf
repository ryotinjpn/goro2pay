# Amplify Hosting (Q-I14=A) — CodeStar Connection は envs 側で別途作成
#
# GitHub 接続について:
# 本 module は oauth_token / access_token を terraform 側で管理しない。
# 初回 apply 後、AWS Console で手動再接続 (Reconnect repository → GitHub App
# インストール) することで repo 紐付け + webhook 設定を確立する。
# CodeStar Connection の Console 承認 (deployment-runbook.md §4) と同じ
# 「初回 1 度だけの手動操作」として運用する。

resource "aws_amplify_app" "web" {
  name                 = "${local.prefix}-web"
  repository           = "https://github.com/${var.github_owner}/${var.github_repo_name}"
  iam_service_role_arn = aws_iam_role.amplify_ssr.arn
  platform             = "WEB_COMPUTE"

  enable_branch_auto_build = true

  build_spec = file(var.amplify_yml_path)

  custom_rule {
    source = "/<*>"
    target = "/index.html"
    status = "404-200"
  }

  # oauth_token / access_token は Console での手動接続後に AWS 側で保持される。
  # terraform 側で空文字を渡すと一部 provider バージョンで ValidationException に
  # なるため、属性自体を省略する。一度接続したあとは drift 扱いされない。
  lifecycle {
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
