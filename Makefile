.PHONY: build run clean test

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
