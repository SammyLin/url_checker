# Project Structure

## Directory Layout

```
url-checker/
├── main.go                 # Backend API server and core logic
├── main_test.go            # Unit and integration tests
├── go.mod                  # Go module definition
├── go.sum                  # Go dependency lock file
├── Dockerfile              # Multi-stage Docker build configuration
├── deploy.sh               # Deployment helper script
├── README.md               # Project documentation (Chinese)
├── .dockerignore            # Docker build exclusions
├── .gcloudignore           # Google Cloud exclusions
├── static/
│   └── index.html          # Frontend HTML with embedded CSS/JS
└── url-tester              # Compiled binary (generated)
```

## File Descriptions

### Backend Files

#### `main.go`
- **Purpose**: Single file containing all backend logic
- **Responsibilities**:
  - HTTP server setup and routing
  - Request handlers (`testURLHandler`, `healthHandler`, `serveStaticHandler`)
  - URL validation logic
  - HTTP client creation and request execution
  - Error formatting and response building
  - Data structures (`TestRequest`, `TestResponse`)
- **Key Functions**:
  - `main()`: Server initialization
  - `testURLHandler()`: POST /api/test endpoint
  - `testURL()`: Core URL testing logic
  - `validateURL()`: Input validation
  - `formatError()`: Error message formatting
  - `isBlocked()`: Block detection (403/429)
  - `createHTTPClient()`: HTTP client configuration

#### `main_test.go`
- **Purpose**: Comprehensive test suite
- **Test Coverage**:
  - URL validation tests
  - Block detection tests
  - HTTP response handling
  - Error handling and formatting
  - API endpoint tests
  - Redirect tracking
  - Body truncation
  - User-Agent header verification
- **Test Utilities**: Uses `httptest.Server` for mocking HTTP endpoints

### Configuration Files

#### `go.mod` & `go.sum`
- **go.mod**: Declares module name (`url-checker`), Go version (1.21), and dependencies
- **go.sum**: Locks dependency versions for reproducible builds
- **Current Dependencies**: Minimal (gorilla/mux available but unused)

#### `Dockerfile`
- **Strategy**: Multi-stage build for minimal image size
- **Build Stage**: `golang:1.21-alpine` - compiles Go binary
- **Runtime Stage**: `alpine:latest` - runs compiled binary
- **Includes**: CA certificates for HTTPS, static files, compiled binary
- **Exposes**: Port 8080
- **Entry Point**: `./main`

### Frontend Files

#### `static/index.html`
- **Purpose**: Single HTML file with embedded CSS and JavaScript
- **Contents**:
  - HTML structure for URL input form
  - Tailwind CSS styling (via CDN)
  - Vanilla JavaScript for form handling and API calls
  - Result display formatting
  - Loading indicators
  - Error message display
- **No Build Step**: Served as-is by the backend

### Documentation & Scripts

#### `README.md`
- **Language**: Chinese (Traditional)
- **Sections**:
  - Feature overview
  - Local development setup
  - Docker usage
  - GCP Cloud Run deployment (3 methods)
  - AWS deployment options
  - API endpoint documentation
  - Environment variables
  - Cost estimation
  - Troubleshooting guide
  - Security recommendations

#### `deploy.sh`
- **Purpose**: Helper script for deployment automation
- **Usage**: Simplifies deployment commands

#### `.dockerignore` & `.gcloudignore`
- **Purpose**: Exclude unnecessary files from Docker/Cloud builds
- **Typical Exclusions**: Git files, test files, documentation

## Code Organization Principles

### Single File Architecture
- All Go code in `main.go` for simplicity
- No package subdirectories
- Easy to understand and modify
- Suitable for small to medium applications

### Separation of Concerns
Within `main.go`:
- **Data Models**: `TestRequest`, `TestResponse` structs
- **Validation**: `validateURL()` function
- **HTTP Client**: `createHTTPClient()` function
- **Core Logic**: `testURL()` function
- **Handlers**: `testURLHandler()`, `healthHandler()`, `serveStaticHandler()`
- **Utilities**: `formatError()`, `isBlocked()`

### Frontend Integration
- Static files served from `static/` directory
- No build process required
- Embedded CSS and JavaScript in single HTML file
- Tailwind CSS loaded from CDN

## API Endpoints

### `GET /`
- **Purpose**: Serve frontend HTML
- **Handler**: `serveStaticHandler()`
- **Response**: `static/index.html`

### `POST /api/test`
- **Purpose**: Test a URL
- **Handler**: `testURLHandler()`
- **Request Body**: JSON with `url` field
- **Response**: JSON with test results

### `GET /health`
- **Purpose**: Health check for container orchestration
- **Handler**: `healthHandler()`
- **Response**: Plain text "OK"

## Build Artifacts

### Compiled Binary
- **Name**: `url-tester` (or `url-checker` in Docker)
- **Generated**: By `go build` or Docker build
- **Size**: ~10-15 MB (Alpine-based)
- **Platform**: Linux (for Docker/Cloud deployment)

### Docker Image
- **Name**: `url-checker` (default)
- **Size**: ~20-30 MB (Alpine runtime)
- **Layers**: 2 (build + runtime)

## Development Workflow

1. **Edit Code**: Modify `main.go` or `main_test.go`
2. **Run Locally**: `go run main.go`
3. **Test**: `go test ./...`
4. **Build Binary**: `go build -o url-tester`
5. **Docker Build**: `docker build -t url-checker .`
6. **Deploy**: Use `gcloud run deploy` or AWS CLI

## Testing Structure

### Test File Organization
- Tests in `main_test.go` alongside source code
- Test functions follow Go naming: `TestFunctionName()`
- Table-driven tests for multiple scenarios
- Uses `httptest.Server` for HTTP mocking

### Test Categories
- **Unit Tests**: Validation, error handling, block detection
- **Integration Tests**: Full request-response cycle
- **Handler Tests**: API endpoint behavior

## Deployment Artifacts

### Docker Image Contents
```
/root/
├── main              # Compiled Go binary
└── static/
    └── index.html    # Frontend files
```

### Environment at Runtime
- **Working Directory**: `/root/`
- **Port**: 8080 (configurable via PORT env var)
- **Logging**: stdout (info), stderr (errors)
- **No Persistent Storage**: Stateless design

## Key Design Decisions

1. **Single File**: Simplicity over modularity for this scale
2. **No Database**: Stateless, each request independent
3. **Standard Library**: Minimal dependencies, easier deployment
4. **Alpine Linux**: Smaller image size, faster deployment
5. **Embedded Frontend**: No separate frontend build/deployment
6. **Multi-stage Docker**: Optimized image size
