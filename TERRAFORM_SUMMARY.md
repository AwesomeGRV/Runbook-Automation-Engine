# Terraform deployment for Runbook Automation Engine

## 🚀 Quick Deployment

### Prerequisites
- Terraform 1.0+
- kubectl configured
- Helm 3.0+
- Kubernetes cluster with ingress controller

### 1. Configure Environment

```bash
# Copy example configuration
cp terraform/terraform.tfvars.example terraform/terraform.tfvars

# Edit configuration
vim terraform/terraform.tfvars
```

### 2. Deploy

```bash
# Make deployment script executable
chmod +x terraform/deploy.sh

# Run full deployment
./terraform/deploy.sh
```

### 3. Access Services

After deployment:
- **Main App**: `https://your-domain.com`
- **API**: `https://your-domain.com/api`
- **Temporal UI**: `https://your-domain.com/temporal`
- **Grafana**: `https://your-domain.com/grafana`

## 📁 Files Created

```
terraform/
├── main.tf              # Core infrastructure (databases, secrets)
├── app.tf               # Application deployments (API, frontend, Temporal)
├── monitoring.tf        # Prometheus, Grafana, HPA, network policies
├── variables.tf         # All configuration variables
├── outputs.tf           # Deployment outputs
├── deploy.sh            # Automated deployment script
├── README.md            # Detailed documentation
├── terraform.tfvars.example    # Example configuration
├── production.tfvars    # Production settings
└── development.tfvars   # Development settings
```

## 🏗️ Architecture Deployed

### Core Services
- **PostgreSQL**: Primary database with persistent storage
- **Redis**: Caching layer with persistence
- **Temporal**: Workflow orchestration (via Helm)
- **API Service**: Go backend with autoscaling
- **Frontend**: React application

### Monitoring Stack
- **Prometheus**: Metrics collection and storage
- **Grafana**: Visualization dashboards
- **HPA**: Horizontal Pod Autoscaling
- **Network Policies**: Security controls

### Infrastructure
- **Namespace**: Isolated environment with random suffix
- **Secrets**: Auto-generated passwords and keys
- **Ingress**: TLS termination with cert-manager
- **PVCs**: Persistent storage for databases

## 🔧 Key Features

### Security
- Random passwords for all services
- Network policies restricting traffic
- TLS certificates via cert-manager
- Non-root containers

### Scalability
- Horizontal Pod Autoscaling for API
- Configurable replica counts
- Resource limits and requests
- High availability options

### Observability
- Complete monitoring stack
- Grafana dashboards
- Prometheus metrics
- Health checks and probes

### Production Ready
- Persistent storage
- Backup configurations
- Disaster recovery ready
- Multi-environment support

## 📊 Configuration Options

### Environment Variables
- `environment`: dev/staging/prod
- `domain_name`: External domain
- `namespace`: Base namespace name
- `ingress_class`: nginx/traefik/etc.

### Resource Sizing
- Database storage and CPU/memory
- Application replicas and scaling
- Monitoring storage size
- Frontend and API resources

### Images
- `api_image`: Backend Docker image
- `frontend_image`: Frontend Docker image
- Helm chart versions

## 🎯 Deployment Commands

```bash
# Initialize only
./terraform/deploy.sh init

# Plan only
./terraform/deploy.sh plan

# Apply only
./terraform/deploy.sh apply

# Destroy all
./terraform/deploy.sh destroy

# Show outputs
./terraform/deploy.sh outputs
```

## 🔍 Access Information

After deployment, get access details:

```bash
cd terraform

# Get all outputs
terraform output

# Get specific URLs
terraform output ingress_url
terraform output temporal_ui_url
terraform output grafana_url

# Get credentials
terraform output postgres_password
terraform output redis_password
terraform output grafana_admin_password
```

## 🛠️ Customization

### Production Environment
```bash
cp terraform/production.tfvars terraform/terraform.tfvars
./terraform/deploy.sh
```

### Development Environment
```bash
cp terraform/development.tfvars terraform/terraform.tfvars
./terraform/deploy.sh
```

### Custom Configuration
Edit `terraform/terraform.tfvars` with your specific values.

## 📈 Scaling

### Manual Scaling
Update `api_replicas` in configuration and apply.

### Auto Scaling
Configure HPA settings:
- `api_min_replicas`: Minimum pods
- `api_max_replicas`: Maximum pods
- `api_cpu_target_utilization`: CPU threshold

## 🔒 Security Notes

- All passwords are randomly generated
- Secrets are stored as Kubernetes secrets
- Network policies restrict inter-pod communication
- TLS is enforced for external access
- Containers run as non-root users

## 🚨 Troubleshooting

### Common Issues
1. **Pods not starting**: Check resource limits
2. **Database connection**: Verify secrets and networking
3. **Ingress issues**: Check ingress controller and DNS
4. **TLS certificates**: Verify cert-manager setup

### Debug Commands
```bash
# Check pods
kubectl get pods -n runbook-engine-*

# Check logs
kubectl logs -f deployment/runbook-api -n runbook-engine-*

# Check events
kubectl get events -n runbook-engine-* --sort-by='.lastTimestamp'
```

This Terraform configuration provides a complete, production-ready deployment of the Runbook Automation Engine with all the enterprise features you requested! 🚀
