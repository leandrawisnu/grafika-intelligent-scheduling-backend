# GIS Stress Test (k6)

Load test untuk Grafika Intelligent Scheduling. Dua mode:

| Mode | Auth | Target `K6_BASE_URL` | Skrip |
|------|------|----------------------|-------|
| **API direct** | `POST /api/v1/auth/masuk` → `X-GIS-Session` | Backend (`:6060`, `:8080`) | `smoke`, `bootstrap`, … |
| **FE end-to-end** | `POST /api/auth/login` → cookie → `/api/v1/*` proxy | Frontend origin (`:3000`, `https://domain`) | `fe-smoke`, `fe-bootstrap`, … |

ML endpoints **not included**.

## Prerequisites

- [k6](https://grafana.com/docs/k6/latest/set-up/install-k6/) installed
- Backend seeded (dari root repo ini):
  ```bash
  make migrate-init && make seed-ganjil && make seed-slots
  ```
- **FE mode**: frontend harus jalan (`http://127.0.0.1:3000` atau domain publik lewat Caddy → frontend)
- Load-test user credentials (hindari admin produksi)

## Setup

```bash
cd stress-test
cp .env.example .env
# edit K6_BASE_URL, K6_USER, K6_PASSWORD
```

| Variable | Description |
|----------|-------------|
| `K6_BASE_URL` | API direct: backend URL. FE mode: frontend origin (bukan port backend) |
| `K6_USER` | Login email |
| `K6_PASSWORD` | Login password |
| `K6_JADWAL_SEMESTER_ID` | Optional — auto-discovered |
| `K6_JADWAL_KELAS_ID` | Optional — auto-discovered |

## Run — API direct

Backend langsung ke Go (tanpa Next.js):

```bash
chmod +x scripts/run.sh

./scripts/run.sh smoke
./scripts/run.sh bootstrap
./scripts/run.sh grid-read
./scripts/run.sh session-overhead
./scripts/run.sh validation-spike
```

Expose backend (opsional): merge `docker-compose.ports.yaml` pada deploy full-stack VM (dari folder GIS: `-f grafika-intelligent-scheduling-backend/docker-compose.ports.yaml`), atau standalone `APP_BIND=0.0.0.0 docker compose up -d`.

## Run — FE end-to-end

Sama seperti browser: login BFF → cookie → `/api/v1/*` lewat Next.js proxy.

```bash
# .env: K6_BASE_URL=http://127.0.0.1:3000  (atau https://domain)

./scripts/run.sh fe-smoke
./scripts/run.sh fe-bootstrap
./scripts/run.sh fe-grid-read
./scripts/run.sh fe-session-overhead
./scripts/run.sh fe-validation-spike
```

**Catatan routing production:** skrip FE butuh request masuk ke **frontend** (Next.js). Kalau Caddy mengarahkan semua `/api/*` langsung ke Go, `/api/auth/login` tidak akan jalan — pakai domain yang lewat frontend, atau tes dari dalam VM ke `http://frontend:3000`.

## Export & analisis latency

Smoke tests emit `results/fe-smoke-summary.json` (per-endpoint p95 + avg bytes).

```bash
./scripts/run.sh fe-smoke
./scripts/summarize-results.sh results/fe-smoke-summary.json results/smoke-api-summary.json
```

Bandingkan FE proxy vs API direct:

```bash
K6_BASE_URL_FE=https://your-domain \
K6_BASE_URL_API=http://127.0.0.1:8080 \
./scripts/compare-smoke.sh
```

Lihat `results/BOTTLENECK.md` untuk analisis terbaru.

## Export hasil mentah

```bash
./scripts/run.sh fe-bootstrap -- --out json=results/fe-bootstrap.json
```

## Scenarios

| Script | VU profile | What it tests |
|--------|------------|---------------|
| `smoke.js` / `smoke-fe.js` | 1 × 30s | Bootstrap batch gate |
| `bootstrap.js` / `fe-bootstrap.js` | 5→50 | Catalog + jadwal load |
| `grid-read.js` / `fe-grid-read.js` | 3→10 | `GET /jadwal-kelas/:id` |
| `session-overhead.js` / `fe-session-overhead.js` | 10→40 | Session reads |
| `validation-spike.js` / `fe-validation-spike.js` | 1→3 | `POST .../validasi` |

FE scripts use `/api/auth/sesi` (BFF) instead of `/api/v1/auth/sesi`, matching `sesi-context.tsx`. Each VU logs in once via `feSession()` (cookie stays in VU scope; `setup()` only discovers jadwal IDs).

Thresholds: `k6/lib/options.js`.

## Slow query analysis

```bash
DATABASE_URL=postgres://... ./scripts/run-pg-slow-queries.sh
# atau: psql "$DATABASE_URL" -f scripts/pg-slow-queries.sql
```

## Structure

```
stress-test/
├── .env.example
├── k6/
│   ├── lib/
│   │   ├── auth.js       # API direct
│   │   └── auth-fe.js    # FE cookie + proxy
│   ├── scenarios/
│   │   ├── bootstrap.js
│   │   ├── fe-bootstrap.js
│   │   └── …
│   ├── smoke.js
│   └── smoke-fe.js
└── scripts/run.sh
```
