import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargets, login, sessionHeaders } from "../lib/auth.js";
import { apiUrl, getConfig, requireCredentials } from "../lib/config.js";
import { stagesValidation, thresholdsValidation } from "../lib/options.js";

/**
 * CPU + DB heavy: POST validasi (full conflict detection on all active slots).
 *
 * Setiap iterasi diawali PUT slot no-op untuk mengibar ulang perlu_validasi.
 * Tanpa re-arm, iterasi kedua dan seterusnya hanya membaca cache konflik
 * (perlu_validasi=false setelah DeteksiKonflik pertama) — sehingga yang
 * terukur bukan deteksi konflik, melainkan SELECT murah.
 *
 * Low VU by design — avoid running on production data without backup.
 */
export const options = {
  stages: stagesValidation,
  thresholds: thresholdsValidation,
};

export function setup() {
  const cfg = getConfig();
  requireCredentials(cfg);
  const token = login(cfg);
  const targets = discoverTargets(cfg, token);

  if (!targets.jadwalSemesterId) {
    throw new Error("No jadwal semester found — run backend seed first");
  }
  if (!targets.jadwalKelasId) {
    throw new Error("No jadwal kelas found — run make seed-demo first");
  }

  // Ambil satu slot untuk re-arm (PUT no-op mengibar ulang perlu_validasi
  // lewat LayananJadwal.PerbaruiSlot → TandaiPerluValidasi).
  const headers = sessionHeaders(token);
  const jkRes = http.get(
    apiUrl(cfg.baseUrl, `/api/v1/jadwal-kelas/${targets.jadwalKelasId}`),
    { headers, tags: { name: "setup_jk_detail" } }
  );
  if (jkRes.status !== 200) {
    throw new Error(`Gagal muat jadwal kelas (${jkRes.status}): ${jkRes.body}`);
  }
  const slots = jkRes.json("slot_jadwal") || [];
  const slot = (Array.isArray(slots) ? slots : []).find((s) => s.id);
  if (!slot) {
    throw new Error("Jadwal kelas belum punya slot — run make seed-demo first");
  }

  return {
    cfg,
    token,
    ...targets,
    slotId: slot.id,
    mingguKe: slot.minggu_ke,
  };
}

export default function (data) {
  const headers = sessionHeaders(data.token);
  const jsId = data.jadwalSemesterId;

  // Re-arm: tulis ulang nilai slot yang sama (no-op data, flag perlu_validasi
  // kembali true) agar validasi menjalankan DeteksiKonflik penuh.
  const rearm = http.put(
    apiUrl(data.cfg.baseUrl, `/api/v1/slot/${data.slotId}`),
    JSON.stringify({ minggu_ke: data.mingguKe }),
    { headers, tags: { name: "validasi_rearm" } }
  );
  check(rearm, { "rearm slot 2xx": (r) => r.status >= 200 && r.status < 300 });

  const res = http.post(
    apiUrl(data.cfg.baseUrl, `/api/v1/jadwal-semester/${jsId}/validasi`),
    null,
    { headers, tags: { name: "validasi" } }
  );

  check(res, {
    "validasi 2xx": (r) => r.status >= 200 && r.status < 300,
  });

  sleep(3 + Math.random() * 2);
}
