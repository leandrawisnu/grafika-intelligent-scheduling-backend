#!/usr/bin/env bash
# Bersihkan sisa data stress test dari DB:
#   - semester + tahun ajaran hasil skenario rollover (nama "stress-*")
#   - baris hari_libur_guru yang disuntik demo-konflik (alasan "demo")
#
# FK jurusan/ruangan/kelas/plotting → semester TANPA CASCADE, jadi
# urutan hapus penting dan tidak bisa lewat API (DELETE /semester/:id
# gagal diam-diam bila masih punya anak).
#
# Usage (dari root backend):
#   DATABASE_URL=postgres://... ./stress-test/scripts/cleanup-stress-semester.sh
#   # atau: ./stress-test/scripts/cleanup-stress-semester.sh   # bila .env ada

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

if [[ -f "$ROOT/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

if [[ -z "${DATABASE_URL:-}" ]]; then
  : "${DB_USER:?set DATABASE_URL atau DB_USER/DB_PASSWORD/DB_NAME di .env}"
  : "${DB_PASSWORD:?}"
  : "${DB_NAME:?}"
  DB_HOST="${DB_HOST:-localhost}"
  DB_PORT="${DB_PORT:-5432}"
  DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"
fi

command -v psql >/dev/null 2>&1 || {
  echo "psql tidak ditemukan" >&2
  exit 1
}

SQL="$(mktemp /tmp/gis-stress-cleanup.XXXXXX.sql)"
trap 'rm -f "$SQL"' EXIT

cat >"$SQL" <<'EOF'
-- Urutan penting: FK tanpa CASCADE dihapus duluan.
-- 1. plotting (semester_id REFERENCES semester, tanpa CASCADE)
DELETE FROM plotting
WHERE semester_id IN (SELECT id FROM semester WHERE nama LIKE 'stress-%');

-- 2. jadwal_semester (mengkaskade jadwal_kelas, slot_jadwal,
--    konflik, resolusi_ai, jadwal_semester_jurusan lewat FK CASCADE)
DELETE FROM jadwal_semester
WHERE semester_id IN (SELECT id FROM semester WHERE nama LIKE 'stress-%');

-- 3. kelas, ruangan, jurusan (REFERS semester tanpa CASCADE;
--    slot/jadwal_kelas sudah hilang lewat kaskade langkah 2)
DELETE FROM kelas
WHERE semester_id IN (SELECT id FROM semester WHERE nama LIKE 'stress-%');
DELETE FROM ruangan
WHERE semester_id IN (SELECT id FROM semester WHERE nama LIKE 'stress-%');
DELETE FROM jurusan
WHERE semester_id IN (SELECT id FROM semester WHERE nama LIKE 'stress-%');

-- 4. semester, lalu tahun ajaran pembungkusnya
DELETE FROM semester WHERE nama LIKE 'stress-%';
DELETE FROM tahun_ajaran WHERE nama LIKE 'stress-%';

-- 5. hari libur guru yang disuntik demo-konflik
DELETE FROM hari_libur_guru WHERE alasan = 'demo';
EOF

echo "[cleanup] menghapus sisa stress test (semester 'stress-%', hari libur demo)"
psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$SQL"
echo "[cleanup] selesai"
