terraform {
  required_version = ">= 1.0"
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.20"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.9"
    }
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.5"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 3.2"
    }
  }
}

provider "kubernetes" {
  config_path = var.kubeconfig_path
  config_context = var.kubeconfig_context
}

provider "helm" {
  kubernetes {
    config_path = var.kubeconfig_path
    config_context = var.kubeconfig_context
  }
}

provider "aws" {
  region = var.aws_region
}

# Random suffix for unique resource names
resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}

# Namespace for the application
resource "kubernetes_namespace" "runbook_engine" {
  metadata {
    name = "${var.namespace}-${random_string.suffix.result}"
    labels = {
      app     = "runbook-engine"
      managed = "terraform"
    }
  }
}

# PostgreSQL password secret
resource "random_password" "postgres_password" {
  length  = 32
  special = true
}

resource "kubernetes_secret" "postgres_credentials" {
  metadata {
    name      = "postgres-credentials"
    namespace = kubernetes_namespace.runbook_engine.metadata[0].name
  }

  data = {
    postgres-password = base64encode(random_password.postgres_password.result)
    postgres-user     = base64encode("runbook_user")
    postgres-db       = base64encode("runbook_engine")
  }

  type = "Opaque"
}

# Redis password secret
resource "random_password" "redis_password" {
  length  = 32
  special = true
}

resource "kubernetes_secret" "redis_credentials" {
  metadata {
    name      = "redis-credentials"
    namespace = kubernetes_namespace.runbook_engine.metadata[0].name
  }

  data = {
    redis-password = base64encode(random_password.redis_password.result)
  }

  type = "Opaque"
}

# JWT secret
resource "random_password" "jwt_secret" {
  length  = 64
  special = true
}

resource "kubernetes_secret" "jwt_secret" {
  metadata {
    name      = "jwt-secret"
    namespace = kubernetes_namespace.runbook_engine.metadata[0].name
  }

  data = {
    jwt-secret = base64encode(random_password.jwt_secret.result)
  }

  type = "Opaque"
}

# PostgreSQL StatefulSet
resource "kubernetes_persistent_volume_claim" "postgres_pvc" {
  metadata {
    name      = "postgres-pvc"
    namespace = kubernetes_namespace.runbook_engine.metadata[0].name
  }

  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = var.postgres_storage_size
      }
    }
    storage_class_name = var.storage_class_name
  }
}

resource "kubernetes_statefulset" "postgres" {
  metadata {
    name      = "postgres"
    namespace = kubernetes_namespace.runbook_engine.metadata[0].name
    labels = {
      app = "postgres"
    }
  }

  spec {
    service_name = "postgres"
    replicas     = 1

    selector {
      match_labels = {
        app = "postgres"
      }
    }

    template {
      metadata {
        labels = {
          app = "postgres"
        }
      }

      spec {
        container {
          name  = "postgres"
          image = "postgres:15"

          port {
            container_port = 5432
            protocol      = "TCP"
          }

          env {
            name  = "POSTGRES_DB"
            value = "runbook_engine"
          }

          env {
            name  = "POSTGRES_USER"
            value = "runbook_user"
          }

          env {
            name  = "POSTGRES_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.postgres_credentials.metadata[0].name
                key  = "postgres-password"
              }
            }
          }

          env {
            name  = "PGDATA"
            value = "/var/lib/postgresql/data/pgdata"
          }

          volume_mount {
            name       = "postgres-storage"
            mount_path  = "/var/lib/postgresql/data"
          }

          resources {
            requests = {
              cpu    = var.postgres_cpu_request
              memory = var.postgres_memory_request
            }
            limits = {
              cpu    = var.postgres_cpu_limit
              memory = var.postgres_memory_limit
            }
          }

          liveness_probe {
            exec {
              command = ["/bin/sh", "-c", "pg_isready -U runbook_user -d runbook_engine"]
            }
            initial_delay_seconds = 30
            period_seconds        = 10
            timeout_seconds       = 5
            failure_threshold     = 3
          }

          readiness_probe {
            exec {
              command = ["/bin/sh", "-c", "pg_isready -U runbook_user -d runbook_engine"]
            }
            initial_delay_seconds = 5
            period_seconds        = 5
            timeout_seconds       = 5
            failure_threshold     = 1
          }
        }
      }

      volume {
        name = "postgres-storage"
        persistent_volume_claim {
          claim_name = kubernetes_persistent_volume_claim.postgres_pvc.metadata[0].name
        }
      }
    }
  }
}

resource "kubernetes_service" "postgres" {
  metadata {
    name      = "postgres"
    namespace = kubernetes_namespace.runbook_engine.metadata[0].name
    labels = {
      app = "postgres"
    }
  }

  spec {
    selector = {
      app = "postgres"
    }

    port {
      port        = 5432
      target_port = 5432
      protocol    = "TCP"
    }

    type = "ClusterIP"
  }
}

# Redis StatefulSet
resource "kubernetes_persistent_volume_claim" "redis_pvc" {
  metadata {
    name      = "redis-pvc"
    namespace = kubernetes_namespace.runbook_engine.metadata[0].name
  }

  spec {
    access_modes = ["ReadWriteOnce"]
    resources {
      requests = {
        storage = var.redis_storage_size
      }
    }
    storage_class_name = var.storage_class_name
  }
}

resource "kubernetes_statefulset" "redis" {
  metadata {
    name      = "redis"
    namespace = kubernetes_namespace.runbook_engine.metadata[0].name
    labels = {
      app = "redis"
    }
  }

  spec {
    service_name = "redis"
    replicas     = 1

    selector {
      match_labels = {
        app = "redis"
      }
    }

    template {
      metadata {
        labels = {
          app = "redis"
        }
      }

      spec {
        container {
          name  = "redis"
          image = "redis:7-alpine"

          port {
            container_port = 6379
            protocol      = "TCP"
          }

          command = ["redis-server"]
          args    = ["--requirepass", "$(REDIS_PASSWORD)", "--appendonly", "yes"]

          env {
            name  = "REDIS_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.redis_credentials.metadata[0].name
                key  = "redis-password"
              }
            }
          }

          volume_mount {
            name       = "redis-storage"
            mount_path  = "/data"
          }

          resources {
            requests = {
              cpu    = var.redis_cpu_request
              memory = var.redis_memory_request
            }
            limits = {
              cpu    = var.redis_cpu_limit
              memory = var.redis_memory_limit
            }
          }

          liveness_probe {
            exec {
              command = ["redis-cli", "--raw", "incr", "ping"]
            }
            initial_delay_seconds = 30
            period_seconds        = 10
            timeout_seconds       = 5
            failure_threshold     = 3
          }

          readiness_probe {
            exec {
              command = ["redis-cli", "--raw", "incr", "ping"]
            }
            initial_delay_seconds = 5
            period_seconds        = 5
            timeout_seconds       = 5
            failure_threshold     = 1
          }
        }
      }

      volume {
        name = "redis-storage"
        persistent_volume_claim {
          claim_name = kubernetes_persistent_volume_claim.redis_pvc.metadata[0].name
        }
      }
    }
  }
}

resource "kubernetes_service" "redis" {
  metadata {
    name      = "redis"
    namespace = kubernetes_namespace.runbook_engine.metadata[0].name
    labels = {
      app = "redis"
    }
  }

  spec {
    selector = {
      app = "redis"
    }

    port {
      port        = 6379
      target_port = 6379
      protocol    = "TCP"
    }

    type = "ClusterIP"
  }
}
