.PHONY: build run clean test docker-build docker-up docker-down docker-logs \
        ci ci-test ci-lint ci-build coverage check deploy help

# Build the application
build:
	@echo "Building application..."
	@go build -o bin/server ./cmd/server
	@echo "Build complete: bin/server"

# Run the application
run:
	@echo "Running application..."
	@go run ./cmd/server

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./... || echo "Install golangci-lint for linting"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Development mode (with auto-restart)
dev:
	@echo "Starting in development mode..."
	@go run ./cmd/server

# Docker commands
docker-build:
	@echo "Building Docker image..."
	@docker-compose build

docker-up:
	@echo "Starting Docker containers..."
	@docker-compose up -d
	@echo "Services started. Access the application at http://localhost:8080"

docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down

docker-logs:
	@echo "Following logs..."
	@docker-compose logs -f app

docker-restart:
	@echo "Restarting Docker containers..."
	@docker-compose restart

docker-clean:
	@echo "Stopping and removing containers, volumes..."
	@docker-compose down -v

# CI/CD targets
ci: ci-lint ci-test ci-build
	@echo "CI pipeline completed successfully"

ci-test:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out

ci-lint:
	@echo "Running linters..."
	@go fmt ./...
	@go vet ./...
	@test -z "$$(gofmt -l .)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)

ci-build:
	@echo "Building for CI..."
	@CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/server ./cmd/server
	@echo "Build successful"

coverage:
	@echo "Generating coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

check: fmt lint test
	@echo "Pre-commit checks passed"

# Deployment targets
deploy-docker:
	@echo "Deploying with Docker Compose..."
	@docker-compose pull
	@docker-compose up -d --build
	@echo "Deployment complete"

deploy-production:
	@echo "Building production binary..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/server-linux-amd64 ./cmd/server
	@echo "Production build complete: bin/server-linux-amd64"

# Help target
help:
	@echo "Available targets:"
	@echo "  make build              - Build the application"
	@echo "  make run                - Run the application locally"
	@echo "  make test               - Run tests"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make fmt                - Format code"
	@echo "  make lint               - Run linter"
	@echo "  make deps               - Install dependencies"
	@echo "  make dev                - Run in development mode"
	@echo ""
	@echo "Docker targets:"
	@echo "  make docker-build       - Build Docker image"
	@echo "  make docker-up          - Start Docker containers"
	@echo "  make docker-down        - Stop Docker containers"
	@echo "  make docker-logs        - View application logs"
	@echo "  make docker-restart     - Restart Docker containers"
	@echo "  make docker-clean       - Remove containers and volumes"
	@echo ""
	@echo "CI/CD targets:"
	@echo "  make ci                 - Run complete CI pipeline"
	@echo "  make ci-test            - Run tests with coverage"
	@echo "  make ci-lint            - Run linters"
	@echo "  make ci-build           - Build for CI"
	@echo "  make coverage           - Generate coverage report"
	@echo "  make check              - Run pre-commit checks"
	@echo "  make deploy-docker      - Deploy with Docker Compose"
	@echo "  make deploy-production  - Build production binary"
	@echo ""
	@echo "  make help               - Show this help message"
