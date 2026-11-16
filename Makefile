.PHONY: help setup dev build test test-integration test-e2e lint security migrate migrate-status health clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

setup: ## Install all dependencies
	cd apps/web && npm install
	cd apps/api && go mod download

dev: ## Start development servers
	@echo "Starting Mjengofiti development environment..."
	@echo "Backend: http://localhost:8080"
	@echo "Frontend: http://localhost:5173"
	@start /b cmd /c "cd apps\api && go run ./cmd/api"
	cd apps/web && npm run dev

dev-api: ## Start only the Go backend
	cd apps/api && go run ./cmd/api

dev-web: ## Start only the frontend
	cd apps/web && npm run dev

build: build-api build-web ## Build all

build-api: ## Build Go backend
	cd apps/api && go build -o bin/api.exe ./cmd/api

build-web: ## Build frontend
	cd apps/web && npm run build

test: test-api test-web ## Run all unit tests

test-api: ## Run Go backend tests
	cd apps/api && go test ./... -v -count=1

test-web: ## Run frontend tests
	cd apps/web && npm run test -- --run

test-integration: ## Run integration tests (requires database)
	cd apps/api && go test ./... -v -tags=integration -count=1

test-e2e: ## Run end-to-end tests
	cd apps/web && npx playwright test

lint: lint-api lint-web ## Run all linters

lint-api: ## Lint Go code
	cd apps/api && go vet ./...

lint-web: ## Lint frontend code
	cd apps/web && npm run lint

security: ## Run security scans
	cd apps/api && go list -json -m all | go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	cd apps/web && npm audit --audit-level=high

migrate: ## Run database migrations
	cd apps/api && go run ./cmd/migrate up

migrate-status: ## Show migration status
	cd apps/api && go run ./cmd/migrate status

migrate-down: ## Rollback last migration
	cd apps/api && go run ./cmd/migrate down

health: ## Check application health
	@curl -s http://localhost:8080/health/live || echo "Backend not running"
	@curl -s http://localhost:8080/health/ready || echo "Backend not ready"

clean: ## Clean build artifacts
	cd apps/api && del /q /s bin 2>nul || true
	cd apps/web && rd /s /q dist 2>nul || true
