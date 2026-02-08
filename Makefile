
# Variables
APP_NAME := pahlawan-pangan
BUILD_DIR := bin
CMD_DIR := ./cmd/server
DOCKER_IMAGE := pahlawan-pangan

# Tools
GO := go
DOCKER := docker
LINT := golangci-lint

# Default target
.PHONY: all
all: clean lint test build

# Build the application
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags="-w -s" -o $(BUILD_DIR)/server $(CMD_DIR)

# Run the application locally
.PHONY: run
run:
	@echo "Running $(APP_NAME)..."
	$(GO) run $(CMD_DIR)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test -v -race -cover ./...

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	$(LINT) run ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

# Docker build
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	$(DOCKER) build -t $(DOCKER_IMAGE) .

# Docker run
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	$(DOCKER) run -p 8080:8080 -p 9090:9090 $(DOCKER_IMAGE)

# Database migration (placeholder using scripts)
.PHONY: migrate
migrate:
	@echo "Running migrations..."
	@sh ./scripts/migrate.sh

# Seed database (placeholder using scripts)
.PHONY: seed
seed:
	@echo "Seeding database..."
	@sh ./scripts/seed.sh

# Help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application locally"
	@echo "  make test          - Run tests"
	@echo "  make lint          - Run linter"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run Docker container"
	@echo "  make migrate       - Run database migrations"
	@echo "  make seed          - Seed database"
