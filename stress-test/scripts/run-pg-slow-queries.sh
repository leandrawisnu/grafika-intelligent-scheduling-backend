#!/usr/bin/env bash
# Jalankan pg-slow-queries.sql setelah k6 run.
#
# Usage:
#   DATABASE_URL=postgres://... ./scripts/run-pg-slow-queries.sh
#   # atau dari .env root GIS / backend jika DATABASE_URL diset

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ -f "$ROOT/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
  echo "Set DATABASE_URL (postgres connection string) to run slow query report." >&2
  echo "Example: DATABASE_URL=postgres://grafika:pass@172.30.200.10:5432/grafika $0" >&2
  exit 1
fi

if ! command -v psql >/dev/null 2>&1; then
  echo "psql not found" >&2
  exit 1
fi

OUT="${1:-$ROOT/results/pg-slow-queries.txt}"
mkdir -p "$(dirname "$OUT")"

echo "[pg] writing slow query snapshot to $OUT"
psql "$DATABASE_URL" -f "$ROOT/scripts/pg-slow-queries.sql" | tee "$OUT"
