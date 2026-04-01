# Configuration
BINARY_NAME := casino
BUILD_DIR   := tmp

# Go tools
GO      := go
GOTEST  := $(GO) test -v
GOFLAGS := -ldflags="-w -s"

# Docker
DC      := docker compose
DC_DEV  := $(DC) -f compose.yaml -f compose.dev.yaml
NETWORK := casino_net

# Quality
GOLANGCI_LINT := $(shell command -v golangci-lint 2>/dev/null || echo "echo 'golangci-lint not installed, skipping...'")

SERVICES := api consumer postgres kafka kafka_init seeder

# Docker image names
DOCKER_IMAGE := casino

# Seeder configuration (defaults)
CASINO_KAFKA_HOST      ?= localhost
CASINO_KAFKA_PORT      ?= 9092
CASINO_KAFKA_TOPIC     ?= transactions
CASINO_SEEDER_COUNT     ?= 10
CASINO_SEEDER_USER_IDS  ?= user1,user2,user3
CASINO_SEEDER_AMOUNT    ?= 500

# Load environment variables from .env file
-include .env
export

# Prevent Makefile variables from being exported to subprocesses
unexport GOFLAGS GOTEST GOLANGCI_LINT SERVICES BINARY_NAME BUILD_DIR

# Phony targets
.PHONY: help test test-coverage lint \
        build up down logs dev \
        all generate seed migrate migrate-down

# Service specific targets
.PHONY: $(SERVICES) \
        $(foreach svc,$(SERVICES),$(svc)-dev) \
        $(foreach svc,$(SERVICES),$(svc)-logs)

# Default goal
.DEFAULT_GOAL := help

all: help

help: ## Show this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -Eh '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-18s %s\n", $$1, $$2}'
	@echo ""
	@echo "Individual Service Management:"
	@echo "  make [service]         Start a service (production)"
	@echo "  make [service]-dev     Start a service (development)"
	@echo "  make [service]-logs    Show logs for a service"
	@echo ""
	@echo "Services: $(SERVICES)"

# Development
dev: ## Start development with hot reload
	$(DC_DEV) up -d

# Testing & Quality
test: ## Run unit tests
	$(GOTEST) ./internal/...

test-integration: ## Run integration tests
	$(GOTEST) -tags=integration ./tests/integration/...

test-coverage: ## Run tests and generate coverage report
	@mkdir -p $(BUILD_DIR)
	@$(GOTEST) -coverprofile=$(BUILD_DIR)/coverage.out ./internal/...
	@grep -v -E "internal/transactions/adapters/postgres/db|internal/transactions/ports/http/gen|_test.go|mock" \
		$(BUILD_DIR)/coverage.out > $(BUILD_DIR)/coverage_filtered.out
	@$(GO) tool cover -func=$(BUILD_DIR)/coverage_filtered.out
	@$(GO) tool cover -html=$(BUILD_DIR)/coverage_filtered.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report generated at $(BUILD_DIR)/coverage.html"

lint: ## Run golangci-lint
	@$(GOLANGCI_LINT) run

# Individual Service Management
define service_rules
$(1): ## Start $(1) (production)
	$(DC) up -d --force-recreate casino_$(1)

$(1)-dev: ## Start $(1) (development)
	$(DC_DEV) up -d --force-recreate casino_$(1)

$(1)-logs: ## Show logs for $(1)
	$(DC) logs -f casino_$(1)
endef

$(foreach svc,$(SERVICES),$(eval $(call service_rules,$(svc))))

up: ## Start all services
	$(DC) up -d

logs: ## Show all logs
	$(DC) logs -f

down: ## Stop all services
	$(DC) down

generate: ## Generate Go code from OpenAPI, SQLC, and Mockery
	$(GO) generate ./...
	@mkdir -p $(BUILD_DIR)
	@# https://github.com/sqlc-dev/sqlc/issues/4065
	@sed '/^\\restrict /d;/^\\unrestrict /d' \
		internal/transactions/adapters/postgres/schema.sql > $(BUILD_DIR)/schema.sql.tmp \
		&& mv $(BUILD_DIR)/schema.sql.tmp internal/transactions/adapters/postgres/schema.sql
	sqlc generate
	mockery

# Migrations
migrate: ## Apply database migrations
	$(DC) run --rm casino_postgres_migrations_up

migrate-down: ## Apply database rollback migration (requires --profile=rollback)
	$(DC) --profile=rollback up -d casino_postgres_migrations_down

# Docker builds
build: ## Build the main application Docker image
	docker build -t $(DOCKER_IMAGE):latest -f Dockerfile .

seed: build ## Run seeder using built image
	docker run --rm \
		--network $(NETWORK) \
		$(DOCKER_IMAGE):latest \
		seeder \
		--kafka-url=$(CASINO_KAFKA_HOST):$(CASINO_KAFKA_PORT) \
		--kafka-topic=$(CASINO_KAFKA_TOPIC) \
		--count=$(CASINO_SEEDER_COUNT) \
		--user-ids=$(CASINO_SEEDER_USER_IDS) \
		--amount=$(CASINO_SEEDER_AMOUNT)
