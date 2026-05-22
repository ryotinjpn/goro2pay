# Amplify module: Amplify Hosting (Next.js App Router SSR) + Branch + SSR IAM Role
# unit-of-work.md §4.1 の `infra/modules/amplify/` (PWA 配信、Unit A/B/C/D/E 共通) 定義に準拠。
# CodeStar Connection は envs/ 側で作成して ARN を渡す (lambda_api と共有)。
#
# resource 定義は機能別ファイル (amplify.tf / iam.tf) に分けて配置している。
# 規約 (terraform-module-design) により main.tf 直下に resource を集中配置せず、
# 各リソースの責務単位でファイル分割する方針を採用。
