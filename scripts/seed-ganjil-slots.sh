#!/usr/bin/env bash
# Seed mapel, guru, lalu slot_jadwal (template mingguan per kelas).
# Usage: ./scripts/seed-ganjil-slots.sh

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

command -v psql >/dev/null 2>&1 || { echo "psql tidak ditemukan" >&2; exit 1; }
command -v python3 >/dev/null 2>&1 || { echo "python3 tidak ditemukan" >&2; exit 1; }

run_sql() {
  echo "[seed-slots] $1"
  psql "$PSQL_URL" -v ON_ERROR_STOP=1 -f "$1"
}

# Master skeleton harus sudah ada
run_sql "$ROOT/database/seeds/ganjil_2026_2027.sql"
run_sql "$ROOT/database/seeds/mapel_guru_ganjil_2026.sql"

GEN="$ROOT/database/seeds/generated_slots.sql"
python3 "$ROOT/scripts/seed-slots.py" > "$GEN"
echo "[seed-slots] generated $(wc -l < "$GEN") lines → $GEN"
run_sql "$GEN"

echo "[seed-slots] selesai"
