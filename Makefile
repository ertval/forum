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
BINARY_NAME=bin/forum
BINARY_UNIX=bin/forum_unix

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
.PHONY: all
all: clean mod fmt vet test build

# Build the binary
.PHONY: build
build:
	@echo "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)Build complete: $(BINARY_NAME)$(NC)"

# Run the application
.PHONY: run
run: build
	@echo "$(BLUE)Running $(BINARY_NAME)...$(NC)"
	./$(BINARY_NAME)

# Run the application with `go run` (no build step)
.PHONY: go
go:
	@echo "$(BLUE)Running with go run ($(MAIN_PACKAGE))...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GOCMD) run $(MAIN_PACKAGE)

# Test scripts directory
TEST_SCRIPTS_DIR=./scripts/tests

# Run all tests (quiet mode - shows only summary)
.PHONY: test
test:
	@echo "$(BLUE)=========================================$(NC)"
	@echo "$(BLUE)Running Complete Test Suite (Quiet Mode)$(NC)"
	@echo "$(BLUE)=========================================$(NC)"
	@echo ""
	@echo "$(BLUE)Step 1/3: Running Standard Go Tests...$(NC)"
	@$(GOTEST) ./... > /dev/null 2>&1 && echo "$(GREEN)✓ Go standard tests passed$(NC)" || (echo "$(RED)✗ Go tests failed$(NC)" && $(GOTEST) ./... && exit 1)
	@echo ""
	@echo "$(BLUE)Step 2/3: Running Integration Go Tests in tests directory...$(NC)"
	@$(GOTEST) ./tests/... > /dev/null 2>&1 && echo "$(GREEN)✓ All Integration Tests directory passed$(NC)" || (echo "$(RED)✗ Tests directory failed$(NC)" && $(GOTEST) ./tests/... && exit 1)
	@echo ""
	@echo "$(BLUE)Step 3/3: Running E2E Audit Test scripts...$(NC)"
	@$(TEST_SCRIPTS_DIR)/run_all_tests.sh --quiet
	@echo ""
	@echo "$(GREEN)=========================================$(NC)"
	@echo "$(GREEN)All tests complete!$(NC)"
	@echo "$(GREEN)=========================================$(NC)"

# Run all tests (verbose mode - shows all test output)
.PHONY: tests
tests:
	@echo "$(BLUE)=========================================$(NC)"
	@echo "$(BLUE)Running Complete Test Suite (Verbose Mode)$(NC)"
	@echo "$(BLUE)=========================================$(NC)"
	@echo ""
	@echo "$(BLUE)Step 1/3: Running Standard Go Tests...$(NC)"
	$(GOTEST) ./...
	@echo "$(GREEN)Go tests complete$(NC)"
	@echo ""
	@echo "$(BLUE)Step 2/3: Running Integration Go Tests in tests directory...$(NC)"
	$(GOTEST) ./tests/...
	@echo "$(GREEN)Tests directory complete$(NC)"
	@echo ""
	@echo "$(BLUE)Step 3/3: Running E2E Audit Test scripts...$(NC)"
	@$(TEST_SCRIPTS_DIR)/run_all_tests.sh
	@echo "$(GREEN)E2E Audit Integration test scripts complete$(NC)"
	@echo ""
	@echo "$(GREEN)=========================================$(NC)"
	@echo "$(GREEN)All tests passed!$(NC)"
	@echo "$(GREEN)=========================================$(NC)"

.PHONY: test-go
# Run only standard Go tests
test-go:
	@echo "$(BLUE)Running all standard Go tests...$(NC)"
	$(GOTEST) ./...

.PHONY: test-script
# Run all test scripts (e2e)
test-script:
	@echo "$(BLUE)Running all test scripts (scripts/tests/run_all_tests.sh)...$(NC)"
	@$(TEST_SCRIPTS_DIR)/run_all_tests.sh

.PHONY: test-script-api
# Run API test script
test-script-api:
	@echo "$(BLUE)Running API test script (scripts/tests/test_api.sh)...$(NC)"
	@$(TEST_SCRIPTS_DIR)/test_api.sh

.PHONY: test-script-html
# Run HTML test script
test-script-html:
	@echo "$(BLUE)Running HTML test script (scripts/tests/test_pages.sh)...$(NC)"
	@$(TEST_SCRIPTS_DIR)/test_pages.sh

