# Amplify SSR Compute IAM Role

resource "aws_iam_role" "amplify_ssr" {
  name = "${local.prefix}-amplify-ssr-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Principal = {
        Service = "amplify.amazonaws.com"
      }
      Action = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "amplify_ssr_managed" {
  role       = aws_iam_role.amplify_ssr.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSAmplifyServerSideRendering"
}
