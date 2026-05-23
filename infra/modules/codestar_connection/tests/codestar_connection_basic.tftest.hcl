mock_provider "aws" {}

variables {
  env = "dev"
}

run "plan_succeeds" {
  command = plan
}

run "resources_present" {
  command = plan

  # mock_provider では output (computed attr 経由) は plan 時点で unknown のため、
  # resource の non-computed 属性で存在を確認する。
  assert {
    condition     = aws_codestarconnections_connection.github.name == "gp-dev-github-conn"
    error_message = "Connection 名が想定と異なる"
  }
  assert {
    condition     = aws_codestarconnections_connection.github.provider_type == "GitHub"
    error_message = "provider_type は GitHub であること"
  }
  assert {
    condition     = aws_codestarconnections_connection.github.tags["Unit"] == "shared"
    error_message = "Unit タグは shared (CodePipeline と Amplify の共有リソースのため)"
  }
}
