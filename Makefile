# rare_backend Makefile

GO ?= go
MIGRATE_VERSION ?= v4.18.2
MIGRATE ?= $(shell $(GO) env GOPATH)/bin/migrate
MIGRATIONS_DIR ?= migrations
MIGRATE_DATABASE_URL ?= $(shell ./scripts/migrate_dsn.sh 2>/dev/null)

.PHONY: help migrate-install migrate-gen-baseline migrate-up migrate-down migrate-down-1 migrate-version migrate-force migrate-create db-reset

help:
	@echo "Database migration (golang-migrate):"
	@echo "  make migrate-install       Install migrate CLI"
	@echo "  make migrate-gen-baseline  Regenerate 000001 from schema_reference.sql"
	@echo "  make migrate-up            Apply pending migrations"
	@echo "  make migrate-down          Rollback one migration"
	@echo "  make migrate-down-1        Alias of migrate-down"
	@echo "  make migrate-version       Show current migration version"
	@echo "  make migrate-force V=1     Force set version (fix dirty state)"
	@echo "  make migrate-create NAME=add_foo  Create empty up/down pair"
	@echo "  make db-reset              Rollback all + migrate up (dev only)"

migrate-install:
	$(GO) install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)
	@echo "installed: $(MIGRATE)"

migrate-gen-baseline:
	python3 scripts/gen_baseline_migration.py

migrate-up: migrate-check
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(MIGRATE_DATABASE_URL)" up

migrate-down migrate-down-1: migrate-check
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(MIGRATE_DATABASE_URL)" down 1

migrate-version: migrate-check
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(MIGRATE_DATABASE_URL)" version

migrate-force: migrate-check
ifndef V
	$(error V is required, e.g. make migrate-force V=1)
endif
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(MIGRATE_DATABASE_URL)" force $(V)

migrate-create: migrate-check
ifndef NAME
	$(error NAME is required, e.g. make migrate-create NAME=add_post_column)
endif
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

db-reset: migrate-check
	-$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(MIGRATE_DATABASE_URL)" down -all
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(MIGRATE_DATABASE_URL)" up

migrate-check:
	@test -n "$(MIGRATE_DATABASE_URL)" || (echo "MYSQL_DSN missing: cp .env.example .env and set MYSQL_DSN" >&2; exit 1)
	@test -x "$(MIGRATE)" || (echo "migrate not found: run make migrate-install" >&2; exit 1)
