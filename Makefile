.PHONY: start dev test build migrate-init migrate-up migrate-down migrate-status migrate-force create-migration seed-ganjil db-up db-down air-install

ifneq (,$(wildcard .env))
include .env
export
endif

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

migrate-init:
	@go run ./src -migrate init

migrate-up:
	@go run ./src -migrate up

migrate-down:
	@go run ./src -migrate down

migrate-status:
	@go run ./src -migrate version

migrate-force:
	@test -n "$(VERSION)" || (echo "usage: make migrate-force VERSION=20260920123000" >&2; exit 1)
	@go run ./src -migrate force $(VERSION)

seed-ganjil:
	@bash scripts/seed-ganjil-2026.sh

seed-slots:
	@bash scripts/seed-ganjil-slots.sh

create-migration:
	@test -n "$(NAME)" || (echo "usage: make create-migration NAME=add_example" >&2; exit 1)
	@case "$(NAME)" in *[!a-z0-9_]*) echo "migration name: lowercase, angka, underscore" >&2; exit 1;; esac
	@ts=$$(date -u +%Y%m%d%H%M%S); \
	up="database/migrations/$${ts}_$(NAME).up.sql"; \
	down="database/migrations/$${ts}_$(NAME).down.sql"; \
	test ! -e "$$up" && test ! -e "$$down" || { echo "migration exists: $${ts}_$(NAME)" >&2; exit 1; }; \
	touch "$$up" "$$down"; echo "created $$up and $$down"
