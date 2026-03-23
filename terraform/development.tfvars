# Development deployment configuration
# Copy this to terraform.tfvars for development environment

aws_region = "us-west-2"
namespace = "runbook-engine-dev"
environment = "development"
domain_name = "runbook-engine-dev.company.com"

# Development resource sizing (smaller)
postgres_storage_size = "20Gi"
postgres_cpu_request = "250m"
postgres_cpu_limit = "500m"
postgres_memory_request = "256Mi"
postgres_memory_limit = "512Mi"

redis_storage_size = "8Gi"
redis_cpu_request = "100m"
redis_cpu_limit = "200m"
redis_memory_request = "128Mi"
redis_memory_limit = "256Mi"

# Development scaling (minimal)
api_image = "your-registry/runbook-engine/api:latest"
api_replicas = 1
api_min_replicas = 1
api_max_replicas = 3
api_cpu_request = "100m"
api_cpu_limit = "500m"
api_memory_request = "128Mi"
api_memory_limit = "512Mi"

frontend_image = "your-registry/runbook-engine/frontend:latest"
frontend_replicas = 1
frontend_cpu_request = "50m"
frontend_cpu_limit = "200m"
frontend_memory_request = "64Mi"
frontend_memory_limit = "256Mi"

# Single instance for development
temporal_server_replicas = 1

# Development monitoring (smaller)
prometheus_storage_size = "20Gi"
grafana_admin_password = "dev-password"