.PHONY: test-coverage
# Run tests with coverage
test-coverage:
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

.PHONY: clean
# Clean build artifacts
clean:
	@echo "$(BLUE)Cleaning...$(NC)"
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean complete$(NC)"

.PHONY: fmt
# Format code
fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GOFMT) ./...
	@echo "$(GREEN)Code formatted$(NC)"

.PHONY: vet
# Vet code
vet:
	@echo "$(BLUE)Vetting code...$(NC)"
	$(GOVET) ./...
	@echo "$(GREEN)Code vetted$(NC)"

.PHONY: tidy
# Tidy modules
tidy:
	@echo "$(BLUE)Tidying modules...$(NC)"
	$(GOMOD) tidy
	@echo "$(GREEN)Modules tidied$(NC)"

.PHONY: mod
# Download dependencies
mod:
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) verify
	@echo "$(GREEN)Dependencies downloaded$(NC)"

.PHONY: migrate
# Run database migrations
migrate:
	@echo "$(BLUE)Running database migrations...$(NC)"
	$(GOCMD) run ./scripts/run_migrations.go
	@echo "$(GREEN)Migrations complete$(NC)"

.PHONY: migration
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

.PHONY: docker-build
# Docker build
docker-build:
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build -t forum .
	@echo "$(GREEN)Docker image built$(NC)"

.PHONY: docker-run
# Docker run
docker-run:
	@echo "$(BLUE)Running Docker container...$(NC)"
	docker run -p 8080:8080 forum

.PHONY: docker-up
# Docker compose up
docker-up:
	@echo "$(BLUE)Starting Docker Compose...$(NC)"
	docker-compose up -d

.PHONY: docker-down
# Docker compose down
docker-down:
	@echo "$(BLUE)Stopping Docker Compose...$(NC)"
	docker-compose down

# Aliases for convenience
# up: Alias to start docker-compose
.PHONY: up
up: docker-up

# down: Alias to stop docker-compose
.PHONY: down
down: docker-down

.PHONY: build-linux
# Cross compilation for Linux
build-linux:
	@echo "$(BLUE)Building for Linux...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) $(MAIN_PACKAGE)
	@echo "$(GREEN)Linux build complete: $(BINARY_UNIX)$(NC)"

.PHONY: dev-setup
# Install development dependencies
dev-setup:
	@echo "$(BLUE)Setting up development environment...$(NC)"
	$(GOMOD) download
	$(GOCMD) install github.com/cosmtrek/air@latest
	@echo "$(GREEN)Development setup complete$(NC)"

.PHONY: dev
# Run with hot reload (requires air)
dev:
	@echo "$(BLUE)Starting development server with hot reload...$(NC)"
	air

.PHONY: check-schema
# Check database schema
check-schema:
	@echo "$(BLUE)Checking database schema...$(NC)"
	$(GOCMD) run ./scripts/check/check_schema.go

.PHONY: seed
# Seed database
seed:
	@echo "$(BLUE)Seeding database...$(NC)"
	bash ./scripts/seed/seed.sh

.PHONY: help
# Help
help:
	@echo "$(BLUE)Available targets:$(NC)"
	@echo "  $(GREEN)all$(NC)             - Run clean, mod, fmt, vet, test, build"
	@echo "  $(GREEN)build$(NC)           - Build the binary"
	@echo "  $(GREEN)run$(NC)             - Build and run the application"
	@echo "  $(GREEN)run-go$(NC)          - Run the application with 'go run' (no build)"
	@echo "  $(GREEN)test$(NC)            - Run all tests (quiet mode - summary only)"
	@echo "  $(GREEN)testv$(NC)           - Run all tests (verbose mode - full output)"
	@echo "  $(GREEN)test-go$(NC)         - Run only standard Go tests"
	@echo "  $(GREEN)test-script$(NC)     - Run all e2e test scripts"
	@echo "  $(GREEN)test-script-api$(NC) - Run API test script"
	@echo "  $(GREEN)test-script-html$(NC) - Run HTML test script"
	@echo "  $(GREEN)test-coverage$(NC)   - Run all tests with coverage report"
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
