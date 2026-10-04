#!/usr/bin/env bash
# Bandingkan fe-smoke (proxy) vs smoke (API direct). Butuh dua .env atau override K6_BASE_URL.
#
# Usage:
#   ./scripts/compare-smoke.sh
#   K6_BASE_URL_FE=https://domain K6_BASE_URL_API=http://127.0.0.1:8080 ./scripts/compare-smoke.sh
#
# Hasil: results/fe-smoke.json, results/smoke-api.json, results/compare-summary.txt

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
mkdir -p "$ROOT/results"

if [[ -f "$ROOT/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

FE_URL="${K6_BASE_URL_FE:-$K6_BASE_URL}"
API_URL="${K6_BASE_URL_API:-http://127.0.0.1:8080}"

if [[ -z "${K6_USER:-}" || -z "${K6_PASSWORD:-}" ]]; then
  echo "Set K6_USER and K6_PASSWORD in stress-test/.env" >&2
  exit 1
fi

echo "[compare] FE proxy  → $FE_URL"
K6_BASE_URL="$FE_URL" "$ROOT/scripts/run.sh" fe-smoke || true

echo "[compare] API direct → $API_URL"
K6_BASE_URL="$API_URL" "$ROOT/scripts/run.sh" smoke || true

"$ROOT/scripts/summarize-results.sh" \
  "$ROOT/results/fe-smoke-summary.json" \
  "$ROOT/results/smoke-api-summary.json" \
  | tee "$ROOT/results/compare-summary.txt"
