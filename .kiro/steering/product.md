# Product Overview

## URL Tester

A lightweight web application that tests URL accessibility and retrieves detailed HTTP response information. Users can check if a URL is reachable, view HTTP status codes, response headers, response time, and detect if requests are being blocked.

### Key Features

- HTTP status code and response time measurement
- Response headers inspection
- Response body preview (first 1000 characters)
- Redirect chain tracking
- Block detection (403/429 status codes)
- SSL/TLS error reporting
- Responsive web interface
- Cloud-ready deployment (GCP Cloud Run, AWS)

### Target Users

- Developers debugging backend services
- DevOps engineers monitoring service availability
- System administrators testing URL accessibility

### Tech Stack

- **Backend**: Go 1.21 (standard library)
- **Frontend**: HTML5 + Tailwind CSS (CDN)
- **Deployment**: Docker, GCP Cloud Run, AWS
- **Testing**: Go testing package

### Architecture

Single-tier stateless application with no database. Backend serves both API endpoints and static frontend files. Designed for containerized deployment with automatic scaling.
