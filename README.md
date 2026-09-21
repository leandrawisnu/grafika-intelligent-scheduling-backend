# Grafika Intelligent Scheduling — Backend

API penjadwalan SMK Grafika (Go + Fiber + PostgreSQL). Struktur mengikuti pola **KAI BIOP**: entrypoint `src/main.go`, kode di `src/`, migrasi di `database/migrations/`.

## Prasyarat

- Go 1.25+
- Docker

Migrasi lewat `go run ./src -migrate …` atau `make migrate-*` (tanpa CLI `migrate` terpisah).

## Konfigurasi

```bash
cp .env.example .env
```

| Variabel | Keterangan |
|----------|------------|
| `APP_HOST` / `APP_PORT` | Bind server (default `0.0.0.0:8080`) |
| `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT` | Postgres (default port **5433** = `docker-compose.db.yaml`) |
| `DATABASE_URL` | Opsional; override `DB_*` |
| `ML_SERVICE_URL` | Layanan ML |
| `GIS_AUTO_MIGRATE` | `true` = GORM AutoMigrate saat start (default off setelah init) |
| `MIGRATIONS_PATH` | Default `database/migrations` (hanya file `*.up.sql` / `*.down.sql`) |

## Quick start (database baru)

```bash
make db-up
make migrate-init
make start
curl -s http://127.0.0.1:8080/health
```

`migrate-init` menjalankan [`database/migrations/init.sql`](database/migrations/init.sql) lalu force versi `20260920123000`.

## Makefile

| Target | Fungsi |
|--------|--------|
| `make db-up` / `db-down` | Postgres dev |
| `make start` | `go run ./src` |
| `make migrate-init` | `init.sql` + force versi |
| `make migrate-up` / `migrate-down` / `migrate-status` / `migrate-force` | Migrasi inkremental |
| `make create-migration NAME=…` | Pasangan `.up.sql` / `.down.sql` baru |

## Migrasi

```bash
go run ./src -migrate init
go run ./src -migrate up
go run ./src -migrate version
go run ./src -migrate force 20260920123000
```

**Jangan** `init` lalu `up` tanpa `force` — schema di `init.sql` sudah mencakup perubahan migrasi terakhir.

## Docker API

```bash
docker network create main-network   # sekali
docker compose up -d --build
```

## API

- `GET /health`
- REST di `/api/v1` — [`src/router/router.go`](src/router/router.go)

## Schema

- **Sumber kebenaran:** `database/migrations/init.sql` + migrasi berversi di folder yang sama.
- **GORM AutoMigrate:** hanya jika `GIS_AUTO_MIGRATE=true`.

## Build

```bash
make build    # menghasilkan ./server
```
