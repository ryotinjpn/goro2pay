# Production Environment (Placeholder)

This directory is reserved for the production environment.

The MVP is designed for hackathon demo only and is not deployed to production.
When productization is decided, the following changes are required:

- Cognito `deletion_protection = "ACTIVE"` (Q-I13)
- ECR `force_delete = false` (Q-I15)
- API Gateway access logging を有効化
- WAF / Cognito Advanced Security Features 導入 (A-NFR-SEC-09)
- CloudWatch Custom Metrics + EMF 出力 (A-NFR-OBS-02 を再評価)
- CloudWatch Log retention 30 日以上に拡大 (Q-I12)
- Email 配信を SES 統合に変更 (Q-I7)
- パスワードリセット機能の追加 (要件外、本番化時に追加)
- Amplify branch を `main` に切替、stage を PRODUCTION に変更 (Q-I14)
- CodePipeline Source branch を `main` に切替 (Q-I15)
- CORS の allow_origins を Frontend ドメインに絞る
- ECR `image_tag_mutability = IMMUTABLE` への変更 (改ざん防止)
- CodePipeline に Deploy stage を追加 (現状は CodeBuild post_build で update-function-code、本来 Deploy stage で表現すべき)
- IAM policy の静的解析を CI に追加 (tflint / checkov / IAM Access Analyzer)。terraform-test の mock_provider では IAM policy JSON の deep assert が困難なため
