.PHONY: help build run test clean migrate-up migrate-down docker-build docker-up docker-down test-integration

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build the application
build: ## Build the application
	@echo "Building application..."
	go build -o bin/digital-wallet-go ./cmd/server

# Run the application locally
run: ## Run the application locally
	@echo "Running application..."
	go run ./cmd/server

# Run tests
test: ## Run unit tests
	@echo "Running unit tests..."
	go test -v ./...

# Run integration tests
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -v ./test/integration/...

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Database migrations
migrate-up: ## Run database migrations
	@echo "Running database migrations..."
	@echo "Migrations are automatically handled by GORM on application startup"
	@echo "No manual SQL execution needed!"

migrate-down: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	@echo "Note: This will drop all tables. Use with caution!"
	@read -p "Are you sure? (y/N): " confirm && [ "$$confirm" = "y" ] || exit 1
	docker-compose exec postgres psql -U postgres -d digital_wallet -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t digital-wallet-go .

docker-up: ## Start services with Docker Compose
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-down: ## Stop services with Docker Compose
	@echo "Stopping services with Docker Compose..."
	docker-compose down

docker-logs: ## View Docker Compose logs
	@echo "Viewing Docker Compose logs..."
	docker-compose logs -f

# Development setup
dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env file from .env.example"; \
	fi
	docker-compose up -d postgres redis
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Development environment ready!"

# Code quality
lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

# Dependencies
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Generate mocks (if using mockgen)
mocks: ## Generate mocks
	@echo "Generating mocks..."
	@if command -v mockgen >/dev/null 2>&1; then \
		mockgen -source=internal/repository/interfaces.go -destination=internal/repository/mocks/mock_repositories.go; \
	else \
		echo "mockgen not installed. Install with: go install github.com/golang/mock/mockgen@latest"; \
	fi
