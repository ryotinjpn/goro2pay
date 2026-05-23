# `infra/modules/cognito/` — Cognito Module

Amazon Cognito User Pool + App Client + Pre Sign-up Lambda Trigger をまとめた module。

## リソース

- `aws_cognito_user_pool.main` — User Pool (PW 8 文字+英大小+数字, MFA OFF, deletion_protection INACTIVE)
- `aws_cognito_user_pool_client.web` — App Client (no secret, Token 8h/30d)
- `aws_lambda_function.pre_signup` — Pre Sign-up Trigger (Node.js 20 arm64, auto-confirm)
- `aws_lambda_permission.cognito_invoke_pre_signup`
- `aws_iam_role.pre_signup_lambda` + inline policy (CloudWatch Logs のみ)
- `aws_cloudwatch_log_group.pre_signup` (retention 7 日)

## 利用例

```hcl
module "cognito" {
  source = "../../modules/cognito"
  env    = "dev"
}
```

## Outputs

| 名前 | 用途 |
|---|---|
| `user_pool_id` | API Gateway Authorizer / Frontend Amplify 等が参照 |
| `user_pool_arn` | IAM 連携用 |
| `user_pool_endpoint` | JWT issuer URL |
| `user_pool_client_id` | Frontend Amplify Auth が参照 |
| `pre_signup_lambda_arn` | (本 module 内で完結、外部参照は通常不要) |
