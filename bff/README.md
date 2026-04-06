# LinkPulse BFF (Backend for Frontend)

Standalone microservice that serves as an API gateway between the frontend and backend services.

## Project Structure

```
bff/
├── cmd/api/
│   └── main.go                    # Entry point
├── internal/
│   ├── adapters/http/            # HTTP clients for backend services
│   │   ├── shortener_client.go
│   │   ├── analytics_client.go
│   │   └── auth_client.go
│   ├── application/              # Business logic
│   │   └── bff_service.go
│   ├── domain/                   # Domain types
│   │   └── types.go
│   ├── ports/                    # Interfaces
│   │   └── clients.go
│   └── presentation/http/        # HTTP handlers & middleware
│       ├── handler.go
│       └── middleware.go
├── config/                       # Configuration files
├── Dockerfile                    # Container configuration
├── Makefile                      # Build automation
├── go.mod                        # Go module definition
└── README.md                     # This file
```

## Key Characteristics

### Independent Dependencies
- **Only 5 direct dependencies** (vs 40+ in monolithic backend)
- No shared backend module imports
- Lightweight and fast to build

### Standalone Architecture
- Completely independent `go.mod`
- Can be deployed separately
- Scales independently from backend
- Own CI/CD pipeline (when needed)

### Clean Boundaries
- HTTP adapter pattern for service communication
- Port/interface-based design
- No direct database access
- No internal service coupling

## Quick Start

### Prerequisites
- Go 1.25+
- Docker (optional)

### Local Development

```bash
cd bff

# Download dependencies
go mod download

# Build
make build

# Run
make run
```

Server starts at `http://localhost:8080`

### Environment Variables

```bash
# Core Configuration
export LINKPULSE_ENVIRONMENT=development
export LINKPULSE_SERVER_PORT=8080
export LINKPULSE_JWT_ACCESS_SECRET=your-secret-key

# Service URLs (with defaults)
export LINKPULSE_AUTH_SERVICE_URL=http://auth-service:8081
export LINKPULSE_SHORTENER_SERVICE_URL=http://shortener-service:8082
export LINKPULSE_ANALYTICS_SERVICE_URL=http://analytics-service:8083
```

**Default Service URLs:**
- Auth Service: `http://auth-service:8081`
- Shortener Service: `http://shortener-service:8082`
- Analytics Service: `http://analytics-service:8083`

### Docker

```bash
# Build image
make docker-build

# Run container
make docker-run
```

## API Endpoints

### Health Checks

```bash
GET /health           # Service health status
GET /readiness        # Ready to accept traffic
```

### Links Management

```bash
POST /api/v1/links                           # Create short link
GET /api/v1/links/:shortCode                 # Get link details
GET /api/v1/workspaces/:workspaceId/links    # List workspace links
DELETE /api/v1/links/:shortCode              # Delete link
```

### Analytics

```bash
GET /api/v1/links/:shortCode/analytics       # Get link analytics
GET /api/v1/workspaces/:workspaceId/analytics # Get workspace analytics
```

## Authentication

All API endpoints (except `/health` and `/readiness`) require JWT authentication:

```bash
Authorization: Bearer <JWT_TOKEN>
```

The JWT token must contain:
- `sub`: User ID
- `workspace_id`: Workspace ID

## Service Dependencies

BFF communicates with:
- **Shortener Service** (http://shortener-service:8082)
- **Analytics Service** (http://analytics-service:8083)
- **Auth Service** (http://auth-service:8081)

These can be configured via environment variables or docker-compose.

## Development

### Run Tests

```bash
make test
```

### Format Code

```bash
make fmt
```

### Build Binary

```bash
make build
# Binary at: ./bin/bff
```

## Deployment

### Docker Compose

The root `docker-compose.yml` includes the BFF service:

```bash
cd ..
docker-compose up -d bff
```

### Kubernetes

See `../BFF_DEPLOYMENT_OPERATIONS.md` for K8s manifests and deployment strategies.

## Monitoring

### Health Endpoint

```bash
curl http://localhost:8080/health
# Response:
# {
#   "status": "healthy",
#   "timestamp": "2026-04-05T10:30:00Z"
# }
```

### Readiness Endpoint

```bash
curl http://localhost:8080/readiness
# Response:
# {
#   "ready": true
# }
```

## Architecture Principles

1. **Separation of Concerns**: BFF only handles API gateway responsibilities
2. **No Business Logic**: All business logic in backend services
3. **Stateless**: No internal state, scales horizontally
4. **Interface-Based**: Easy to mock and test
5. **Minimal Dependencies**: Lightweight and fast

## Code Quality

### Recent Improvements (v2026.04)
- **Constants Extraction**: Eliminated duplicate string literals across adapters and services
  - `shortener_client.go`: 10 constants for endpoints, headers, and error messages
  - `bff_service.go`: 8 constants for validation and error formatting
- **DRY Principle**: Single source of truth for all error messages and configuration
- **Docker Optimization**: Merged consecutive RUN instructions to reduce image layers
- **JWT Middleware**: Robust token extraction and claim validation with proper error handling

## Contributing

1. Format code: `make fmt`
2. Run tests: `make test`
3. Build binary: `make build`
4. Verify: `make run`

## Debugging

### Enable Debug Logging

```bash
LINKPULSE_ENVIRONMENT=development make run
```

### View Logs

```bash
# Docker
docker logs linkpulse-bff

# Local
LINKPULSE_ENVIRONMENT=development ./bin/bff 2>&1 | grep ERROR
```

## Documentation

- **Strategy**: See `../BFF_STANDALONE_STRATEGY.md`
- **Implementation**: See `../BFF_IMPLEMENTATION_GUIDE.md`
- **Deployment**: See `../BFF_DEPLOYMENT_OPERATIONS.md`

## License

See `../LICENSE`

---

**Last Updated**: 2026-04-05
