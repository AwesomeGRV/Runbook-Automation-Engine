# Outputs from Terraform configuration

output "namespace_name" {
  description = "Name of the created namespace"
  value       = kubernetes_namespace.runbook_engine.metadata[0].name
}

output "postgres_service_name" {
  description = "Name of the PostgreSQL service"
  value       = kubernetes_service.postgres.metadata[0].name
}

output "redis_service_name" {
  description = "Name of the Redis service"
  value       = kubernetes_service.redis.metadata[0].name
}

output "api_service_name" {
  description = "Name of the API service"
  value       = kubernetes_service.api.metadata[0].name
}

output "frontend_service_name" {
  description = "Name of the frontend service"
  value       = kubernetes_service.frontend.metadata[0].name
}

output "temporal_service_name" {
  description = "Name of the Temporal service"
  value       = helm_release.temporal.name
}

output "ingress_url" {
  description = "URL for accessing the application"
  value       = "https://${var.domain_name}"
}

output "postgres_password" {
  description = "PostgreSQL password (sensitive)"
  value       = random_password.postgres_password.result
  sensitive   = true
}

output "redis_password" {
  description = "Redis password (sensitive)"
  value       = random_password.redis_password.result
  sensitive   = true
}

output "jwt_secret" {
  description = "JWT secret (sensitive)"
  value       = random_password.jwt_secret.result
  sensitive   = true
}

output "grafana_url" {
  description = "URL for accessing Grafana"
  value       = "https://${var.domain_name}/grafana"
}

output "prometheus_url" {
  description = "URL for accessing Prometheus"
  value       = "https://${var.domain_name}/prometheus"
}

output "temporal_ui_url" {
  description = "URL for accessing Temporal UI"
  value       = "https://${var.domain_name}/temporal"
}
