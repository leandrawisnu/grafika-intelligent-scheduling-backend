.PHONY: start dev test build migrate-init migrate-up migrate-down migrate-status migrate-force migrate-install create-migration seed-demo db-up db-down air-install

ifneq (,$(wildcard .env))
include .env
export
endif
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= grafika
DB_NAME ?= grafika
DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATIONS := database/migrations
LATEST_VERSION := $(shell ls $(MIGRATIONS)/*_*.up.sql 2>/dev/null | sed 's/.*\/\([0-9]*\)_.*/\1/' | sort -n | tail -1)

start:
	@go run ./src

air-install:
	@go install github.com/air-verse/air@latest

dev:
	@export PATH="$$(go env GOPATH)/bin:$$PATH"; \
	command -v air >/dev/null 2>&1 || { echo "Install: make air-install" >&2; exit 1; }; \
	air

test:
	@go test ./...

build:
	@go build -o server ./src

db-up:
	@docker compose -f docker-compose.db.yaml up -d

db-down:
	@docker compose -f docker-compose.db.yaml down

migrate-install:
	@command -v migrate >/dev/null 2>&1 || { echo "Install: https://github.com/golang-migrate/migrate (paket: golang-migrate)" >&2; exit 1; }

migrate-init: migrate-install
	@command -v psql >/dev/null 2>&1 || { echo "psql tidak ditemukan. Install PostgreSQL client." >&2; exit 1; }
	@test -n "$(LATEST_VERSION)" || { echo "tidak ada file migrasi di $(MIGRATIONS)" >&2; exit 1; }
	@psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f database/migrations/init.sql
	@migrate -path $(MIGRATIONS) -database "$(DATABASE_URL)" force $(LATEST_VERSION)

migrate-up: migrate-install
	@migrate -path $(MIGRATIONS) -database "$(DATABASE_URL)" up

migrate-down: migrate-install
	@migrate -path $(MIGRATIONS) -database "$(DATABASE_URL)" down 1

migrate-status: migrate-install
	@migrate -path $(MIGRATIONS) -database "$(DATABASE_URL)" version

migrate-force: migrate-install
	@test -n "$(VERSION)" || (echo "usage: make migrate-force VERSION=20260920123000" >&2; exit 1)
	@migrate -path $(MIGRATIONS) -database "$(DATABASE_URL)" force $(VERSION)

seed-demo:
	@bash scripts/seed-demo.sh

create-migration:
	@test -n "$(NAME)" || (echo "usage: make create-migration NAME=add_example" >&2; exit 1)
	@case "$(NAME)" in *[!a-z0-9_]*) echo "migration name: lowercase, angka, underscore" >&2; exit 1;; esac
	@ts=$$(date -u +%Y%m%d%H%M%S); \
	up="database/migrations/$${ts}_$(NAME).up.sql"; \
	down="database/migrations/$${ts}_$(NAME).down.sql"; \
	test ! -e "$$up" && test ! -e "$$down" || { echo "migration exists: $${ts}_$(NAME)" >&2; exit 1; }; \
	touch "$$up" "$$down"; echo "created $$up and $$down"
