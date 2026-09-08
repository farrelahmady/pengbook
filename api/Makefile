# Migration commands using goose
# Install goose: go install github.com/pressly/goose/v3/cmd/goose@latest

# Database connection (reads from environment variables)
DB_DRIVER ?= postgres
DB_STRING ?= "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)"

# Migration directory
MIGRATE_DIR ?= ./migrations

.PHONY: migrate-up migrate-down migrate-status migrate-create migrate-reset help

## help: Show this help message
help:
	@echo "Usage: make [command]"
	@echo ""
	@echo "Commands:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

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
