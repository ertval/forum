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

# Docker compose command detection (prefer modern plugin syntax)
DOCKER_COMPOSE_CMD=$(shell if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then echo "docker compose"; elif command -v docker-compose >/dev/null 2>&1; then echo "docker-compose"; else echo ""; fi)

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

# Default target
all: clean mod fmt vet test build
.PHONY: all

# Build the binary
build:
	@echo "$(BLUE)Building $(BINARY_NAME)...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "$(GREEN)Build complete: $(BINARY_NAME)$(NC)"
.PHONY: build

# Run the application
run: build
	@echo "$(BLUE)Running $(BINARY_NAME)...$(NC)"
	./$(BINARY_NAME)
.PHONY: run

# Run the application with `go run` (no build step)
go:
	@echo "$(BLUE)Running with go run ($(MAIN_PACKAGE))...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GOCMD) run $(MAIN_PACKAGE)
.PHONY: go

# Test scripts directory
TEST_SCRIPTS_DIR=./scripts/tests

# Run all tests (quiet mode - shows only summary)
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
	@bash $(TEST_SCRIPTS_DIR)/run_all_tests.sh --quiet
	@echo ""
	@echo "$(GREEN)=========================================$(NC)"
	@echo "$(GREEN)All tests complete!$(NC)"
	@echo "$(GREEN)=========================================$(NC)"
.PHONY: test

# Run all tests (verbose mode - shows all test output)
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
	@bash $(TEST_SCRIPTS_DIR)/run_all_tests.sh
	@echo "$(GREEN)E2E Audit Integration test scripts complete$(NC)"
	@echo ""
	@echo "$(GREEN)=========================================$(NC)"
	@echo "$(GREEN)All tests passed!$(NC)"
	@echo "$(GREEN)=========================================$(NC)"
.PHONY: tests

# Run only standard Go tests
test-go:
	@echo "$(BLUE)Running all standard Go tests...$(NC)"
	$(GOTEST) ./...
.PHONY: test-go

# Run all test scripts (e2e)
test-script:
	@echo "$(BLUE)Running all test scripts (scripts/tests/run_all_tests.sh)...$(NC)"
	@bash $(TEST_SCRIPTS_DIR)/run_all_tests.sh
.PHONY: test-script

# Run API test script
test-script-api:
	@echo "$(BLUE)Running API test script (scripts/tests/test_api.sh)...$(NC)"
	@bash $(TEST_SCRIPTS_DIR)/test_api.sh
.PHONY: test-script-api

# Run HTML test script
test-script-html:
	@echo "$(BLUE)Running HTML test script (scripts/tests/test_pages.sh)...$(NC)"
	@bash $(TEST_SCRIPTS_DIR)/test_pages.sh
.PHONY: test-script-html

# Run tests with coverage
test-coverage:
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"
.PHONY: test-coverage

# Run tests and show only failures
test-fail:
	@echo "$(BLUE)Running tests and showing only failures...$(NC)"
	@$(GOTEST) ./... | grep "FAIL" || echo "$(GREEN)✓ Go tests: No failures$(NC)"
	@$(GOTEST) ./tests/... | grep "FAIL" || echo "$(GREEN)✓ Integration tests: No failures$(NC)"
	@bash $(TEST_SCRIPTS_DIR)/run_all_tests.sh --quiet | grep "✗" || echo "$(GREEN)✓ Script tests: No failures$(NC)"
.PHONY: test-fail

# Clean build artifacts
clean:
	@echo "$(BLUE)Cleaning...$(NC)"
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f coverage.out coverage.html
	@echo "$(GREEN)Clean complete$(NC)"
.PHONY: clean

# Format code
fmt:
	@echo "$(BLUE)Formatting code...$(NC)"
	$(GOFMT) ./...
	@echo "$(GREEN)Code formatted$(NC)"
.PHONY: fmt

# Vet code
vet:
	@echo "$(BLUE)Vetting code...$(NC)"
	$(GOVET) ./...
	@echo "$(GREEN)Code vetted$(NC)"
.PHONY: vet

# Tidy modules
tidy:
	@echo "$(BLUE)Tidying modules...$(NC)"
	$(GOMOD) tidy
	@echo "$(GREEN)Modules tidied$(NC)"
.PHONY: tidy

# Download dependencies
mod:
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) verify
	@echo "$(GREEN)Dependencies downloaded$(NC)"
.PHONY: mod

# Run database migrations
migrate:
	@echo "$(BLUE)Running database migrations...$(NC)"
	@bash ./scripts/seed/run_migrations.sh
	@echo "$(GREEN)Migrations complete$(NC)"
.PHONY: migrate

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
.PHONY: migration

