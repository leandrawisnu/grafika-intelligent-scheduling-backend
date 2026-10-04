import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargets, login, sessionHeaders } from "../lib/auth.js";
import { apiUrl, getConfig, requireCredentials } from "../lib/config.js";
import { stagesRollover, thresholdsRollover } from "../lib/options.js";

/**
 * Benchmark rollover semester (business process baru):
 *   POST /semester/:id/salin menyalin snapshot master (jurusan, ruangan,
 *   kelas, plotting) ke semester baru + jadwal_semester kosong dalam satu
 *   transaksi.
 *
 * Operasi admin sekali per semester → 1 VU. Setiap iterasi membuat semester
 * baru dengan nama unik. Sisa hasil run dibersihkan lewat
 * scripts/cleanup-stress-semester.sh (FK tanpa CASCADE, API tidak bisa).
 */
export const options = {
  stages: stagesRollover,
  thresholds: thresholdsRollover,
};

export function setup() {
  const cfg = getConfig();
  requireCredentials(cfg);
  const token = login(cfg);
  const targets = discoverTargets(cfg, token);

  if (!targets.jadwalSemesterId) {
    throw new Error("No jadwal semester found — run backend seed first");
  }

  const headers = sessionHeaders(token);

  // Master semester ID (bukan jadwal semester ID) — didapat dari detail jadwal.
  const detail = http.get(
    apiUrl(cfg.baseUrl, `/api/v1/jadwal-semester/${targets.jadwalSemesterId}`),
    { headers, tags: { name: "setup_jadwal_detail" } }
  );
  if (detail.status !== 200) {
    throw new Error(`Gagal muat detail jadwal (${detail.status}): ${detail.body}`);
  }
  const semesterSumber = detail.json("semester_id");
  if (!semesterSumber) {
    throw new Error("Detail jadwal tidak mengembalikan semester_id");
  }

  // Tahun ajaran khusus stress supaya (tahun_ajaran, semester_ke) unik per run.
  const runId = `stress-${new Date()
    .toISOString()
    .replace(/[-:T.]/g, "")
    .slice(0, 14)}`;
  const ta = http.post(
    apiUrl(cfg.baseUrl, "/api/v1/tahun-ajaran"),
    JSON.stringify({ nama: runId }),
    { headers, tags: { name: "setup_tahun_ajaran" } }
  );
  if (ta.status !== 201) {
    throw new Error(`Gagal buat tahun ajaran stress (${ta.status}): ${ta.body}`);
  }

  return {
    cfg,
    token,
    ...targets,
    semesterSumber,
    tahunAjaranId: ta.json("id"),
    runId,
  };
}

export default function (data) {
  const headers = sessionHeaders(data.token);

  const salin = http.post(
    apiUrl(data.cfg.baseUrl, `/api/v1/semester/${data.semesterSumber}/salin`),
    JSON.stringify({
      tahun_ajaran_id: data.tahunAjaranId,
      semester_ke: 1,
      nama: `${data.runId}-${__ITER}`,
    }),
    { headers, tags: { name: "rollover_salin" } }
  );
  check(salin, {
    "rollover salin 2xx": (r) => r.status >= 200 && r.status < 300,
  });

  // Operasi berat + transaksi → jeda panjang antar iterasi.
  sleep(5 + Math.random() * 2);
}
