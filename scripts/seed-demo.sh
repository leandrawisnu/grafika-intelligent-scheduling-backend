#!/usr/bin/env bash
# Seed data demo — Ganjil 2026/2027 (SMKN 4 Malang).
# Satu-satunya entrypoint seeding. Idempotent: aman dijalankan ulang.
#
# Urutan:
#   1. fix_duplicate_ganjil_semester.sql — bersihkan semester duplikat
#   2. ganjil_2026_2027.sql             — skeleton (tahun ajaran, jurusan,
#                                          jam, ruangan, kelas, jadwal semester)
#   3. mapel_guru_ganjil_2026.sql       — master mapel + guru
#   4. demo_slots.sql                   — template slot mingguan per kelas
#
# Usage (dari root backend, butuh .env + psql):
#   ./scripts/seed-demo.sh            # full: skeleton + master + slot
#   ./scripts/seed-demo.sh --skeleton-only   # tanpa slot (jadwal kosong)

set -euo pipefail

SKELETON_ONLY=false
if [[ "${1:-}" == "--skeleton-only" ]]; then
  SKELETON_ONLY=true
elif [[ $# -gt 0 ]]; then
  echo "usage: $(basename "$0") [--skeleton-only]" >&2
  exit 1
fi

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

run_sql() {
  local f="$ROOT/database/seeds/$1"
  [[ -f "$f" ]] || { echo "file tidak ada: $f" >&2; exit 1; }
  echo "[seed-demo] menjalankan $1"
  psql "$PSQL_URL" -v ON_ERROR_STOP=1 -f "$f"
}

run_sql "fix_duplicate_ganjil_semester.sql"
run_sql "ganjil_2026_2027.sql"
run_sql "mapel_guru_ganjil_2026.sql"

if [[ "$SKELETON_ONLY" == "true" ]]; then
  echo "[seed-demo] selesai (skeleton saja, tanpa slot)"
  exit 0
fi

run_sql "demo_slots.sql"
echo "[seed-demo] selesai (full + slot)"
