# Tech Stack & Build System

## Technology Stack

### Backend
- **Language**: Go 1.21
- **HTTP Framework**: Go standard library (`net/http`)
- **Dependencies**: Minimal (only `github.com/gorilla/mux` in go.mod, though not currently used)
- **Testing**: Go standard `testing` package

### Frontend
- **Markup**: HTML5
- **Styling**: Tailwind CSS (CDN)
- **JavaScript**: Vanilla JavaScript (no frameworks)

### Deployment & Infrastructure
- **Containerization**: Docker (multi-stage build)
- **Platforms**: GCP Cloud Run, AWS (App Runner, ECS Fargate, Lambda)
- **Base Images**: 
  - Build: `golang:1.21-alpine`
  - Runtime: `alpine:latest`

## Build & Development Commands

### Local Development

```bash
# Download dependencies
go mod download

# Run the application locally
go run main.go

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Docker

```bash
# Build Docker image
docker build -t url-checker .

# Run container locally
docker run -p 8080:8080 url-checker

# Run with custom port
docker run -p 9000:8080 -e PORT=8080 url-checker
```

### Deployment

```bash
# GCP Cloud Run (direct deployment)
gcloud run deploy url-checker \
  --source . \
  --platform managed \
  --region asia-east1 \
  --allow-unauthenticated

# View Cloud Run logs
gcloud run services logs read url-checker --region asia-east1

# AWS App Runner (after pushing to ECR)
aws apprunner create-service \
  --service-name url-checker \
  --source-configuration ImageRepository={ImageIdentifier=...}
```

## Project Configuration

### Environment Variables

| Variable | Default | Purpose |
|----------|---------|---------|
| `PORT` | `8080` | HTTP server listening port |

### Go Module

- **Module Name**: `url-checker`
- **Go Version**: 1.21
- **Dependencies**: Minimal (gorilla/mux available but not required)

## Code Style & Conventions

### Go Code Style

- Follow standard Go conventions (gofmt)
- Use `CamelCase` for exported functions/types
- Use `snake_case` for JSON field tags
- Keep functions focused and small
- Use error handling with explicit checks (no panic in production code)
- Prefer standard library over external packages

### Naming Conventions

- **Types**: `TestRequest`, `TestResponse` (PascalCase)
- **Functions**: `testURL`, `validateURL`, `formatError` (camelCase)
- **Constants**: `REQUEST_TIMEOUT` (UPPER_SNAKE_CASE)
- **JSON fields**: `statusCode`, `responseTime`, `finalUrl` (camelCase)

### File Organization

- Single `main.go` file for all backend logic
- Static files in `static/` directory
- Tests in `*_test.go` files alongside source code
- No subdirectories for Go packages (keep it flat)

## Testing Strategy

### Unit Tests

- Test validation logic (`validateURL`, `isBlocked`)
- Test error handling and formatting
- Test response parsing and truncation
- Use `httptest.Server` for HTTP testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific test
go test -run TestValidateURL

# Run with coverage report
go test -cover ./...

# Generate coverage HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Coverage Goals

- Core logic: > 80%
- Error handling: > 90%
- API endpoints: 100%

## Performance Targets

- API response time: < 100ms (excluding target URL request time)
- Static file serving: < 50ms
- HTTP request timeout: 30 seconds
- Response body preview limit: 1000 characters

## Security Considerations

- URL validation required before making requests
- User-Agent header set to avoid blocking
- SSL/TLS errors handled gracefully
- Error messages don't expose internal system details
- No sensitive data logged to stdout
- Errors logged to stderr for debugging

## Dependency Management

- Minimize external dependencies
- Prefer Go standard library
- Use `go mod tidy` to clean up unused dependencies
- Lock dependencies with `go.sum`
- Update dependencies carefully and test thoroughly
