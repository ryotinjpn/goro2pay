# dev env outputs
output "user_pool_id" { value = module.cognito.user_pool_id }
output "user_pool_client_id" { value = module.cognito.user_pool_client_id }
output "api_endpoint" { value = module.api_gateway.api_endpoint }
output "amplify_default_domain" { value = module.amplify.amplify_default_domain }
output "ecr_repository_url" { value = module.lambda_api.ecr_repository_url }
output "codepipeline_name" { value = module.lambda_api.codepipeline_name }

# Unit C / 横串 observability outputs
output "order_history_table_name" { value = module.order_history.dynamodb_table_name }
output "sns_topic_arn" { value = module.observability.sns_topic_arn }
