# Runbook Automation Engine

##  Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (for local development)
- Node.js 18+ (for local development)

### Start with Docker Compose

1. **Clone and navigate to the project**
   ```bash
   cd "c:\Users\Gaurav Mishra\Downloads\AwesomeGRV Repo\Runbook Automation Engine"
   ```

2. **Start all services**
   ```bash
   docker-compose up -d
   ```

3. **Wait for services to be ready** (check health status)
   ```bash
   docker-compose ps
   ```

4. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Temporal UI: http://localhost:8088
   - Database: localhost:5432
   - Redis: localhost:6379

### Local Development

#### Backend
```bash
# Install dependencies
go mod download

# Run the server
go run cmd/server/main.go
```

#### Frontend
```bash
cd web

# Install dependencies
npm install

# Start development server
npm run dev
```

##  Project Structure

```
runbook-engine/
├── cmd/                    # Application entry points
│   └── server/            # Main server application
├── internal/              # Private application code
│   ├── api/              # API handlers
│   ├── config/           # Configuration management
│   ├── database/         # Database operations
│   ├── models/           # Data models
│   ├── services/         # Business logic
│   └── workers/          # Temporal workers
├── pkg/                   # Public library code
│   └── kubernetes/       # Kubernetes client
├── web/                   # React frontend
│   ├── src/
│   │   ├── components/   # React components
│   │   ├── pages/        # Page components
│   │   ├── contexts/     # React contexts
│   │   └── stores/       # State management
│   ├── package.json
│   └── vite.config.ts
├── migrations/            # Database migrations
├── build/                 # Dockerfiles
├── deploy/                # Kubernetes manifests
├── docker-compose.yml     # Development environment
└── config.yaml           # Configuration file
```

##  Available Features

###  Implemented
- **Backend API**: REST endpoints for runbooks, executions, triggers
- **Database Schema**: PostgreSQL with migrations
- **Kubernetes Integration**: Client for K8s operations
- **Temporal Workers**: Workflow execution engine
- **React Frontend**: Basic structure with Vite
- **Docker Setup**: Complete development environment

###  In Progress
- Frontend components (needs dependency installation)
- Authentication system
- Complete workflow builder UI
- Real-time execution monitoring

###  TODO
- Complete API handlers
- Add more worker types
- Implement authentication
- Build comprehensive UI
- Add monitoring and logging

##  Configuration

The application can be configured via:
1. `config.yaml` file
2. Environment variables
3. Command-line flags

Key configuration options:
- Database connection
- Redis connection
- Temporal connection
- JWT secret
- Kubernetes access

##  Database Schema

The application uses PostgreSQL with the following main tables:
- `users` - User management
- `teams` - Team/organization management
- `runbooks` - Runbook definitions
- `executions` - Runbook execution history
- `triggers` - Trigger configurations
- `integrations` - External integrations

##  Docker Services

- **postgres**: PostgreSQL 15 database
- **redis**: Redis 7 cache
- **temporal**: Temporal workflow engine
- **api**: Backend API server
- **frontend**: React development server

##  Testing

```bash
# Backend tests
go test ./...

# Frontend tests
cd web && npm test
```

##  Development Notes

### TypeScript Errors
The TypeScript errors in the frontend are expected because dependencies haven't been installed yet. To fix:

```bash
cd web
npm install
```

### Database Migrations
Database migrations run automatically when PostgreSQL starts. The initial schema is in `migrations/001_initial_schema.sql`.

### API Endpoints
- `GET /health` - Health check
- `GET /api/v1/runbooks` - List runbooks
- `POST /api/v1/runbooks` - Create runbook
- `POST /api/v1/runbooks/{id}/execute` - Execute runbook

##  Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

##  License

MIT License - see LICENSE file for details.
