# Amplify Hosting (Q-I14=A) + CodeStar Connection
# Frontend 自動デプロイ。CodeStar Connection は CodePipeline でも再利用される。

resource "aws_codestarconnections_connection" "github" {
  name          = "${local.prefix}-github-conn"
  provider_type = "GitHub"

  # 注意: 作成直後は Pending 状態。AWS Console で手動承認が必要 (1 度だけ)。
  # README.md の デプロイ手順 §3 を参照。
}

resource "aws_amplify_app" "web" {
  name         = "${local.prefix}-web"
  repository   = "https://github.com/${var.github_owner}/${var.github_repo_name}"
  iam_service_role_arn = aws_iam_role.amplify_ssr.arn
  platform     = "WEB_COMPUTE" # Next.js App Router SSR

  enable_branch_auto_build = true

  # Amplify monorepo: web/amplify.yml が build spec。Console 設定は inline で持つ。
  build_spec = file("${path.module}/../../../web/amplify.yml")

  # Connection 経由で GitHub と連携 (OAuth Token 不要)
  oauth_token = ""
  access_token = ""

  custom_rule {
    source = "/<*>"
    target = "/index.html"
    status = "404-200"
  }
}

resource "aws_amplify_branch" "develop" {
  app_id      = aws_amplify_app.web.id
  branch_name = var.github_branch
  framework   = "Next.js - SSR"
  stage       = "DEVELOPMENT"

  enable_auto_build = true

  environment_variables = {
    # AMPLIFY_MONOREPO_APP_ROOT は monorepo build に必須 (公式仕様)
    AMPLIFY_MONOREPO_APP_ROOT = "web"

    # Browser に露出する公開 env (NEXT_PUBLIC_*)
    NEXT_PUBLIC_USER_POOL_ID        = aws_cognito_user_pool.main.id
    NEXT_PUBLIC_USER_POOL_CLIENT_ID = aws_cognito_user_pool_client.web.id
    NEXT_PUBLIC_AWS_REGION          = var.region

    # server-only env (NEXT_PUBLIC_ なし)。catch-all Route Handler が利用
    API_ENDPOINT = aws_apigatewayv2_api.main.api_endpoint
  }
}
