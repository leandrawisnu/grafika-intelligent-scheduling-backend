# GIS Stress Test (k6)

Load test untuk Grafika Intelligent Scheduling — mengikuti business process terkini:
plotting guru → jadwal kelas/slot → validasi konflik → conflict resolver (AI) →
publikasi, plus rollover semester (admin).

Dua mode:

| Mode | Auth | Target `K6_BASE_URL` | Skrip |
|------|------|----------------------|-------|
| **API direct** | `POST /api/v1/auth/masuk` → `X-GIS-Session` | Backend (`:6060`, `:8080`) | `smoke`, `bootstrap`, … |
| **FE end-to-end** | `POST /api/auth/login` → cookie → `/api/v1/*` proxy | Frontend origin (`:3000`, `https://domain`) | `fe-smoke`, `fe-bootstrap`, … |

ML endpoints (`prediksi-konflik`, `ai/tanya`, `konflik/:id/jelaskan`) **tidak
disertakan** — butuh layanan ML :8000 terpisah.

## Prerequisites

- [k6](https://grafana.com/docs/k6/latest/set-up/install-k6/) installed
- Backend seeded (dari root backend):
  ```bash
  make migrate-init && make seed-demo
  ```
  (`seed-demo` = skeleton + master + slot mingguan, idempotent. Untuk jadwal
  kosong tanpa slot: `./scripts/seed-demo.sh --skeleton-only`)
- **FE mode**: frontend harus jalan (`http://127.0.0.1:3000` atau domain publik lewat Caddy → frontend)
- Load-test user credentials (hindari admin produksi)
- **Rate limit**: backend membatasi endpoint berat per pengguna (`GIS_RATE_LIMIT`, default 60 permintaan/menit). Skenario `resolver-flow` dan `plotting-crud` bisa melampaui itu — naikkan di `.env` backend (`GIS_RATE_LIMIT=600`) atau matikan (`GIS_RATE_LIMIT=0`) saat load test.

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

## Daftar test case

| # | Skenario | VU profile | Endpoint yang diukur | Apa yang diuji |
|---|----------|-----------|----------------------|----------------|
| 1 | `smoke` / `fe-smoke` | 1 × 30s | `katalog` (ETag), `auth_sesi`, `jadwal_list`, `jadwal_detail`, `jk_aktif_ringkas`, `ringkasan` | Bootstrap batch gate — permintaan pertama saat halaman jadwal dibuka |
| 2 | `bootstrap` / `fe-bootstrap` | 5→50 | + `konflik` | Catalog + jadwal load lengkap (matching browser jadwal-context) |
| 3 | `grid-read` / `fe-grid-read` | 3→10 | `jadwal_kelas_grid` (preload dalam), `jk_aktif_full` | Heavy read: grid jadwal kelas penuh |
| 4 | `session-overhead` / `fe-session-overhead` | 10→40 | `auth_sesi`, `katalog`, `jadwal_list`, `ringkasan` (acak) | Overhead middleware sesi — tiap request hit DB lookup sesi |
| 5 | `validation-spike` / `fe-validation-spike` | 1→3 | `validasi_rearm` (PUT slot no-op), `validasi` (POST validasi) | CPU + DB heavy: **deteksi konflik penuh** atas semua slot aktif. Tiap iterasi di-arm ulang (`perlu_validasi=true`) via PUT slot no-op — tanpa re-arm, iterasi 2+ hanya membaca cache konflik |
| 6 | `resolver-flow` / `fe-resolver-flow` | 1→3 | `konflik_list`, `resolver_selesaikan`, `resolver_resolusi`, `resolver_terima` | **BP conflict resolver**: list konflik terbuka → minta alternatif AI (deterministik, tanpa ML) → lihat resolusi → terima (menerapkan perubahan + deteksi ulang). Konflik dijamin via `demo-konflik` |
| 7 | `plotting-crud` / `fe-plotting-crud` | 3→10 | `plotting_list`, `plotting_create`, `plotting_update`, `plotting_delete` | **BP plotting guru**: siklus CRUD penuh per iterasi (bersih, tanpa sisa data) |
| 8 | `rollover` (API direct saja) | 1 × 1m | `rollover_salin` (`POST /semester/:id/salin`) | **BP rollover semester**: salin snapshot master (jurusan, ruangan, kelas, plotting) + jadwal kosong dalam satu transaksi. Operasi admin sekali/semester → benchmark 1 VU. **Menyisakan data** — bersihkan dengan `cleanup-stress-semester.sh` |

Catatan data:

- `resolver-flow` memakai `POST /demo-konflik` (khusus semester bernama "demo")
  dan **memutasi data demo**: guru slot diacak, dan resolusi yang diterima
  menugaskan ulang guru. Jangan jalankan di data asli.
- `validation-spike` menulis slot (no-op) tiap iterasi.
- `rollover` membuat tahun ajaran + semester baru tiap iterasi.

## Run — API direct

Backend langsung ke Go (tanpa Next.js):

```bash
chmod +x scripts/run.sh

./scripts/run.sh smoke
./scripts/run.sh bootstrap
./scripts/run.sh grid-read
./scripts/run.sh session-overhead
./scripts/run.sh validation-spike
./scripts/run.sh resolver-flow
./scripts/run.sh plotting-crud
./scripts/run.sh rollover
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
./scripts/run.sh fe-resolver-flow
./scripts/run.sh fe-plotting-crud
```

**Catatan routing production:** skrip FE butuh request masuk ke **frontend** (Next.js). Kalau Caddy mengarahkan semua `/api/*` langsung ke Go, `/api/auth/login` tidak akan jalan — pakai domain yang lewat frontend, atau tes dari dalam VM ke `http://frontend:3000`.

## Urutan run yang disarankan

```bash
# 1. Pastikan data demo ada + konflik demo tersuntik
./scripts/run.sh smoke          # kering, cek semua endpoint 2xx

# 2. Read path
./scripts/run.sh bootstrap
./scripts/run.sh grid-read
./scripts/run.sh session-overhead

# 3. BP interaktif
./scripts/run.sh plotting-crud
./scripts/run.sh resolver-flow

# 4. Heavy (deteksi konflik) — hindari di produksi tanpa backup
./scripts/run.sh validation-spike

# 5. Benchmark admin (jangan lupa cleanup setelahnya)
./scripts/run.sh rollover
DATABASE_URL=... ./scripts/cleanup-stress-semester.sh
```

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

## Slow query analysis

```bash
DATABASE_URL=postgres://... ./scripts/run-pg-slow-queries.sh
# atau: psql "$DATABASE_URL" -f scripts/pg-slow-queries.sql
```

Butuh `pg_stat_statements` aktif di Postgres target.

## Cleanup sisa stress test

Skenario `rollover` menyisakan semester + tahun ajaran (FK tanpa CASCADE,
tidak bisa dihapus lewat API). Setelah run:

```bash
DATABASE_URL=postgres://... ./scripts/cleanup-stress-semester.sh
```

Menghapus: semester/tahun ajaran `stress-%` beserta plotting, kelas, ruangan,
jurusan, jadwal_semester (mengkaskade slot/konflik/resolusi), dan baris
`hari_libur_guru` hasil `demo-konflik`.

## Structure

```
stress-test/
├── .env.example
├── k6/
│   ├── lib/
│   │   ├── auth.js            # API direct (X-GIS-Session)
│   │   ├── auth-fe.js         # FE cookie + proxy
│   │   ├── bootstrap-smoke.js # batch bootstrap bersama
│   │   ├── config.js          # env → config
│   │   ├── katalog-etag.js    # cache ETag per VU
│   │   ├── options.js         # stages + thresholds
│   │   ├── smoke-metrics.js   # response size
│   │   └── summary.js         # handleSummary → JSON
│   ├── scenarios/
│   │   ├── bootstrap.js           fe-bootstrap.js
│   │   ├── grid-read.js           fe-grid-read.js
│   │   ├── session-overhead.js    fe-session-overhead.js
│   │   ├── validation-spike.js    fe-validation-spike.js
│   │   ├── resolver-flow.js       fe-resolver-flow.js
│   │   ├── plotting-crud.js       fe-plotting-crud.js
│   │   └── rollover.js            (API direct saja)
│   ├── smoke.js
│   └── smoke-fe.js
└── scripts/
    ├── run.sh
    ├── compare-smoke.sh
    ├── summarize-results.sh
    ├── run-pg-slow-queries.sh
    ├── pg-slow-queries.sql
    └── cleanup-stress-semester.sh
```
