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

  // Katalog dan endpoint plotting memakai semester id (bukan jadwal_semester id).
  const detail = http.get(
    apiUrl(cfg.baseUrl, `/api/v1/jadwal-semester/${targets.jadwalSemesterId}`),
    { ...opts, tags: { name: "setup_jadwal_detail" } }
  );
  if (detail.status !== 200) {
    throw new Error(
      `Gagal muat detail jadwal (${detail.status}): ${detail.body}`
    );
  }
  const semesterId = detail.json("semester_id");

  const kat = http.get(
    apiUrl(cfg.baseUrl, `/api/v1/katalog?semester_id=${semesterId}`),
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

  return { cfg, ...targets, palette, semesterId };
}

export default function (data) {
  const opts = feOpts(feSession(data.cfg));
  const semId = data.semesterId;
  const p = data.palette;
  const n = __ITER;
  const pick = (arr, geser = 0) => arr[(n + geser) % arr.length];

  // 1. List plotting semester.
  const list = http.get(
    apiUrl(data.cfg.baseUrl, `/api/v1/semester/${semId}/plotting`),
    { ...opts, tags: { name: "plotting_list" } }
  );
  check(list, { "plotting list 2xx": (r) => r.status >= 200 && r.status < 300 });

  // Slot (kelas, hari, jam) yang sudah terisi. Create dan update
  // harus ke slot kosong — constraint uni_plotting_sel menolak duplikat,
  // dan PerbaruiPlotting memakai service BuatPlotting (tanpa pengecualian
  // baris sendiri), jadi update juga harus pindah slot.
  const terisi = new Set();
  const baris = list.json();
  for (const r of Array.isArray(baris) ? baris : []) {
    if (r.kelas_id && r.hari_id && r.jam_pelajaran_id) {
      terisi.add(`${r.kelas_id}|${r.hari_id}|${r.jam_pelajaran_id}`);
    }
  }
  const cariSlotBebas = (geser) => {
    for (let k = 0; k < p.kelas.length; k++) {
      const kelas = p.kelas[(k + geser) % p.kelas.length];
      for (let h = 0; h < p.hari.length; h++) {
        const hari = p.hari[(h + geser) % p.hari.length];
        for (let j = 0; j < p.jam.length; j++) {
          const jam = p.jam[(j + geser) % p.jam.length];
          if (!terisi.has(`${kelas}|${hari}|${jam}`)) {
            return { kelas_id: kelas, hari_id: hari, jam_pelajaran_id: jam };
          }
        }
      }
    }
    return null;
  };

  // 2. Tambah satu baris plotting ke slot kosong.
  const slotA = cariSlotBebas(n % Math.max(p.kelas.length, 1));
  if (!slotA) {
    sleep(1);
    return;
  }
  const body = {
    ...slotA,
    mata_pelajaran_id: pick(p.mapel),
    guru_id: pick(p.guru),
  };
  if (p.ruangan.length > 0) body.ruangan_id = pick(p.ruangan);

  const buat = http.post(
    apiUrl(data.cfg.baseUrl, `/api/v1/semester/${semId}/plotting`),
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
  terisi.add(`${slotA.kelas_id}|${slotA.hari_id}|${slotA.jam_pelajaran_id}`);

  // 3. Perbarui — pindah ke slot kosong lain (guru ikut berganti).
  if (plottingId && p.guru.length > 1) {
    const slotB = cariSlotBebas((n + 1) % Math.max(p.kelas.length, 1));
    if (slotB) {
      const perbaruiBody = {
        ...slotB,
        mata_pelajaran_id: pick(p.mapel, 1),
        guru_id: pick(p.guru, 1),
      };
      if (p.ruangan.length > 0) perbaruiBody.ruangan_id = pick(p.ruangan, 1);
      const perbarui = http.put(
        apiUrl(data.cfg.baseUrl, `/api/v1/plotting/${plottingId}`),
        JSON.stringify(perbaruiBody),
        { ...opts, tags: { name: "plotting_update" } }
      );
      check(perbarui, {
        "plotting update 2xx": (r) => r.status >= 200 && r.status < 300,
      });
    }
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
