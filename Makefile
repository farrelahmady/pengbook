# Pengbook - Root Makefile
# All commands run from the project root.

# Load .env file if exists (environment variables take precedence)
-include ./api/.env

# Migration
DB_DRIVER ?= postgres
DB_STRING ?= "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)"
MIGRATE_DIR ?= ./api/migrations
SEED_DIR ?= ./api/seeders
GOOSE_MIGRATION_TABLE ?= goose_db_version
GOOSE_SEED_TABLE ?= goose_seed_version

COMPOSE_DEV = docker compose -f docker-compose.dev.yml

.PHONY: help dev dev-infra dev-api dev-web dev-down dev-reset \
        api-run api-test api-build api-lint api-fmt \
        web-build web-lint \
        docker-build docker-up docker-down \
        deploy \
        setup \
        swagger-gen \
        migrate-up migrate-down migrate-status migrate-create migrate-reset migrate-fix \
        db-reset \
        seed-up seed-down seed-reset seed-status seed-create \
        clean

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
	@echo "Web (local):"
	@grep -E '^## web-' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Docker Production:"
	@grep -E '^## docker-' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Deploy:"
	@grep -E '^## deploy' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Setup:"
	@grep -E '^## setup' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Swagger:"
	@grep -E '^## swagger-' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Migrations:"
	@grep -E '^## migrate-' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Seeds:"
	@grep -E '^## seed-' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Database:"
	@grep -E '^## db-' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Clean:"
	@grep -E '^## clean' $(MAKEFILE_LIST) | sed 's/## /  /'

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
	pnpm --filter pengbook-web dev

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
	goose -dir $(MIGRATE_DIR) -table $(GOOSE_MIGRATION_TABLE) $(DB_DRIVER) $(DB_STRING) up

## migrate-down: Rollback the last migration
migrate-down:
	goose -dir $(MIGRATE_DIR) -table $(GOOSE_MIGRATION_TABLE) $(DB_DRIVER) $(DB_STRING) down

## migrate-status: Show migration status
migrate-status:
	goose -dir $(MIGRATE_DIR) -table $(GOOSE_MIGRATION_TABLE) $(DB_DRIVER) $(DB_STRING) status

## migrate-create: Create a new migration file (usage: make migrate-create name=add_posts)
migrate-create:
	@if [ -z "$(name)" ]; then echo "Error: name is required. Usage: make migrate-create name=add_posts"; exit 1; fi
	goose -dir $(MIGRATE_DIR) create $(name) sql

## migrate-reset: Rollback all migrations
migrate-reset:
	goose -dir $(MIGRATE_DIR) -table $(GOOSE_MIGRATION_TABLE) $(DB_DRIVER) $(DB_STRING) reset

## migrate-fix: Apply sequential ordering to migrations
migrate-fix:
	goose -dir $(MIGRATE_DIR) fix

# ──────────────────────────────────────────────
#  Build
# ──────────────────────────────────────────────

## api-build: Build Go API binary
api-build:
	cd api && go build -o bin/api ./cmd/api

## web-build: Build Next.js for production
web-build:
	pnpm --filter pengbook-web build

# ──────────────────────────────────────────────
#  Lint & Format
# ──────────────────────────────────────────────

## api-lint: Run golangci-lint on API code
api-lint:
	cd api && golangci-lint run ./...

## api-fmt: Format Go code
api-fmt:
	cd api && go fmt ./...

## web-lint: Run ESLint on web code
web-lint:
	pnpm --filter pengbook-web lint

# ──────────────────────────────────────────────
#  Docker Production
# ──────────────────────────────────────────────

## docker-build: Build all Docker images
docker-build:
	docker compose build

## docker-up: Start production stack
docker-up:
	docker compose up -d

## docker-down: Stop production stack
docker-down:
	docker compose down

# ──────────────────────────────────────────────
#  Deploy
# ──────────────────────────────────────────────

## deploy: Generate JWT secret (if missing), build, and start production stack
deploy:
	@if [ ! -f .env ]; then echo "Error: .env file not found. Run 'make setup' first."; exit 1; fi
	@if grep -q "^JWT_SECRET=" .env && [ -n "$$(grep '^JWT_SECRET=' .env | cut -d'=' -f2)" ]; then \
		echo "JWT_SECRET already set in .env"; \
	else \
		echo "Generating JWT_SECRET..."; \
		JWT_VAL=$$(openssl rand -hex 32 2>/dev/null || powershell -Command "[System.Convert]::ToBase64String([System.Security.Cryptography.RandomNumberGenerator]::GetBytes(32))"); \
		if grep -q "^JWT_SECRET=" .env; then \
			sed -i "s/^JWT_SECRET=.*/JWT_SECRET=$$JWT_VAL/" .env; \
		else \
			echo "JWT_SECRET=$$JWT_VAL" >> .env; \
		fi; \
		echo "JWT_SECRET generated and saved to .env"; \
	fi
	$(MAKE) docker-build
	$(MAKE) docker-up
	@echo ""
	@echo "Deploy complete! Services starting up..."
	@echo "  API: http://localhost:8080"
	@echo "  Web: http://localhost:3000"

# ──────────────────────────────────────────────
#  Setup
# ──────────────────────────────────────────────

## setup: Copy .env.example to .env and install dependencies
setup:
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo ".env created from .env.example — please edit with your values"; \
	else \
		echo ".env already exists, skipping copy"; \
	fi
	cd api && go mod download
	pnpm install
	@echo "Setup complete!"

# ──────────────────────────────────────────────
#  Swagger
# ──────────────────────────────────────────────

## swagger-gen: Regenerate Swagger API docs
swagger-gen:
	cd api && swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

# ──────────────────────────────────────────────
#  Database
# ──────────────────────────────────────────────

## db-reset: Full database reset (rollback all, then re-apply)
db-reset: seed-reset migrate-reset migrate-up

## seed-up: Apply all pending seeds
seed-up:
	goose -dir $(SEED_DIR) -table $(GOOSE_SEED_TABLE) $(DB_DRIVER) $(DB_STRING) up

## seed-down: Rollback the last seed
seed-down:
	goose -dir $(SEED_DIR) -table $(GOOSE_SEED_TABLE) $(DB_DRIVER) $(DB_STRING) down

## seed-reset: Rollback all seeds
seed-reset:
	goose -dir $(SEED_DIR) -table $(GOOSE_SEED_TABLE) $(DB_DRIVER) $(DB_STRING) reset

## seed-status: Show seed status
seed-status:
	goose -dir $(SEED_DIR) -table $(GOOSE_SEED_TABLE) $(DB_DRIVER) $(DB_STRING) status

## seed-create: Create a new seed file (usage: make seed-create name=add_products)
seed-create:
	@if [ -z "$(name)" ]; then echo "Error: name is required. Usage: make seed-create name=add_products"; exit 1; fi
	goose -dir $(SEED_DIR) create $(name) sql

# ──────────────────────────────────────────────
#  Clean
# ──────────────────────────────────────────────

## clean: Remove build artifacts and Docker resources
clean:
	-rm -rf api/bin
	docker compose down -v --rmi local 2>/dev/null || true
	@echo "Clean complete!"
