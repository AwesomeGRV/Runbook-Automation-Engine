# Runbook Automation Engine

A production-grade runbook automation platform for SRE teams that enables visual creation, management, and execution of automated runbooks with one-click incident resolution.

## Features

- **Visual Workflow Builder**: Drag-and-drop interface for creating runbooks
- **Multi-Trigger Support**: Manual, alert-based, scheduled, and ChatOps triggers
- **Kubernetes Integration**: Native K8s operations (restart, scale, rollback)
- **Temporal Workflow Engine**: Durable, scalable workflow execution
- **Enterprise Security**: RBAC, secret management, audit logs
- **Real-time Monitoring**: Metrics, logs, and distributed tracing
- **One-Click Resolution**: Automated incident response workflows

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React UI      │    │   API Gateway   │    │  Microservices  │
│                 │◄──►│                 │◄──►│                 │
│ - Workflow      │    │ - Auth/Rate     │    │ - Runbook       │
│   Builder       │    │   Limiting      │    │ - Execution     │
│ - Dashboard     │    │ - Routing       │    │ - Trigger       │
│ - Monitoring    │    │                 │    │ - Integration   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                       │
                                                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Temporal      │    │   PostgreSQL    │    │   Kubernetes    │
│                 │    │                 │    │                 │
│ - Workflow      │    │ - Runbooks      │    │ - Deployments   │
│   Orchestration │    │ - Executions    │    │ - Services      │
│ - Workers       │    │ - Users         │    │ - Pods          │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Technology Stack

- **Frontend**: React 18, TypeScript, React Flow, Tailwind CSS
- **Backend**: Go, gRPC, REST APIs
- **Workflow Engine**: Temporal
- **Database**: PostgreSQL, Redis
- **Container**: Docker, Kubernetes
- **Monitoring**: Prometheus, Grafana, Jaeger
- **Security**: HashiCorp Vault, RBAC

## Project Structure

```
runbook-engine/
├── cmd/                    # Application entry points
├── internal/               # Private application code
│   ├── api/               # API handlers
│   ├── services/          # Business logic
│   ├── workers/           # Temporal workers
│   └── models/            # Data models
├── pkg/                   # Public library code
├── web/                   # React frontend
├── deploy/                # Kubernetes manifests
├── helm/                  # Helm charts
├── build/                 # Dockerfiles
├── migrations/            # Database migrations
└── docs/                  # Documentation
```

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- Docker & Kubernetes
- PostgreSQL 15+
- Redis 7+
- Temporal

### Local Development

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd runbook-engine
   ```

2. **Start infrastructure**
   ```bash
   docker-compose up -d postgres redis temporal
   ```

3. **Run database migrations**
   ```bash
   go run cmd/migrate/main.go up
   ```

4. **Start backend services**
   ```bash
   go run cmd/server/main.go
   ```

5. **Start frontend**
   ```bash
   cd web
   npm install
   npm run dev
   ```

6. **Access the application**
   - Frontend: http://localhost:3000
   - API: http://localhost:8080
   - Temporal UI: http://localhost:8088

## Documentation

- [Architecture Overview](./docs/architecture.md)
- [API Documentation](./docs/api.md)
- [Development Guide](./docs/development.md)
- [Deployment Guide](./docs/deployment.md)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.
