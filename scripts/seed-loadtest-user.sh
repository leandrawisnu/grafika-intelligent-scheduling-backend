#!/usr/bin/env bash
# Seed a load-test account (k6) into the pengguna table.
# Idempotent: safe to re-run (upsert by email).
#
# Usage (from backend root, requires .env + psql + go):
#   ./scripts/seed-loadtest-user.sh                      # use defaults
#   ./scripts/seed-loadtest-user.sh <email> <password>   # custom
#
# Defaults: loadtest@grafika.sch.id / GrafikaLoadTest2026!
# The script prints the credentials — copy them into stress-test/.env
# (K6_USER / K6_PASSWORD).

set -euo pipefail

EMAIL="${1:-loadtest@grafika.sch.id}"
PASSWORD="${2:-GrafikaLoadTest2026!}"

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
  echo "psql not found. Install the PostgreSQL client." >&2
  exit 1
}

echo "[seed-loadtest] hashing password (argon2id) for $EMAIL"
HASH="$(go run ./scripts/hash-sandi "$PASSWORD")"

echo "[seed-loadtest] upserting user $EMAIL (role=admin) into $DB_NAME"
psql "$PSQL_URL" -v ON_ERROR_STOP=1 \
  -v email="$EMAIL" \
  -v hash="$HASH" \
  <<'SQL'
INSERT INTO pengguna (email, password_hash, peran, jurusan_id, aktif)
VALUES (:'email', :'hash', 'admin', NULL, true)
ON CONFLICT (email) DO UPDATE SET
  password_hash = EXCLUDED.password_hash,
  aktif = true,
  updated_at = now();
SQL

echo "[seed-loadtest] done. Credentials for stress-test/.env:"
echo "  K6_USER=$EMAIL"
echo "  K6_PASSWORD=$PASSWORD"
