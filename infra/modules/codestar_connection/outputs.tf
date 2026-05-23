output "connection_arn" {
  description = "CodeStar Connection ARN (CodePipeline / Amplify が参照)"
  value       = aws_codestarconnections_connection.github.arn
}

output "connection_name" {
  description = "CodeStar Connection 名"
  value       = aws_codestarconnections_connection.github.name
}
