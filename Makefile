# Service Paths
API_DIR := ./api_service
TRANS_DIR := ./transaction_service
CONV_DIR := ./conversion_service

.PHONY: all build test test-e2e wire swagger clean up down help

all: build test

build: ## Build all microservices
	$(MAKE) -C $(API_DIR) build
	$(MAKE) -C $(TRANS_DIR) build
	$(MAKE) -C $(CONV_DIR) build

test: ## Run unit tests for all microservices
	$(MAKE) -C $(API_DIR) test
	$(MAKE) -C $(TRANS_DIR) test
	$(MAKE) -C $(CONV_DIR) test

test-e2e: ## Run end-to-end tests
	go test -v ./e2e_tests/...

wire: ## Generate Wire DI code for all microservices
	$(MAKE) -C $(API_DIR) wire
	$(MAKE) -C $(TRANS_DIR) wire
	$(MAKE) -C $(CONV_DIR) wire

swagger: ## Generate Swagger docs for the API Gateway
	$(MAKE) -C $(API_DIR) swagger

clean: ## Remove all build artifacts
	$(MAKE) -C $(API_DIR) clean
	$(MAKE) -C $(TRANS_DIR) clean
	$(MAKE) -C $(CONV_DIR) clean
	rm -rf bin/

up: ## Start the full system using Docker Compose
	docker compose up --build

down: ## Stop the full system
	docker compose down

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
