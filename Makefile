.PHONY: build test coverage lint vulncheck clean all help

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

# Help target
help:
	@echo "Available targets:"
	@echo "  make build      - Build the application"
	@echo "  make test       - Run tests"
	@echo "  make coverage   - Run tests with coverage report"
	@echo "  make lint       - Run linter (golangci-lint)"
	@echo "  make vulncheck  - Run vulnerability check (govulncheck)"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make run        - Build and run the application"
	@echo "  make deps       - Install/update dependencies"
	@echo "  make fmt        - Format code"
	@echo "  make check      - Run all checks (lint, test, vulncheck)"
	@echo "  make all        - Run lint, test, and build (default)"
	@echo "  make help       - Show this help message"
