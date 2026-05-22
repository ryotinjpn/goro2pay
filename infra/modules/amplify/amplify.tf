# Amplify Hosting (Q-I14=A) — CodeStar Connection は envs 側で別途作成

resource "aws_amplify_app" "web" {
  name                 = "${local.prefix}-web"
  repository           = "https://github.com/${var.github_owner}/${var.github_repo_name}"
  iam_service_role_arn = aws_iam_role.amplify_ssr.arn
  platform             = "WEB_COMPUTE"

  enable_branch_auto_build = true

  build_spec = file("${path.module}/${var.amplify_yml_path}")

  oauth_token  = ""
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
    AMPLIFY_MONOREPO_APP_ROOT = "web"

    # Browser に露出する公開 env (NEXT_PUBLIC_*)
    NEXT_PUBLIC_USER_POOL_ID        = var.cognito_user_pool_id
    NEXT_PUBLIC_USER_POOL_CLIENT_ID = var.cognito_user_pool_client_id
    NEXT_PUBLIC_AWS_REGION          = var.region

    # server-only env (NEXT_PUBLIC_ なし、BFF Route Handler 用)
    API_ENDPOINT = var.api_endpoint
  }
}
