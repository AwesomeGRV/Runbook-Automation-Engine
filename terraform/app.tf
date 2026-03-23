# Temporal deployment using Helm
resource "helm_release" "temporal" {
  name       = "temporal"
  repository = "https://helm.temporal.io"
  chart      = "temporal"
  namespace  = kubernetes_namespace.runbook_engine.metadata[0].name
  version    = var.temporal_helm_version

  set {
    name  = "server.replicaCount"
    value = var.temporal_server_replicas
  }

  set {
    name  = "server.config.persistence.enabled"
    value = true
  }

  set {
    name  = "server.config.persistence.default.driver"
    value = "sql"
  }

  set {
    name  = "server.config.persistence.sql.driver"
    value = "postgres12"
  }

  set {
    name  = "server.config.persistence.sql.host"
    value = kubernetes_service.postgres.metadata[0].name
  }

  set {
    name  = "server.config.persistence.sql.port"
    value = 5432
  }

  set {
    name  = "server.config.persistence.sql.user"
    value = "runbook_user"
  }

  set {
    name  = "server.config.persistence.sql.password"
    value = random_password.postgres_password.result
  }

  set {
    name  = "server.config.persistence.sql.database"
    value = "temporal"
  }

  set {
    name  = "server.config.persistence.sql.maxConns"
    value = 20
  }

  set {
    name  = "server.config.persistence.sql.maxIdleConns"
    value = 10
  }

  set {
    name  = "server.config.persistence.sql.maxConnLifetime"
    value = "1h"
  }

  set {
    name  = "web.enabled"
    value = true
  }

  set {
    name  = "web.replicaCount"
    value = 1
  }

  set {
    name  = "ui.enabled"
    value = true
  }

  set {
    name  = "ui.replicaCount"
    value = 1
  }

  set {
    name  = "admission-webhook.enabled"
    value = false
  }

  set {
    name  = "prometheus.enabled"
    value = true
  }

  set {
    name  = "grafana.enabled"
    value = false
  }

  set {
    name  = "elasticsearch.enabled"
    value = false
  }

  depends_on = [
    kubernetes_service.postgres
  ]
}

