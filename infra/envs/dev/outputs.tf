# dev env として最低限の output を公開する
output "user_pool_id"           { value = module.auth.user_pool_id }
output "user_pool_client_id"    { value = module.auth.user_pool_client_id }
output "api_endpoint"           { value = module.auth.api_endpoint }
output "amplify_default_domain" { value = module.auth.amplify_default_domain }
output "ecr_repository_url"     { value = module.auth.ecr_repository_url }
output "codepipeline_name"      { value = module.auth.codepipeline_name }
