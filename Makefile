.PHONY: build run clean test docker-build docker-up docker-down docker-logs

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
