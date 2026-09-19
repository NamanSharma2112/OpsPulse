# OpsPulse developer tasks. Run `make help` for the list.
.DEFAULT_GOAL := help
.PHONY: help up down logs reset api web test test-api build build-api build-web \
        fmt lint tidy migrate-up migrate-down

API_DIR := services/api
DATABASE_URL ?= postgres://opspulse:opspulse@localhost:5432/opspulse?sslmode=disable

help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

up: ## Start the whole stack in the background
	docker compose up -d --build

down: ## Stop the stack
	docker compose down

logs: ## Follow logs from every service
	docker compose logs -f

reset: ## Stop the stack and delete the database volume
	docker compose down -v

api: ## Run the API on the host (needs a reachable Postgres)
	cd $(API_DIR) && DATABASE_URL="$(DATABASE_URL)" go run ./cmd/api

web: ## Run the dashboard in dev mode
	npm run dev

test: test-api ## Run every test suite

test-api: ## Run the Go tests
	cd $(API_DIR) && go test ./...

build: build-api build-web ## Build both applications

build-api: ## Compile the API binary to services/api/bin/api
	cd $(API_DIR) && go build -o bin/api ./cmd/api

build-web: ## Build the dashboard
	npm run build

fmt: ## Format Go sources
	cd $(API_DIR) && gofmt -w ./cmd ./internal

lint: ## Vet Go sources and typecheck the dashboard
	cd $(API_DIR) && go vet ./...
	npm run typecheck

tidy: ## Tidy Go module dependencies
	cd $(API_DIR) && go mod tidy

migrate-up: ## Apply every .up.sql migration in order
	@for f in infra/migrations/*.up.sql; do \
		echo "applying $$f"; \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -q -f "$$f" || exit 1; \
	done

migrate-down: ## Roll back every .down.sql migration in reverse order
	@for f in $$(ls -r infra/migrations/*.down.sql); do \
		echo "reverting $$f"; \
		psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -q -f "$$f" || exit 1; \
	done
