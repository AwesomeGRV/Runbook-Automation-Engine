# Terraform deployment script for Runbook Automation Engine

#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if terraform is installed
    if ! command -v terraform &> /dev/null; then
        print_error "Terraform is not installed. Please install Terraform first."
        exit 1
    fi
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed. Please install kubectl first."
        exit 1
    fi
    
    # Check if helm is installed
    if ! command -v helm &> /dev/null; then
        print_error "Helm is not installed. Please install Helm first."
        exit 1
    fi
    
    # Check if kubectl can connect to cluster
    if ! kubectl cluster-info &> /dev/null; then
        print_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
        exit 1
    fi
    
    print_status "Prerequisites check completed successfully!"
}

# Initialize Terraform
init_terraform() {
    print_status "Initializing Terraform..."
    
    cd terraform
    
    terraform init
    
    cd ..
    
    print_status "Terraform initialization completed!"
}

# Plan Terraform deployment
plan_deployment() {
    print_status "Planning Terraform deployment..."
    
    cd terraform
    
    terraform plan -out=tfplan
    
    cd ..
    
    print_status "Terraform plan completed!"
}

# Apply Terraform deployment
apply_deployment() {
    print_status "Applying Terraform deployment..."
    
    cd terraform
    
    terraform apply tfplan
    
    cd ..
    
    print_status "Terraform deployment completed!"
}

# Show outputs
show_outputs() {
    print_status "Deployment outputs:"
    
    cd terraform
    
    terraform output
    
    cd ..
}

# Wait for pods to be ready
wait_for_pods() {
    print_status "Waiting for pods to be ready..."
    
    NAMESPACE=$(cd terraform && terraform output -raw namespace_name)
    
    # Wait for all pods to be ready
    kubectl wait --for=condition=ready pod -l app=postgres -n $NAMESPACE --timeout=300s
    kubectl wait --for=condition=ready pod -l app=redis -n $NAMESPACE --timeout=300s
    kubectl wait --for=condition=ready pod -l app=runbook-api -n $NAMESPACE --timeout=300s
    kubectl wait --for=condition=ready pod -l app=runbook-frontend -n $NAMESPACE --timeout=300s
    
    print_status "All pods are ready!"
}

# Show access information
show_access_info() {
    print_status "Access information:"
    
    cd terraform
    
    INGRESS_URL=$(terraform output -raw ingress_url)
    GRAFANA_URL=$(terraform output -raw grafana_url)
    PROMETHEUS_URL=$(terraform output -raw prometheus_url)
    TEMPORAL_UI_URL=$(terraform output -raw temporal_ui_url)
    
    cd ..
    
    echo ""
    echo "Application URLs:"
    echo "  Main Application: $INGRESS_URL"
    echo "  Grafana Dashboard: $GRAFANA_URL"
    echo "  Prometheus: $PROMETHEUS_URL"
    echo "  Temporal UI: $TEMPORAL_UI_URL"
    echo ""
    echo "To get Grafana admin password:"
    echo "  cd terraform && terraform output grafana_admin_password"
    echo ""
    echo "To get database credentials:"
    echo "  cd terraform && terraform output postgres_password"
    echo "  cd terraform && terraform output redis_password"
    echo ""
}

# Cleanup function
cleanup() {
    print_status "Cleaning up..."
    rm -f terraform/tfplan
}

# Main deployment function
main() {
    print_status "Starting Runbook Automation Engine deployment..."
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    # Check prerequisites
    check_prerequisites
    
    # Initialize Terraform
    init_terraform
    
    # Plan deployment
    plan_deployment
    
    # Ask for confirmation
    echo ""
    read -p "Do you want to proceed with the deployment? (y/N): " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        # Apply deployment
        apply_deployment
        
        # Show outputs
        show_outputs
        
        # Wait for pods
        wait_for_pods
        
        # Show access information
        show_access_info
        
        print_status "Deployment completed successfully!"
    else
        print_warning "Deployment cancelled."
        exit 0
    fi
}

# Handle script arguments
case "${1:-}" in
    "init")
        check_prerequisites
        init_terraform
        ;;
    "plan")
        check_prerequisites
        plan_deployment
        ;;
    "apply")
        check_prerequisites
        apply_deployment
        ;;
    "destroy")
        check_prerequisites
        cd terraform
        terraform destroy
        cd ..
        ;;
    "outputs")
        show_outputs
        ;;
    "help")
        echo "Usage: $0 [init|plan|apply|destroy|outputs|help]"
        echo ""
        echo "Commands:"
        echo "  init     - Initialize Terraform"
        echo "  plan     - Plan Terraform deployment"
        echo "  apply    - Apply Terraform deployment"
        echo "  destroy  - Destroy all resources"
        echo "  outputs  - Show deployment outputs"
        echo "  help     - Show this help message"
        echo ""
        echo "Default behavior: Full deployment with confirmation"
        ;;
    *)
        main
        ;;
esac
