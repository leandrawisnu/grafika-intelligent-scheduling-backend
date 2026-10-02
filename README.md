# Grafika Intelligent Scheduling — Backend

API penjadwalan SMK Grafika (Go + Fiber + PostgreSQL). Entrypoint `src/main.go`, migrasi di `database/migrations/`.

## Prasyarat

- Go 1.25+
- PostgreSQL lokal (default `localhost:5432`) — Docker **opsional**

Migrasi dari folder `src`: `go run main.go -migrate` atau `make migrate-up`.

## Konfigurasi

```bash
cp .env.example .env
```

| Variabel | Keterangan |
|----------|------------|
| `DB_*` | Postgres lokal (default **5432**) |
| `APP_HOST` / `APP_PORT` | Server (default `0.0.0.0:8080`) |
| `ML_SERVICE_URL` | Layanan ML |
| `DATABASE_URL` | Opsional override `DB_*` |

Buat user/database di Postgres lokal sesuai `.env`, lalu:

## Quick start (database baru)

```bash
make migrate-init
make seed-demo   # opsional: data demo Ganjil 2026/2027 (master + jadwal + slot)
make start
curl -s http://127.0.0.1:8080/health
```

**Opsional — Postgres via Docker** (hindari bentrok port: set `DB_PORT=5433` di `.env`):

```bash
make db-up
```

## Makefile

| Target | Fungsi |
|--------|--------|
| `make dev` | Live reload dengan [Air](https://github.com/air-verse/air) (`.air.toml`) |
| `make start` | `go run ./src` (tanpa reload) |
| `make air-install` | `go install` Air ke `$(go env GOPATH)/bin` |
| `make migrate-init` / `migrate-up` / … | Migrasi |
| `make seed-demo` | Data demo Ganjil 2026/2027 (master + jadwal + slot); `--skeleton-only` via `./scripts/seed-demo.sh --skeleton-only` |
| `make db-up` / `db-down` | Postgres Docker saja (opsional) |

## Docker API (`web-api`)

```bash
docker network create main-network
docker compose up -d --build
```

`DB_HOST=host.docker.internal` di container; port DB mengikuti `.env`.

## Coolify (backend only)

Polyrepo ini deploy ke Coolify per-folder — backend pakai `docker-compose.coolify.yaml`
(isinya 3 service internal: `backend` + `postgres` + `minio`, tanpa
`container_name`/`ports`, data persisten di volume `gis_pgdata`/`gis_minio`).

1. Coolify → New Resource → Docker Compose → pilih repo + branch.
2. Base Directory arahkan ke `grafika-intelligent-scheduling-backend/`,
   Compose file: `docker-compose.coolify.yaml`.
3. Environment Variables: copy dari `.env.coolify.example`
   (wajib: `DB_PASSWORD`, `MINIO_ROOT_PASSWORD`).
4. Domain → service `backend`, port `8080`. Cek via `/health`.
5. Deploy — migrasi jalan otomatis via `docker-entrypoint.sh`,
   bucket MinIO dibuat otomatis oleh backend.
   Set `GIS_SKIP_MIGRATE=true` untuk skip bila perlu.

## API

- `GET /health`
- `/api/v1` — [`src/router/router.go`](src/router/router.go)
