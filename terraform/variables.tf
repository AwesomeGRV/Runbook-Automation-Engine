# Variables for Terraform configuration

variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-west-2"
}

variable "kubeconfig_path" {
  description = "Path to kubeconfig file"
  type        = string
  default     = "~/.kube/config"
}

variable "kubeconfig_context" {
  description = "Kubernetes context to use"
  type        = string
  default     = ""
}

variable "namespace" {
  description = "Base namespace name"
  type        = string
  default     = "runbook-engine"
}

variable "environment" {
  description = "Environment (dev, staging, prod)"
  type        = string
  default     = "production"
}

variable "domain_name" {
  description = "Domain name for the application"
  type        = string
  default     = "runbook-engine.example.com"
}

variable "ingress_class" {
  description = "Ingress class to use"
  type        = string
  default     = "nginx"
}

variable "cluster_issuer" {
  description = "Cluster issuer for TLS certificates"
  type        = string
  default     = "letsencrypt-prod"
}

variable "storage_class_name" {
  description = "Storage class for persistent volumes"
  type        = string
  default     = "gp2"
}

# Database configuration
variable "postgres_storage_size" {
  description = "Storage size for PostgreSQL"
  type        = string
  default     = "20Gi"
}

variable "postgres_cpu_request" {
  description = "CPU request for PostgreSQL"
  type        = string
  default     = "250m"
}

variable "postgres_cpu_limit" {
  description = "CPU limit for PostgreSQL"
  type        = string
  default     = "500m"
}

variable "postgres_memory_request" {
  description = "Memory request for PostgreSQL"
  type        = string
  default     = "256Mi"
}

variable "postgres_memory_limit" {
  description = "Memory limit for PostgreSQL"
  type        = string
  default     = "512Mi"
}

# Redis configuration
variable "redis_storage_size" {
  description = "Storage size for Redis"
  type        = string
  default     = "8Gi"
}

variable "redis_cpu_request" {
  description = "CPU request for Redis"
  type        = string
  default     = "100m"
}

variable "redis_cpu_limit" {
  description = "CPU limit for Redis"
  type        = string
  default     = "200m"
}

variable "redis_memory_request" {
  description = "Memory request for Redis"
  type        = string
  default     = "128Mi"
}

variable "redis_memory_limit" {
  description = "Memory limit for Redis"
  type        = string
  default     = "256Mi"
}

# API configuration
variable "api_image" {
  description = "Docker image for API"
  type        = string
  default     = "runbook-engine/api:latest"
}

variable "api_replicas" {
  description = "Number of API replicas"
  type        = number
  default     = 2
}

variable "api_min_replicas" {
  description = "Minimum number of API replicas for HPA"
  type        = number
  default     = 1
}

variable "api_max_replicas" {
  description = "Maximum number of API replicas for HPA"
  type        = number
  default     = 10
}

variable "api_cpu_request" {
  description = "CPU request for API"
  type        = string
  default     = "100m"
}

variable "api_cpu_limit" {
  description = "CPU limit for API"
  type        = string
  default     = "500m"
}

variable "api_memory_request" {
  description = "Memory request for API"
  type        = string
  default     = "128Mi"
}

variable "api_memory_limit" {
  description = "Memory limit for API"
  type        = string
  default     = "512Mi"
}

variable "api_cpu_target_utilization" {
  description = "CPU target utilization for HPA"
  type        = number
  default     = 70
}

variable "api_memory_target_utilization" {
  description = "Memory target utilization for HPA"
  type        = number
  default     = 80
}

# Frontend configuration
variable "frontend_image" {
  description = "Docker image for frontend"
  type        = string
  default     = "runbook-engine/frontend:latest"
}

variable "frontend_replicas" {
  description = "Number of frontend replicas"
  type        = number
  default     = 2
}

variable "frontend_cpu_request" {
  description = "CPU request for frontend"
  type        = string
  default     = "50m"
}

variable "frontend_cpu_limit" {
  description = "CPU limit for frontend"
  type        = string
  default     = "200m"
}

variable "frontend_memory_request" {
  description = "Memory request for frontend"
  type        = string
  default     = "64Mi"
}

variable "frontend_memory_limit" {
  description = "Memory limit for frontend"
  type        = string
  default     = "256Mi"
}

# Temporal configuration
variable "temporal_helm_version" {
  description = "Temporal Helm chart version"
  type        = string
  default     = "0.39.0"
}

variable "temporal_server_replicas" {
  description = "Number of Temporal server replicas"
  type        = number
  default     = 1
}

# Monitoring configuration
variable "prometheus_helm_version" {
  description = "Prometheus Helm chart version"
  type        = string
  default     = "48.1.0"
}

variable "prometheus_storage_size" {
  description = "Storage size for Prometheus"
  type        = string
  default     = "50Gi"
}

variable "grafana_admin_password" {
  description = "Grafana admin password"
  type        = string
  default     = "admin123"
  sensitive   = true
}
