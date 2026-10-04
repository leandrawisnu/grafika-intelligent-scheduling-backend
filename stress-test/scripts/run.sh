#!/usr/bin/env bash
# Jalankan k6 stress test GIS.
# Usage:
#   ./scripts/run.sh                    # smoke (default)
#   ./scripts/run.sh bootstrap
#   ./scripts/run.sh grid-read
#   ./scripts/run.sh session-overhead
#   ./scripts/run.sh validation-spike
#   ./scripts/run.sh resolver-flow
#   ./scripts/run.sh plotting-crud
#   ./scripts/run.sh rollover
#   ./scripts/run.sh smoke -- --out json=results/smoke.json
#
# E2E via Next.js proxy (cookie auth, same as browser):
#   ./scripts/run.sh fe-smoke
#   ./scripts/run.sh fe-bootstrap

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCENARIO="${1:-smoke}"
shift || true

# Preserve K6_BASE_URL when caller overrides (e.g. compare-smoke.sh API direct).
_PRESET_K6_BASE_URL="${K6_BASE_URL-}"

if [[ -f "$ROOT/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

if [[ -n "$_PRESET_K6_BASE_URL" ]]; then
  export K6_BASE_URL="$_PRESET_K6_BASE_URL"
fi

if ! command -v k6 >/dev/null 2>&1; then
  echo "k6 not found. Install: https://grafana.com/docs/k6/latest/set-up/install-k6/" >&2
  exit 1
fi

SCRIPT="$ROOT/k6/smoke.js"
case "$SCENARIO" in
  smoke) SCRIPT="$ROOT/k6/smoke.js" ;;
  fe-smoke) SCRIPT="$ROOT/k6/smoke-fe.js" ;;
  bootstrap|grid-read|session-overhead|validation-spike|resolver-flow|plotting-crud|rollover)
    SCRIPT="$ROOT/k6/scenarios/${SCENARIO}.js"
    ;;
  fe-bootstrap|fe-grid-read|fe-session-overhead|fe-validation-spike|fe-resolver-flow|fe-plotting-crud)
    SCRIPT="$ROOT/k6/scenarios/${SCENARIO}.js"
    ;;
  *)
    echo "Unknown scenario: $SCENARIO" >&2
    echo "Available:" >&2
    echo "  API direct: smoke, bootstrap, grid-read, session-overhead, validation-spike, resolver-flow, plotting-crud, rollover" >&2
    echo "  FE proxy:   fe-smoke, fe-bootstrap, fe-grid-read, fe-session-overhead, fe-validation-spike, fe-resolver-flow, fe-plotting-crud" >&2
    exit 1
    ;;
esac

mkdir -p "$ROOT/results"

echo "[stress-test] scenario=$SCENARIO base=${K6_BASE_URL:-http://127.0.0.1:6060}"
exec k6 run "$SCRIPT" "$@"
