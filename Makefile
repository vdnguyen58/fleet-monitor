.PHONY: build test coverage lint vulncheck clean all help docker-build docker-run docker-stop docker-clean docker-logs docker-e2e

# Docker variables
DOCKER_IMAGE_NAME = fleet-monitor
DOCKER_CONTAINER_NAME = fleet-monitor-container
DOCKER_PORT = 6733

# Default target
all: lint test build

# Build the application
build:
	@echo "Building fleet-monitor..."
	@go build -o bin/fleet-monitor .
	@echo "Build complete: bin/fleet-monitor"

# Run tests
test:
	@echo "Running tests..."
	@go test ./... -v

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | grep total
	@echo "Coverage report generated: coverage.html"

# Run linter (requires golangci-lint to be installed)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run ./...; \
	elif [ -f $(shell go env GOPATH)/bin/golangci-lint ]; then \
		$(shell go env GOPATH)/bin/golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		$(shell go env GOPATH)/bin/golangci-lint run ./...; \
	fi

# Run govulncheck (requires govulncheck to be installed)
vulncheck:
	@echo "Running vulnerability check..."
	@if command -v govulncheck > /dev/null 2>&1; then \
		govulncheck ./...; \
	elif [ -f $(shell go env GOPATH)/bin/govulncheck ]; then \
		$(shell go env GOPATH)/bin/govulncheck ./...; \
	else \
		echo "govulncheck not found. Installing..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
		$(shell go env GOPATH)/bin/govulncheck ./...; \
	fi

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Run the application
run: build
	@echo "Running fleet-monitor..."
	@./bin/fleet-monitor

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

# Run all checks (lint, test, vulncheck)
check: lint test vulncheck

# Docker targets

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE_NAME):latest .
	@echo "Docker image built: $(DOCKER_IMAGE_NAME):latest"

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -d \
		--name $(DOCKER_CONTAINER_NAME) \
		-p $(DOCKER_PORT):$(DOCKER_PORT) \
		$(DOCKER_IMAGE_NAME):latest
	@echo "Container started: $(DOCKER_CONTAINER_NAME)"
	@echo "Application available at http://localhost:$(DOCKER_PORT)"

# Stop Docker container
docker-stop:
	@echo "Stopping Docker container..."
	@docker stop $(DOCKER_CONTAINER_NAME) 2>/dev/null || true
	@docker rm $(DOCKER_CONTAINER_NAME) 2>/dev/null || true
	@echo "Container stopped and removed"

# Clean Docker resources
docker-clean: docker-stop
	@echo "Removing Docker image..."
	@docker rmi $(DOCKER_IMAGE_NAME):latest 2>/dev/null || true
	@echo "Docker resources cleaned"

# View Docker container logs
docker-logs:
	@docker logs -f $(DOCKER_CONTAINER_NAME)

# Run e2e tests against Docker container
docker-e2e: docker-build
	@echo "Running e2e tests against Docker container..."
	@docker stop $(DOCKER_CONTAINER_NAME) 2>/dev/null || true
	@docker rm $(DOCKER_CONTAINER_NAME) 2>/dev/null || true
	@docker run -d \
		--name $(DOCKER_CONTAINER_NAME) \
		-p $(DOCKER_PORT):$(DOCKER_PORT) \
		$(DOCKER_IMAGE_NAME):latest
	@echo "Waiting for container to be ready..."
	@sleep 3
	@echo "Running Docker e2e tests..."
	@go test -v -run TestDocker ./...
	@echo "Stopping container..."
	@docker stop $(DOCKER_CONTAINER_NAME) 2>/dev/null || true
	@docker rm $(DOCKER_CONTAINER_NAME) 2>/dev/null || true
	@echo "Docker e2e tests complete"

# Help target
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build & Run:"
	@echo "  make build        - Build the application"
	@echo "  make run          - Build and run the application"
	@echo ""
	@echo "Testing:"
	@echo "  make test         - Run tests"
	@echo "  make coverage     - Run tests with coverage report"
	@echo "  make docker-e2e   - Run e2e tests against Docker container"
	@echo ""
	@echo "Code Quality:"
	@echo "  make lint         - Run linter (golangci-lint)"
	@echo "  make vulncheck    - Run vulnerability check (govulncheck)"
	@echo "  make fmt          - Format code"
	@echo "  make check        - Run all checks (lint, test, vulncheck)"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker container"
	@echo "  make docker-stop  - Stop Docker container"
	@echo "  make docker-clean - Remove Docker container and image"
	@echo "  make docker-logs  - View Docker container logs"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make deps         - Install/update dependencies"
	@echo "  make all          - Run lint, test, and build (default)"
	@echo "  make help         - Show this help message"
