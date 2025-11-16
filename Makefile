.PHONY: build run go test clean fmt vet mod tidy migrate migration docker-build docker-run docker-up docker-down up down help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Main package
MAIN_PACKAGE=./cmd/forum
BINARY_NAME=forum
BINARY_UNIX=$(BINARY_NAME)_unix

# Build flags
CGO_ENABLED=1
LDFLAGS=-ldflags "-w -s"

# Database
MIGRATIONS_DIR=./migrations

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

# Default target
all: clean mod fmt vet test build

# Build the binary
build:
	@echo "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)Build complete: $(BINARY_NAME)$(NC)"

# Run the application
run: build
	@echo "$(BLUE)Running $(BINARY_NAME)...$(NC)"
	./$(BINARY_NAME)

# Run the application with `go run` (no build step)
go:
	@echo "$(BLUE)Running with go run ($(MAIN_PACKAGE))...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GOCMD) run $(MAIN_PACKAGE)

# Test scripts directory
TEST_SCRIPTS_DIR=./scripts/tests

# Run full test suite (uses the centralized script)
test:
	@echo "$(BLUE)Running full test suite (scripts/tests/run_all_tests.sh)...$(NC)"
	@$(TEST_SCRIPTS_DIR)/run_all_tests.sh

# Run only API tests
test-api:
	@echo "$(BLUE)Running API tests (scripts/tests/test_api.sh)...$(NC)"
	@$(TEST_SCRIPTS_DIR)/test_api.sh

# Run only Page/HTML tests
test-html:
	@echo "$(BLUE)Running Page/HTML tests (scripts/tests/test_pages.sh)...$(NC)"
	@$(TEST_SCRIPTS_DIR)/test_pages.sh

# Run tests with coverage
test-coverage:
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

# Clean build artifacts
clean:
	@echo "$(BLUE)Cleaning...$(NC)"
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean complete$(NC)"

# Format code
fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GOFMT) ./...
	@echo "$(GREEN)Code formatted$(NC)"

# Vet code
vet:
	@echo "$(BLUE)Vetting code...$(NC)"
	$(GOVET) ./...
	@echo "$(GREEN)Code vetted$(NC)"

# Tidy modules
tidy:
	@echo "$(BLUE)Tidying modules...$(NC)"
	$(GOMOD) tidy
	@echo "$(GREEN)Modules tidied$(NC)"

# Download dependencies
mod:
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) verify
	@echo "$(GREEN)Dependencies downloaded$(NC)"

# Run database migrations
migrate:
	@echo "$(BLUE)Running database migrations...$(NC)"
	$(GOCMD) run ./scripts/run_migrations.go
	@echo "$(GREEN)Migrations complete$(NC)"

# Create a new migration file
migration:
	@echo "$(YELLOW)Usage: make migration NAME=migration_name$(NC)"
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)Error: NAME is required$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Creating migration: $(NAME)$(NC)"
	@timestamp=$$(date +%s); \
	filename=$(MIGRATIONS_DIR)/$${timestamp}_$(NAME).sql; \
	cp $(MIGRATIONS_DIR)/template/000_template_migration.sql $${filename}; \
	echo "$(GREEN)Migration created: $${filename}$(NC)"

# Docker build
docker-build:
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build -t forum .
	@echo "$(GREEN)Docker image built$(NC)"

# Docker run
docker-run:
	@echo "$(BLUE)Running Docker container...$(NC)"
	docker run -p 8080:8080 forum

# Docker compose up
docker-up:
	@echo "$(BLUE)Starting Docker Compose...$(NC)"
	docker-compose up -d

# Docker compose down
docker-down:
	@echo "$(BLUE)Stopping Docker Compose...$(NC)"
	docker-compose down

# Aliases for convenience
up: docker-up

down: docker-down

# Cross compilation for Linux
build-linux:
	@echo "$(BLUE)Building for Linux...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) $(MAIN_PACKAGE)
	@echo "$(GREEN)Linux build complete: $(BINARY_UNIX)$(NC)"

# Install development dependencies
dev-setup:
	@echo "$(BLUE)Setting up development environment...$(NC)"
	$(GOMOD) download
	$(GOCMD) install github.com/cosmtrek/air@latest
	@echo "$(GREEN)Development setup complete$(NC)"

# Run with hot reload (requires air)
dev:
	@echo "$(BLUE)Starting development server with hot reload...$(NC)"
	air

# Check database schema
check-schema:
	@echo "$(BLUE)Checking database schema...$(NC)"
	$(GOCMD) run ./scripts/check/check_schema.go

# Seed database
seed:
	@echo "$(BLUE)Seeding database...$(NC)"
	bash ./scripts/seed/seed.sh

# Help
help:
	@echo "$(BLUE)Available targets:$(NC)"
	@echo "  $(GREEN)all$(NC)             - Run clean, mod, fmt, vet, test, build"
	@echo "  $(GREEN)build$(NC)           - Build the binary"
	@echo "  $(GREEN)run$(NC)             - Build and run the application"
	@echo "  $(GREEN)run-go$(NC)          - Run the application with 'go run' (no build)"
	@echo "  $(GREEN)test$(NC)            - Run tests"
	@echo "  $(GREEN)test-coverage$(NC)   - Run tests with coverage report"
	@echo "  $(GREEN)clean$(NC)           - Clean build artifacts"
	@echo "  $(GREEN)fmt$(NC)             - Format code"
	@echo "  $(GREEN)vet$(NC)             - Vet code"
	@echo "  $(GREEN)tidy$(NC)            - Tidy modules"
	@echo "  $(GREEN)mod$(NC)             - Download dependencies"
	@echo "  $(GREEN)migrate$(NC)         - Run database migrations"
	@echo "  $(GREEN)migration$(NC)       - Create new migration (NAME=required)"
	@echo "  $(GREEN)docker-build$(NC)    - Build Docker image"
	@echo "  $(GREEN)docker-run$(NC)      - Run Docker container"
	@echo "  $(GREEN)docker-up$(NC)       - Start Docker Compose"
	@echo "  $(GREEN)docker-down$(NC)     - Stop Docker Compose"
	@echo "  $(GREEN)up$(NC)              - Alias for docker-up"
	@echo "  $(GREEN)down$(NC)            - Alias for docker-down"
	@echo "  $(GREEN)build-linux$(NC)     - Cross compile for Linux"
	@echo "  $(GREEN)dev-setup$(NC)       - Setup development environment"
	@echo "  $(GREEN)dev$(NC)             - Run with hot reload"
	@echo "  $(GREEN)check-schema$(NC)    - Check database schema"
	@echo "  $(GREEN)seed$(NC)            - Seed database"
	@echo "  $(GREEN)help$(NC)            - Show this help"
