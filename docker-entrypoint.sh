#!/bin/sh
set -e

export GIS_MODULE_ROOT="${GIS_MODULE_ROOT:-/app}"

if [ "${GIS_SKIP_MIGRATE}" = "true" ]; then
  exec ./main
fi

echo "Menunggu database..."
for _ in $(seq 1 30); do
  if ./main -migrate version >/tmp/gis-migrate-version 2>&1; then
    break
  fi
  sleep 2
done

if ! ./main -migrate version >/tmp/gis-migrate-version 2>&1; then
  echo "Database tidak dapat dihubungi setelah 60 detik." >&2
  cat /tmp/gis-migrate-version >&2 || true
  exit 1
fi

if grep -q "belum ada migrasi" /tmp/gis-migrate-version; then
  echo "Database kosong — menjalankan init.sql..."
  ./main -migrate init
else
  echo "Menjalankan migrate up..."
  ./main -migrate up
fi

exec ./main