# Docker build
docker-build: certs
	@echo "$(BLUE)Building Docker image...$(NC)"
	docker build --rm --force-rm -t forum .
	@echo "$(GREEN)Docker image built$(NC)"
.PHONY: docker-build

# Docker run (with named container and volumes for persistence)
docker-run:
	@echo "$(BLUE)Running Docker container...$(NC)"
	@if docker ps -a --format '{{.Names}}' | grep -q '^forum$$'; then \
		echo "$(YELLOW)Container 'forum' already exists. Starting it...$(NC)"; \
		docker start -a forum; \
	else \
		docker run --name forum \
			-p 8080:8080 \
			-v forum-data:/app/data \
			-v forum-uploads:/app/static/uploads \
			forum; \
	fi
.PHONY: docker-run

# Docker compose up
docker-up:
	@echo "$(BLUE)Starting Docker Compose...$(NC)"
	@if [ -z "$(DOCKER_COMPOSE_CMD)" ]; then \
		echo "$(RED)Docker Compose is not available in this environment.$(NC)"; \
		echo "$(YELLOW)Install Docker Desktop/Engine with Compose plugin, then retry.$(NC)"; \
		exit 127; \
	fi
	$(DOCKER_COMPOSE_CMD) up -d
.PHONY: docker-up

# Docker compose down
docker-down:
	@echo "$(BLUE)Stopping Docker Compose...$(NC)"
	@if [ -z "$(DOCKER_COMPOSE_CMD)" ]; then \
		echo "$(RED)Docker Compose is not available in this environment.$(NC)"; \
		exit 127; \
	fi
	$(DOCKER_COMPOSE_CMD) down
.PHONY: docker-down

# up: Alias to start docker-compose
up: docker-up
.PHONY: up

# down: Alias to stop docker-compose
down: docker-down
.PHONY: down

# Cross compilation for Linux
build-linux:
	@echo "$(BLUE)Building for Linux...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_UNIX) $(MAIN_PACKAGE)
	@echo "$(GREEN)Linux build complete: $(BINARY_UNIX)$(NC)"
.PHONY: build-linux

# Check database schema
check-schema:
	@echo "$(BLUE)Checking database schema...$(NC)"
	$(GOCMD) run ./scripts/check/check_schema.go
.PHONY: check-schema

# Seed database
seed:
	@echo "$(BLUE)Seeding database...$(NC)"
	bash ./scripts/seed/seed.sh
.PHONY: seed

# Generate TLS certificates (self-signed, for development/Docker use)
certs:
	@echo "$(BLUE)Generating TLS certificates...$(NC)"
	bash ./scripts/seed/generate_certs.sh
	@echo "$(GREEN)Certificates ready in ./certs/$(NC)"
.PHONY: certs

# Generate all audit evidence artifacts
audit-evidence:
	@echo "$(BLUE)Generating audit evidence artifacts...$(NC)"
	bash ./scripts/audit/run_evidence_all.sh
	@echo "$(GREEN)Audit evidence complete$(NC)"
.PHONY: audit-evidence

# Docker build/run/health evidence
audit-evidence-docker:
	@echo "$(BLUE)Generating Docker health evidence...$(NC)"
	bash ./scripts/audit/evidence_docker_health.sh
.PHONY: audit-evidence-docker

# Page sweep evidence
audit-evidence-pages:
	@echo "$(BLUE)Generating page sweep evidence...$(NC)"
	bash ./scripts/audit/evidence_page_sweep.sh
.PHONY: audit-evidence-pages

# Performance smoke evidence
audit-evidence-perf:
	@echo "$(BLUE)Generating performance evidence...$(NC)"
	bash ./scripts/audit/evidence_performance.sh
.PHONY: audit-evidence-perf

# Dead reference audit/cleanup (DB + template/static refs)
dead-refs-dry-run:
	@echo "$(BLUE)Running dead reference audit (dry-run)...$(NC)"
	bash ./scripts/audit/clean_dead_refs.sh --dry-run
.PHONY: dead-refs-dry-run

dead-refs-apply:
	@echo "$(YELLOW)Applying dead reference cleanup...$(NC)"
	bash ./scripts/audit/clean_dead_refs.sh --apply
.PHONY: dead-refs-apply

# Docker cleanup process (safe dry run)
docker-prune-dry-run:
	@echo "$(BLUE)Running Docker prune dry-run process...$(NC)"
	bash ./scripts/audit/docker_cleanup_prune.sh --dry-run
.PHONY: docker-prune-dry-run

# Docker cleanup process (destructive)
docker-prune-apply:
	@echo "$(YELLOW)Applying Docker prune process...$(NC)"
	bash ./scripts/audit/docker_cleanup_prune.sh --apply
