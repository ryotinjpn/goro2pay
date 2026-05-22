output "amplify_app_id" {
  description = "Amplify App ID"
  value       = aws_amplify_app.web.id
}

output "amplify_default_domain" {
  description = "Amplify default domain (xxxxx.amplifyapp.com)"
  value       = aws_amplify_app.web.default_domain
}
