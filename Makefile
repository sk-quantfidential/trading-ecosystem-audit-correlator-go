# Audit Correlator Go - Makefile

.PHONY: help test test-unit test-integration test-all build clean lint

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Test targets
test: test-unit ## Run unit tests (default)

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	go test -tags=unit ./internal/... -v

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	go test -tags=integration ./internal/... -v

test-all: ## Run all tests (unit + integration)
	@echo "Running all tests..."
	go test -tags="unit integration" ./internal/... -v

test-short: ## Run tests in short mode (skip slow tests)
	@echo "Running tests in short mode..."
	go test -tags=unit ./internal/... -short -v

# Build targets
build: ## Build the audit correlator binary
	@echo "Building audit correlator..."
	go build -o audit-correlator ./cmd/server

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f audit-correlator server
	go clean -testcache

# Development targets
lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# Info targets
test-list: ## List all available tests
	@echo "Unit tests:"
	@go test -tags=unit ./internal/... -list=. 2>/dev/null || echo "  (Run 'make test-unit' to see unit tests)"
	@echo ""
	@echo "Integration tests:"
	@go test -tags=integration ./internal/... -list=. 2>/dev/null || echo "  (Run 'make test-integration' to see integration tests)"

test-files: ## Show test files
	@echo "Test files in audit-correlator-go:"
	@find . -name "*_test.go" -exec echo "  {}" \;

# Status check
status: ## Check current test status
	@echo "=== Audit Correlator Go Test Status ==="
	@echo ""
	@echo "Unit Tests (tags=unit):"
	@go test -tags=unit ./internal/... -v 2>&1 | grep -E "(PASS|FAIL|SKIP|===)" | head -10 || echo "  No unit tests found"
	@echo ""
	@echo "Integration Tests (tags=integration):"
	@go test -tags=integration ./internal/... -v 2>&1 | grep -E "(PASS|FAIL|SKIP|===)" | head -10 || echo "  No integration tests found"