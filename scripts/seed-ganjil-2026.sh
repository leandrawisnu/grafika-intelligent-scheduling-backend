#!/usr/bin/env bash
# Seed skeleton Ganjil 2026/2027 — master + jadwal_semester + kelas (slot kosong).
# Usage: ./scripts/seed-ganjil-2026.sh   (dari root backend, butuh .env + psql)

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

if [[ -n "${DATABASE_URL:-}" ]]; then
  PSQL_URL="$DATABASE_URL"
else
  : "${DB_USER:?set DB_USER in .env}"
  : "${DB_PASSWORD:?set DB_PASSWORD in .env}"
  : "${DB_NAME:?set DB_NAME in .env}"
  DB_HOST="${DB_HOST:-localhost}"
  DB_PORT="${DB_PORT:-5432}"
  PSQL_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"
fi

command -v psql >/dev/null 2>&1 || {
  echo "psql tidak ditemukan. Install PostgreSQL client." >&2
  exit 1
}

SQL="$ROOT/database/seeds/ganjil_2026_2027.sql"
[[ -f "$SQL" ]] || { echo "file tidak ada: $SQL" >&2; exit 1; }

echo "[seed] menjalankan $SQL"
psql "$PSQL_URL" -v ON_ERROR_STOP=1 -f "$SQL"
echo "[seed] ganjil-2026 selesai"
