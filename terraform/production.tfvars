# Production deployment configuration
# Copy this to terraform.tfvars and customize for production

aws_region = "us-west-2"
namespace = "runbook-engine-prod"
environment = "production"
domain_name = "runbook-engine.company.com"

# Production resource sizing
postgres_storage_size = "500Gi"
postgres_cpu_request = "1000m"
postgres_cpu_limit = "4000m"
postgres_memory_request = "2Gi"
postgres_memory_limit = "8Gi"

redis_storage_size = "50Gi"
redis_cpu_request = "500m"
redis_cpu_limit = "2000m"
redis_memory_request = "512Mi"
redis_memory_limit = "2Gi"

# Production scaling
api_image = "your-registry/runbook-engine/api:v1.0.0"
api_replicas = 5
api_min_replicas = 3
api_max_replicas = 20
api_cpu_request = "500m"
api_cpu_limit = "2000m"
api_memory_request = "512Mi"
api_memory_limit = "2Gi"

frontend_image = "your-registry/runbook-engine/frontend:v1.0.0"
frontend_replicas = 3
frontend_cpu_request = "100m"
frontend_cpu_limit = "500m"
frontend_memory_request = "128Mi"
frontend_memory_limit = "512Mi"

# High availability
temporal_server_replicas = 3

# Production monitoring
prometheus_storage_size = "200Gi"
grafana_admin_password = "super-secure-production-password"
