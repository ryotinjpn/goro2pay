locals {
  # NFRC-C20 / Q-N8 = B / Q-I10 = A: Claude Haiku 4.5, JP Inference Profile
  # 旧 (Claude 3.5 Haiku / apac.) は ap-northeast-1 で list-inference-profiles
  # に表示されず、Lambda 実行時に "Cross-region model not found" になるため、
  # 現行 ACTIVE な Haiku 4.5 + JP profile に更新する (data 越境を防ぐ目的でも JP)。
  model_id             = "anthropic.claude-haiku-4-5-20251001-v1:0"
  inference_profile_id = "jp.anthropic.claude-haiku-4-5-20251001-v1:0"

  policy_name = "gp-${var.env}-bedrock-inference-policy"
  account_id  = data.aws_caller_identity.current.account_id

  # Foundation Model は region 横断 ARN ("*"::foundation-model/...) を許可。
  # Inference Profile は region + account 別の ARN を必要とする。
  model_arns = [
    "arn:aws:bedrock:*::foundation-model/${local.model_id}",
    "arn:aws:bedrock:${var.region}:${local.account_id}:inference-profile/${local.inference_profile_id}",
  ]
}
