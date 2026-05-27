# Amplify Hosting GitHub PAT (SSM Parameter Store SecureString)
#
# 事前作業 (1 度だけ):
#   aws ssm put-parameter \
#     --name "/goro2pay/dev/amplify/github_oauth_token" \
#     --type SecureString \
#     --value "<GitHub PAT>" \
#     --region ap-northeast-1 \
#     --profile dev-kyoto-sso-administrator
data "aws_ssm_parameter" "amplify_github_token" {
  name            = "/goro2pay/${local.env}/amplify/github_oauth_token"
  with_decryption = true
}
