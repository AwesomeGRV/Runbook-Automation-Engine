# Terraform Deployment Guide

## Overview

This Terraform configuration deploys the Runbook Automation Engine to Kubernetes with all required components including databases, monitoring, and networking.

## Prerequisites

Before deploying, ensure you have:

1. **Terraform** (v1.0+)
2. **kubectl** configured for your target cluster
3. **Helm** (v3.0+)
4. **Ingress Controller** (nginx-ingress recommended)
5. **cert-manager** for TLS certificates

## Quick Start

### 1. Configure Variables

Create a `terraform.tfvars` file to customize your deployment:

```hcl
# terraform.tfvars
aws_region           = "us-west-2"
domain_name          = "runbook-engine.yourdomain.com"
environment          = "production"
ingress_class        = "nginx"
cluster_issuer       = "letsencrypt-prod"
storage_class_name   = "gp2"

# Image versions
api_image            = "your-registry/runbook-engine/api:v1.0.0"
frontend_image       = "your-registry/runbook-engine/frontend:v1.0.0"

# Resource sizing
postgres_storage_size = "100Gi"
redis_storage_size    = "20Gi"
prometheus_storage_size = "100Gi"

# Scaling
api_replicas         = 3
api_min_replicas     = 1
api_max_replicas     = 10

# Security
grafana_admin_password = "your-secure-password"
```

### 2. Deploy the Application

Use the deployment script for easy deployment:

```bash
# Make the script executable
chmod +x terraform/deploy.sh

# Run the full deployment
./terraform/deploy.sh
```

Or use Terraform directly:

```bash
cd terraform

# Initialize Terraform
terraform init

# Plan the deployment
terraform plan

# Apply the deployment
terraform apply

# Show outputs
terraform output
```

## Architecture

The Terraform configuration deploys:

### Core Services
- **PostgreSQL**: Primary database with persistent storage
- **Redis**: Caching layer with persistence
- **Temporal**: Workflow orchestration engine
- **API Service**: Main application backend
- **Frontend**: React web application

### Observability
- **Prometheus**: Metrics collection
- **Grafana**: Visualization dashboards
- **AlertManager**: Alert management

### Infrastructure
- **Namespace**: Isolated environment
- **Services**: Internal service discovery
- **Ingress**: External access with TLS
- **HPA**: Horizontal Pod Autoscaling
- **Network Policies**: Security controls

## File Structure

```
terraform/
├── main.tf          # Core infrastructure (databases, secrets)
├── app.tf           # Application deployments
├── monitoring.tf    # Monitoring and observability
├── variables.tf     # Configuration variables
├── outputs.tf       # Output values
└── deploy.sh        # Deployment script
```

## Configuration Options

### Database Configuration
- `postgres_storage_size`: PostgreSQL storage allocation
- `postgres_cpu_*`: CPU requests/limits
- `postgres_memory_*`: Memory requests/limits

### Application Scaling
- `api_replicas`: Number of API pods
- `api_min_replicas`: HPA minimum replicas
- `api_max_replicas`: HPA maximum replicas
- `api_cpu_target_utilization`: HPA CPU target

### Monitoring
- `prometheus_storage_size`: Prometheus storage
- `grafana_admin_password`: Grafana admin password

### Networking
- `domain_name`: External domain for the application
- `ingress_class`: Ingress controller class
- `cluster_issuer`: TLS certificate issuer

## Accessing Services

After deployment, access the services at:

- **Main Application**: `https://your-domain.com`
- **API**: `https://your-domain.com/api`
- **Temporal UI**: `https://your-domain.com/temporal`
- **Grafana**: `https://your-domain.com/grafana`
- **Prometheus**: `https://your-domain.com/prometheus`

## Security Considerations

### Secrets
All secrets are generated randomly and stored as Kubernetes secrets:
- Database passwords
- Redis password
- JWT secret
- Grafana admin password

### Network Security
- Network policies restrict traffic between pods
- Only necessary ports are exposed
- TLS termination at ingress level

### RBAC
The deployment uses service accounts with minimal permissions.

## Monitoring and Alerting

### Prometheus Metrics
The application exposes metrics at `/metrics` endpoint:
- HTTP request metrics
- Database connection metrics
- Workflow execution metrics
- Resource utilization metrics

### Grafana Dashboards
Pre-configured dashboards include:
- Application performance
- Database performance
- Kubernetes cluster metrics
- Temporal workflow metrics

## Troubleshooting

### Common Issues

1. **Pods not starting**: Check resource limits and node capacity
2. **Database connection failures**: Verify secrets and network policies
3. **Ingress not working**: Check ingress controller and DNS settings
4. **TLS certificate issues**: Verify cert-manager configuration

### Debug Commands

```bash
# Check pod status
kubectl get pods -n runbook-engine-xxxx

# Check pod logs
kubectl logs -f deployment/runbook-api -n runbook-engine-xxxx

# Check events
kubectl get events -n runbook-engine-xxxx --sort-by='.lastTimestamp'

# Check services
kubectl get svc -n runbook-engine-xxxx

# Check ingress
kubectl get ingress -n runbook-engine-xxxx
```

## Maintenance

### Updates
To update the application:

1. Update image versions in `terraform.tfvars`
2. Run `terraform apply`
3. Monitor the rollout

### Scaling
Adjust scaling parameters in `terraform.tfvars`:
- `api_replicas` for manual scaling
- HPA settings for automatic scaling

### Backups
- Database backups should be configured separately
- Consider using a managed database service for production

## Cost Optimization

### Resource Sizing
- Adjust CPU/memory requests based on actual usage
- Use smaller instance types for development environments
- Consider spot instances for non-critical workloads

### Storage
- Use appropriate storage classes for your cloud provider
- Implement data retention policies
- Monitor storage usage and clean up unused data

## Production Best Practices

1. **Use separate environments** for dev/staging/prod
2. **Implement proper backup strategies** for databases
3. **Set up comprehensive monitoring** and alerting
4. **Use managed services** for databases when possible
5. **Implement proper CI/CD** pipelines
6. **Regular security updates** and patching
7. **Disaster recovery planning** and testing
