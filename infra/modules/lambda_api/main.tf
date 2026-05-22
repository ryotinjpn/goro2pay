# Lambda API module: API Lambda 本体 + ECR + CodePipeline + CodeBuild + S3 artifacts
# unit-of-work.md §4.1 の `infra/modules/lambda_api/` (Unit 横串、API Lambda + ECR) 定義に準拠。
# 他 Unit (B/C/D/E) は本 module の output (api_lambda_role_arn 等) に
# 権限を attach する形で機能を追加する。
#
# resource 定義は機能別ファイル (api_lambda.tf / ecr.tf / codepipeline.tf / iam.tf) に配置。
