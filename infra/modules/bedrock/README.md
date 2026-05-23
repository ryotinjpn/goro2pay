# `modules/bedrock`

Unit C / D 共有: Bedrock Claude 3.5 Haiku 呼出用 IAM Policy。

## リソース

- `aws_iam_policy.bedrock_inference`: `bedrock:InvokeModel` + `bedrock:Converse` を Foundation Model + Inference Profile ARN に限定して許可

## Outputs

| Output | 用途 |
|---|---|
| `bedrock_policy_arn` | `modules/lambda_api` の `additional_policy_arns` に渡す |

## 仕様根拠

- NFRC-C20 / Q-N8 = B (Claude 3.5 Haiku)
- Q-I3 = A (最小権限、`bedrock:*` ではなく特定モデルに限定)
- Q-I10 = A (APAC Inference Profile 経由)

## モデル ID

- Foundation: `anthropic.claude-3-5-haiku-20241022-v1:0`
- Inference Profile (APAC): `apac.anthropic.claude-3-5-haiku-20241022-v1:0`
