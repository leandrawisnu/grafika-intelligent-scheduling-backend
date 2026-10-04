import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargetsFe, feOpts, feSession, loginViaFe } from "../lib/auth-fe.js";
import { apiUrl, getConfig, requireCredentials } from "../lib/config.js";
import { stagesPlotting, thresholdsPlotting } from "../lib/options.js";

/**
 * E2E CRUD Plotting guru via Next.js /api/v1 proxy:
 *   list → tambah → perbarui → hapus. Siklus penuh per iterasi,
 *   tidak menyisakan data.
 */
export const options = {
  stages: stagesPlotting,
  thresholds: thresholdsPlotting,
};

export function setup() {
  const cfg = getConfig();
  requireCredentials(cfg);
  const session = loginViaFe(cfg);
  const targets = discoverTargetsFe(cfg, session);

  if (!targets.jadwalSemesterId) {
    throw new Error("No jadwal semester found — run backend seed first");
  }

  const opts = feOpts(session);
  const kat = http.get(
    apiUrl(cfg.baseUrl, `/api/v1/katalog?semester_id=${targets.jadwalSemesterId}`),
    { ...opts, tags: { name: "setup_katalog" } }
  );
  if (kat.status !== 200) {
    throw new Error(`Gagal muat katalog (${kat.status}): ${kat.body}`);
  }
  const k = kat.json();
  const ids = (rows) =>
    (Array.isArray(rows) ? rows : []).map((r) => r.id).filter(Boolean);

  const palette = {
    kelas: ids(k.kelas),
    hari: ids((k.hari || []).filter((h) => !h.akhir_pekan)),
    jam: ids((k.jam_pelajaran || []).filter((j) => !j.istirahat)),
    mapel: ids(k.mata_pelajaran),
    guru: ids((k.guru || []).filter((g) => g.aktif !== false)),
    ruangan: ids((k.ruangan || []).filter((r) => r.aktif !== false)),
  };

  if (
    palette.kelas.length === 0 ||
    palette.hari.length === 0 ||
    palette.jam.length === 0 ||
    palette.mapel.length === 0 ||
    palette.guru.length === 0
  ) {
    throw new Error(
      "Katalog tidak lengkap (butuh kelas/hari/jam/mapel/guru) — run make seed-demo first"
    );
  }

  return { cfg, ...targets, palette };
}

export default function (data) {
  const opts = feOpts(feSession(data.cfg));
  const jsId = data.jadwalSemesterId;
  const p = data.palette;
  const n = __ITER;
  const pick = (arr) => arr[n % arr.length];

  // 1. List plotting semester.
  const list = http.get(
    apiUrl(data.cfg.baseUrl, `/api/v1/semester/${jsId}/plotting`),
    { ...opts, tags: { name: "plotting_list" } }
  );
  check(list, { "plotting list 2xx": (r) => r.status >= 200 && r.status < 300 });

  // 2. Tambah satu baris plotting.
  const body = {
    kelas_id: pick(p.kelas),
    hari_id: pick(p.hari),
    jam_pelajaran_id: pick(p.jam),
    mata_pelajaran_id: pick(p.mapel),
    guru_id: pick(p.guru),
  };
  if (p.ruangan.length > 0) body.ruangan_id = pick(p.ruangan);

  const buat = http.post(
    apiUrl(data.cfg.baseUrl, `/api/v1/semester/${jsId}/plotting`),
    JSON.stringify(body),
    { ...opts, tags: { name: "plotting_create" } }
  );
  check(buat, {
    "plotting create 2xx": (r) => r.status >= 200 && r.status < 300,
  });
  if (buat.status < 200 || buat.status >= 300) {
    sleep(1);
    return;
  }
  const plottingId = buat.json("id");

  // 3. Perbarui (ganti guru; PerbaruiPlotting memakai body lengkap).
  if (plottingId && p.guru.length > 1) {
    const perbarui = http.put(
      apiUrl(data.cfg.baseUrl, `/api/v1/plotting/${plottingId}`),
      JSON.stringify({ ...body, guru_id: p.guru[(n + 1) % p.guru.length] }),
      { ...opts, tags: { name: "plotting_update" } }
    );
    check(perbarui, {
      "plotting update 2xx": (r) => r.status >= 200 && r.status < 300,
    });
  }

  // 4. Hapus — siklus bersih, tidak menyisakan data.
  if (plottingId) {
    const hapus = http.del(
      apiUrl(data.cfg.baseUrl, `/api/v1/plotting/${plottingId}`),
      null,
      { ...opts, tags: { name: "plotting_delete" } }
    );
    check(hapus, {
      "plotting delete 2xx": (r) => r.status >= 200 && r.status < 300,
    });
  }

  sleep(0.5 + Math.random() * 0.5);
}
