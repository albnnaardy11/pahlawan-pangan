.PHONY: help build test run docker-build docker-run k8s-deploy clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the Go binary
	@echo "Building..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/server cmd/server/main.go

test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	go test -v -tags=integration ./...

run: ## Run locally
	@echo "Starting server..."
	go run cmd/server/main.go

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t pahlawan-pangan:latest .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 -p 9090:9090 \
		-e DATABASE_URL=${DATABASE_URL} \
		-e REDIS_URL=${REDIS_URL} \
		-e NATS_URL=${NATS_URL} \
		pahlawan-pangan:latest

compose-up: ## Start Docker Compose stack
	@echo "Starting infrastructure..."
	docker-compose up -d

compose-down: ## Stop Docker Compose stack
	@echo "Stopping infrastructure..."
	docker-compose down

k8s-deploy: ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	kubectl create namespace pahlawan-pangan --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f k8s/redis-cluster.yaml
	kubectl apply -f k8s/deployment.yaml

k8s-delete: ## Delete Kubernetes resources
	@echo "Deleting Kubernetes resources..."
	kubectl delete -f k8s/deployment.yaml
	kubectl delete -f k8s/redis-cluster.yaml

lint: ## Run linters
	@echo "Running linters..."
	golangci-lint run ./...

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

mod-tidy: ## Tidy Go modules
	@echo "Tidying modules..."
	go mod tidy
	go mod verify

load-test: ## Run load tests (requires k6)
	@echo "Running load tests..."
	k6 run tests/load/surplus_post.js

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

db-migrate: ## Run database migrations
	@echo "Running migrations..."
	psql ${DATABASE_URL} < db/schema.sql

db-seed: ## Seed database with test data
	@echo "Seeding database..."
	go run scripts/seed.go

metrics: ## View Prometheus metrics
	@echo "Opening Prometheus..."
	open http://localhost:9090

traces: ## View Jaeger traces
	@echo "Opening Jaeger..."
	open http://localhost:16686

grafana: ## View Grafana dashboards
	@echo "Opening Grafana..."
	open http://localhost:3000

all: fmt lint test build ## Run all checks and build
