#!/usr/bin/env bash
# Bandingkan summary JSON dari handleSummary (fe-smoke vs smoke-api).
#
# Usage:
#   ./scripts/summarize-results.sh
#   ./scripts/summarize-results.sh results/fe-smoke-summary.json results/smoke-api-summary.json

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FE="${1:-$ROOT/results/fe-smoke-summary.json}"
API="${2:-$ROOT/results/smoke-api-summary.json}"

if ! command -v jq >/dev/null 2>&1; then
  echo "jq required for summarize-results.sh" >&2
  exit 1
fi

echo "=== GIS smoke comparison ==="
printf "%-20s %12s %12s %12s %12s\n" "endpoint" "FE p95 ms" "API p95 ms" "FE bytes" "API bytes"
echo "--------------------------------------------------------------------------------"

TAGS=(katalog auth_sesi jadwal_list jadwal_detail jk_aktif_ringkas ringkasan)
for tag in "${TAGS[@]}"; do
  fe_p95="n/a"
  api_p95="n/a"
  fe_bytes="n/a"
  api_bytes="n/a"
  if [[ -f "$FE" ]]; then
    fe_p95=$(jq -r ".endpoints.${tag}.p95_ms // \"n/a\"" "$FE")
    fe_bytes=$(jq -r ".endpoints.${tag}.avg_bytes // \"n/a\"" "$FE" | awk '{printf "%.0f", $1}')
  fi
  if [[ -f "$API" ]]; then
    api_p95=$(jq -r ".endpoints.${tag}.p95_ms // \"n/a\"" "$API")
    api_bytes=$(jq -r ".endpoints.${tag}.avg_bytes // \"n/a\"" "$API" | awk '{printf "%.0f", $1}')
  fi
  printf "%-20s %12s %12s %12s %12s\n" "$tag" "$fe_p95" "$api_p95" "$fe_bytes" "$api_bytes"
done

echo ""
if [[ -f "$FE" ]]; then
  echo "FE overall p95: $(jq -r '.overall_p95_ms' "$FE") ms | failed: $(jq -r '.http_req_failed_rate' "$FE")"
fi
if [[ -f "$API" ]]; then
  echo "API overall p95: $(jq -r '.overall_p95_ms' "$API") ms | failed: $(jq -r '.http_req_failed_rate' "$API")"
fi