# Runbook Engine API Deployment
resource "kubernetes_deployment" "api" {
  metadata {
    name      = "runbook-api"
    namespace = kubernetes_namespace.runbook_engine.metadata[0].name
    labels = {
      app = "runbook-api"
    }
  }

  spec {
    replicas = var.api_replicas

    selector {
      match_labels = {
        app = "runbook-api"
      }
    }

    template {
      metadata {
        labels = {
          app = "runbook-api"
        }
      }

      spec {
        container {
          name  = "api"
          image = var.api_image

          port {
            container_port = 8080
            protocol      = "TCP"
          }

          env {
            name  = "DATABASE_HOST"
            value = kubernetes_service.postgres.metadata[0].name
          }

          env {
            name  = "DATABASE_PORT"
            value = "5432"
          }

          env {
            name  = "DATABASE_USER"
            value = "runbook_user"
          }

          env {
            name  = "DATABASE_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.postgres_credentials.metadata[0].name
                key  = "postgres-password"
              }
            }
          }

          env {
            name  = "DATABASE_NAME"
            value = "runbook_engine"
          }

          env {
            name  = "REDIS_HOST"
            value = kubernetes_service.redis.metadata[0].name
          }

          env {
            name  = "REDIS_PORT"
            value = "6379"
          }

          env {
            name  = "REDIS_PASSWORD"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.redis_credentials.metadata[0].name
                key  = "redis-password"
              }
            }
          }

          env {
            name  = "TEMPORAL_HOST"
            value = "${helm_release.temporal.name}-frontend.${kubernetes_namespace.runbook_engine.metadata[0].name}.svc.cluster.local"
          }

          env {
            name  = "TEMPORAL_PORT"
            value = "7233"
          }

          env {
            name  = "JWT_SECRET"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.jwt_secret.metadata[0].name
                key  = "jwt-secret"
              }
            }
          }

          env {
            name  = "ENVIRONMENT"
            value = var.environment
          }

          resources {
            requests = {
              cpu    = var.api_cpu_request
              memory = var.api_memory_request
            }
            limits = {
              cpu    = var.api_cpu_limit
              memory = var.api_memory_limit
            }
          }

          liveness_probe {
            http_get {
              path = "/health"
              port = 8080
            }
            initial_delay_seconds = 30
            period_seconds        = 10
            timeout_seconds       = 5
            failure_threshold     = 3
          }

          readiness_probe {
            http_get {
              path = "/health"
              port = 8080
            }
            initial_delay_seconds = 5
            period_seconds        = 5
            timeout_seconds       = 5
            failure_threshold     = 1
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "api" {
  metadata {
    name      = "runbook-api"
    namespace = kubernetes_namespace.runbook_engine.metadata[0].name
    labels = {
      app = "runbook-api"
    }
  }

  spec {
    selector = {
      app = "runbook-api"
    }

    port {
      port        = 80
      target_port = 8080
      protocol    = "TCP"
    }

    type = "ClusterIP"
  }
}

# Frontend Deployment
resource "kubernetes_deployment" "frontend" {
  metadata {
    name      = "runbook-frontend"
    namespace = kubernetes_namespace.runbook_engine.metadata[0].name
    labels = {
      app = "runbook-frontend"
    }
  }

  spec {
    replicas = var.frontend_replicas

    selector {
      match_labels = {
        app = "runbook-frontend"
      }
    }

    template {
      metadata {
        labels = {
          app = "runbook-frontend"
        }
      }

      spec {
        container {
          name  = "frontend"
          image = var.frontend_image

          port {
            container_port = 80
            protocol      = "TCP"
          }

          env {
            name  = "VITE_API_URL"
            value = "http://${kubernetes_service.api.metadata[0].name}.${kubernetes_namespace.runbook_engine.metadata[0].name}.svc.cluster.local"
          }

          resources {
            requests = {
              cpu    = var.frontend_cpu_request
              memory = var.frontend_memory_request
            }
            limits = {
              cpu    = var.frontend_cpu_limit
              memory = var.frontend_memory_limit
            }
          }

          liveness_probe {
            http_get {
              path = "/"
              port = 80
            }
            initial_delay_seconds = 30
            period_seconds        = 10
            timeout_seconds       = 5
            failure_threshold     = 3
          }

          readiness_probe {
            http_get {
              path = "/"
              port = 80
            }
            initial_delay_seconds = 5
            period_seconds        = 5
            timeout_seconds       = 5
            failure_threshold     = 1
          }
        }
      }
    }
  }
}

resource "kubernetes_service" "frontend" {
  metadata {
    name      = "runbook-frontend"
    namespace = kubernetes_namespace.runbook_engine.metadata[0].name
    labels = {
      app = "runbook-frontend"
    }
  }

  spec {
    selector = {
      app = "runbook-frontend"
    }

    port {
      port        = 80
      target_port = 80
      protocol    = "TCP"
    }

    type = "ClusterIP"
  }
}

# Ingress for external access
resource "kubernetes_ingress" "main" {
  metadata {
    name      = "runbook-engine-ingress"
    namespace = kubernetes_namespace.runbook_engine.metadata[0].name
    annotations = {
      "kubernetes.io/ingress.class"           = var.ingress_class
      "cert-manager.io/cluster-issuer"        = var.cluster_issuer
      "nginx.ingress.kubernetes.io/ssl-redirect": "true"
    }
  }

  spec {
    tls {
      hosts       = [var.domain_name]
      secret_name = "runbook-engine-tls"
    }

    rule {
      host = var.domain_name
      http {
        path {
          path = "/"
          backend {
            service_name = kubernetes_service.frontend.metadata[0].name
            service_port = 80
          }
        }
      }
    }

    rule {
      host = var.domain_name
      http {
        path {
          path = "/api"
          backend {
            service_name = kubernetes_service.api.metadata[0].name
            service_port = 80
          }
        }
      }
    }

    rule {
      host = var.domain_name
      http {
        path {
          path = "/temporal"
          backend {
            service_name = "${helm_release.temporal.name}-frontend"
            service_port = 8080
          }
        }
      }
    }
  }
}
