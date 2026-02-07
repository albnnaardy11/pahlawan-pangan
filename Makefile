.PHONY: help build test run docker-build docker-run k8s-deploy clean lint fmt tidy
.DEFAULT_GOAL := help

# --- Professional Automation ---

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

lint: ## Run professional linters (golangci-lint)
	@echo "🔍 Running linters..."
	golangci-lint run ./...

fmt: ## Professional code formatting
	@echo "✨ Formatting code..."
	go fmt ./...
	goimports -w .

test: ## Run unit tests with race detection and coverage
	@echo "🧪 Running unit tests..."
	go test -v -race -coverprofile=coverage.out ./...

test-integration: ## Run professional integration tests
	@echo "🧬 Running integration tests..."
	go test -v -tags=integration ./...

build: ## Build optimized production binary
	@echo "🏗️ Building production binary..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/server cmd/server/main.go

docker-build: ## Multi-stage Docker build
	@echo "🐳 Building Docker image..."
	docker build -t pahlawan-pangan:v1 .

run: ## Run server locally
	@echo "🚀 Starting server..."
	go run cmd/server/main.go

tidy: ## Tidy and verify go modules
	@echo "🧹 Tidying modules..."
	go mod tidy
	go mod verify

all: fmt tidy lint test build ## Run all professional checks and build
