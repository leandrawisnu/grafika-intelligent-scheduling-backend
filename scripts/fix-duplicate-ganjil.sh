#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
[[ -f .env ]] && set -a && source .env && set +a
PSQL_URL="${DATABASE_URL:-postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST:-localhost}:${DB_PORT:-5432}/${DB_NAME}?sslmode=disable}"
psql "$PSQL_URL" -v ON_ERROR_STOP=1 -f "$ROOT/database/seeds/fix_duplicate_ganjil_semester.sql"
echo "[fix] semester ganjil duplikat dibersihkan"