.PHONY: docker-prune-apply

# ── Port-kill helpers ────────────────────────────────────────────────────────
# Kill whatever process is listening on port 8080 (HTTP)
kill-http:
	@echo "$(YELLOW)Killing process on port 8080...$(NC)"
	@if command -v fuser >/dev/null 2>&1; then \
		fuser -k 8080/tcp 2>/dev/null && echo "$(GREEN)Port 8080 cleared$(NC)" || echo "$(BLUE)Nothing on port 8080$(NC)"; \
	elif command -v lsof >/dev/null 2>&1; then \
		pid=$$(lsof -ti tcp:8080 2>/dev/null); \
		if [ -n "$$pid" ]; then kill -9 $$pid && echo "$(GREEN)Port 8080 cleared (PID $$pid)$(NC)"; else echo "$(BLUE)Nothing on port 8080$(NC)"; fi; \
	else \
		echo "$(YELLOW)fuser/lsof not found — on Windows use: netstat -ano | findstr :8080$(NC)"; \
	fi
.PHONY: kill-http

# Kill whatever process is listening on port 8443 (HTTPS)
kill-https:
	@echo "$(YELLOW)Killing process on port 8443...$(NC)"
	@if command -v fuser >/dev/null 2>&1; then \
		fuser -k 8443/tcp 2>/dev/null && echo "$(GREEN)Port 8443 cleared$(NC)" || echo "$(BLUE)Nothing on port 8443$(NC)"; \
	elif command -v lsof >/dev/null 2>&1; then \
		pid=$$(lsof -ti tcp:8443 2>/dev/null); \
		if [ -n "$$pid" ]; then kill -9 $$pid && echo "$(GREEN)Port 8443 cleared (PID $$pid)$(NC)"; else echo "$(BLUE)Nothing on port 8443$(NC)"; fi; \
	else \
		echo "$(YELLOW)fuser/lsof not found — on Windows use: netstat -ano | findstr :8443$(NC)"; \
	fi
.PHONY: kill-https

# Kill processes on both default ports (8080 + 8443)
kill: kill-http kill-https
	@echo "$(GREEN)All forum ports cleared$(NC)"
.PHONY: kill

# Help
help:
	@echo "$(BLUE)Available targets:$(NC)"
	@echo "  $(GREEN)all$(NC)             - Run clean, mod, fmt, vet, test, build"
	@echo "  $(GREEN)build$(NC)           - Build the binary"
	@echo "  $(GREEN)run$(NC)             - Build and run the application"
	@echo "  $(GREEN)go$(NC)              - Run the application with 'go run' (no build)"
	@echo "  $(GREEN)test$(NC)            - Run all tests (quiet mode - summary only)"
	@echo "  $(GREEN)tests$(NC)           - Run all tests (verbose mode - full output)"
	@echo "  $(GREEN)test-go$(NC)         - Run only standard Go tests"
	@echo "  $(GREEN)test-script$(NC)     - Run all e2e test scripts"
	@echo "  $(GREEN)test-script-api$(NC) - Run API test script"
	@echo "  $(GREEN)test-script-html$(NC) - Run HTML test script"
	@echo "  $(GREEN)test-fail$(NC)       - Run all tests and show only failures"
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
	@echo "  $(GREEN)check-schema$(NC)    - Check database schema"
	@echo "  $(GREEN)audit-evidence$(NC)  - Run all audit evidence scripts"
	@echo "  $(GREEN)audit-evidence-docker$(NC) - Generate docker build/run/health artifacts"
	@echo "  $(GREEN)audit-evidence-pages$(NC) - Generate page sweep artifacts"
	@echo "  $(GREEN)audit-evidence-perf$(NC) - Generate performance artifacts"
	@echo "  $(GREEN)dead-refs-dry-run$(NC) - Audit dead refs in templates/static/DB without changes"
	@echo "  $(GREEN)dead-refs-apply$(NC) - Apply safe cleanup of dead DB references"
	@echo "  $(GREEN)docker-prune-dry-run$(NC) - Show docker cleanup process without deleting"
	@echo "  $(GREEN)docker-prune-apply$(NC) - Apply docker cleanup process"
	@echo "  $(GREEN)kill$(NC)            - Kill processes on ports 8080 and 8443"
	@echo "  $(GREEN)kill-http$(NC)       - Kill process on port 8080 (HTTP)"
	@echo "  $(GREEN)kill-https$(NC)      - Kill process on port 8443 (HTTPS)"
	@echo "  $(GREEN)seed$(NC)            - Seed database"
	@echo "  $(GREEN)certs$(NC)           - Generate self-signed TLS certificates"
	@echo "  $(GREEN)help$(NC)            - Show this help"
.PHONY: help
