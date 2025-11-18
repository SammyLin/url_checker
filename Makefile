.PHONY: build run test test-coverage docker-build docker-run docker-push help

# Variables
BINARY_NAME=url-tester
DOCKER_IMAGE=sammylin/url_checker
DOCKER_TAG?=latest

# Default target
help:
	@echo "URL Tester - Available Commands"
	@echo ""
	@echo "Local Development:"
	@echo "  make build              - Compile Go application"
	@echo "  make run                - Run application locally"
	@echo "  make test               - Run all tests"
	@echo "  make test-coverage      - Run tests with coverage report"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build       - Build Docker image"
	@echo "  make docker-run         - Run Docker container locally"
	@echo "  make docker-push        - Push Docker image to registry"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean              - Remove compiled binary"
	@echo "  make help               - Show this help message"

# Build the Go application
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"

# Run the application locally
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BINARY_NAME)

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build Docker image
docker-build:
	@echo "Building Docker image: $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest
	@echo "Docker image built successfully"

# Run Docker container locally
docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

# Push Docker image to registry
docker-push:
	@echo "Pushing Docker image to registry..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest
	@echo "Docker image pushed successfully"

# Clean up
clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "Clean complete"
