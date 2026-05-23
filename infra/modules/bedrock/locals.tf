locals {
  # NFRC-C20 / Q-N8 = B / Q-I10 = A: Claude 3.5 Haiku, APAC Inference Profile
  model_id             = "anthropic.claude-3-5-haiku-20241022-v1:0"
  inference_profile_id = "apac.anthropic.claude-3-5-haiku-20241022-v1:0"

  policy_name = "gp-${var.env}-bedrock-inference-policy"
  account_id  = data.aws_caller_identity.current.account_id

  # Foundation Model は region 横断 ARN ("*"::foundation-model/...) を許可。
  # Inference Profile は region + account 別の ARN を必要とする。
  model_arns = [
    "arn:aws:bedrock:*::foundation-model/${local.model_id}",
    "arn:aws:bedrock:${var.region}:${local.account_id}:inference-profile/${local.inference_profile_id}",
  ]

  tags = var.tags
}
