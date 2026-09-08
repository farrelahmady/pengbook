# Pengbook - Root Makefile
# All commands run from the project root.

# Migration
DB_DRIVER ?= postgres
DB_STRING ?= "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)"
MIGRATE_DIR ?= ./api/migrations

COMPOSE_DEV = docker compose -f docker-compose.dev.yml

.PHONY: help dev dev-infra dev-api dev-web dev-down \
        api-run api-test \
        migrate-up migrate-down migrate-status migrate-create migrate-reset migrate-fix

# ──────────────────────────────────────────────
#  Development
# ──────────────────────────────────────────────

## help: Show this help message
help:
	@echo "Usage: make [command]"
	@echo ""
	@echo "Development:"
	@grep -E '^## dev' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "API (local):"
	@grep -E '^## api-' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Migrations:"
	@grep -E '^## migrate-' $(MAKEFILE_LIST) | sed 's/## /  /'

## dev: Start everything (infra + api + web)
dev: dev-infra
	@echo ""
	@echo "Infrastructure is ready. Starting API and Web..."
	@echo "  API: http://localhost:8080"
	@echo "  Web: http://localhost:3000"
	@echo ""
	@echo "Run in separate terminals:"
	@echo "  make dev-api"
	@echo "  make dev-web"
	@echo ""
	@echo "Or start both in background (Windows):"
	@echo "  start /b cmd /c \"make dev-api\""
	@echo "  start /b cmd /c \"make dev-web\""

## dev-infra: Start postgres + redis only
dev-infra:
	$(COMPOSE_DEV) up -d
	@echo "Waiting for services to be healthy..."
	@timeout /t 3 /nobreak >nul 2>&1 || sleep 3
	@echo "Infrastructure ready."

## dev-api: Run API server locally (hot reload)
dev-api:
	cd api && go run ./cmd/api

## dev-web: Run Web locally (hot reload)
dev-web:
	cd web && pnpm dev

## dev-down: Stop infrastructure
dev-down:
	$(COMPOSE_DEV) down

## dev-reset: Stop infrastructure and remove volumes
dev-reset:
	$(COMPOSE_DEV) down -v

# ──────────────────────────────────────────────
#  API (standalone)
# ──────────────────────────────────────────────

## api-run: Start the API server
api-run:
	cd api && go run ./cmd/api

## api-test: Run API tests
api-test:
	cd api && go test ./...

# ──────────────────────────────────────────────
#  Migrations
# ──────────────────────────────────────────────

## migrate-up: Apply all pending migrations
migrate-up:
	goose -dir $(MIGRATE_DIR) $(DB_DRIVER) $(DB_STRING) up

## migrate-down: Rollback the last migration
migrate-down:
	goose -dir $(MIGRATE_DIR) $(DB_DRIVER) $(DB_STRING) down

## migrate-status: Show migration status
migrate-status:
	goose -dir $(MIGRATE_DIR) $(DB_DRIVER) $(DB_STRING) status

## migrate-create: Create a new migration file (usage: make migrate-create name=add_posts)
migrate-create:
	@if [ -z "$(name)" ]; then echo "Error: name is required. Usage: make migrate-create name=add_posts"; exit 1; fi
	goose -dir $(MIGRATE_DIR) -s create $(name) sql

## migrate-reset: Rollback all migrations
migrate-reset:
	goose -dir $(MIGRATE_DIR) $(DB_DRIVER) $(DB_STRING) reset

## migrate-fix: Apply sequential ordering to migrations
migrate-fix:
	goose -dir $(MIGRATE_DIR) fix
