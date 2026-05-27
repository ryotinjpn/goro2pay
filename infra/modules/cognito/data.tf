# Pre Sign-up Lambda zip (Q-I3=A: archive_file)
data "archive_file" "pre_signup" {
  type        = "zip"
  source_file = "${path.module}/../../functions/pre-signup/index.js"
  output_path = "${path.module}/.terraform/tmp/pre-signup.zip"
}
