# Bedrock IAM Policy 最小権限検証 (Q-I13 = C)
# mock_provider で AWS API 呼出なし。

mock_provider "aws" {}

variables {
  env    = "dev"
  region = "ap-northeast-1"
  tags   = { Project = "goro2pay", Env = "dev", ManagedBy = "terraform" }
}

run "policy_name" {
  command = plan

  assert {
    condition     = aws_iam_policy.bedrock_inference.name == "gp-dev-bedrock-inference-policy"
    error_message = "policy name must be gp-dev-bedrock-inference-policy (Q-I4)"
  }
}

run "policy_actions_limited_to_invoke_and_converse" {
  command = plan

  assert {
    condition     = length(jsondecode(aws_iam_policy.bedrock_inference.policy).Statement[0].Action) == 2
    error_message = "Action must contain exactly 2 entries (InvokeModel + Converse)"
  }

  assert {
    condition     = contains(jsondecode(aws_iam_policy.bedrock_inference.policy).Statement[0].Action, "bedrock:InvokeModel")
    error_message = "Action must include bedrock:InvokeModel"
  }

  assert {
    condition     = contains(jsondecode(aws_iam_policy.bedrock_inference.policy).Statement[0].Action, "bedrock:Converse")
    error_message = "Action must include bedrock:Converse"
  }

  assert {
    condition     = !contains(jsondecode(aws_iam_policy.bedrock_inference.policy).Statement[0].Action, "bedrock:*")
    error_message = "Action must NOT include bedrock:* (least privilege violation)"
  }

  assert {
    condition     = !contains(jsondecode(aws_iam_policy.bedrock_inference.policy).Statement[0].Action, "bedrock:InvokeModelWithResponseStream")
    error_message = "Streaming action must NOT be granted in MVP (NFRC-C20)"
  }
}

run "policy_resources_limited_to_haiku_arns" {
  command = plan

  assert {
    condition     = length(jsondecode(aws_iam_policy.bedrock_inference.policy).Statement[0].Resource) == 2
    error_message = "Resource must contain exactly 2 ARNs (Foundation Model + Inference Profile)"
  }

  assert {
    condition     = anytrue([for r in jsondecode(aws_iam_policy.bedrock_inference.policy).Statement[0].Resource : strcontains(r, "claude-3-5-haiku")])
    error_message = "Resource ARNs must include claude-3-5-haiku"
  }

  assert {
    condition     = !anytrue([for r in jsondecode(aws_iam_policy.bedrock_inference.policy).Statement[0].Resource : strcontains(r, "claude-3-opus") || strcontains(r, "claude-3-sonnet")])
    error_message = "Other Claude models must NOT be granted"
  }
}
