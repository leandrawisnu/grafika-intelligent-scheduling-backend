import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargetsFe, feOpts, feSession, loginViaFe } from "../lib/auth-fe.js";
import { apiUrl, getConfig, requireCredentials } from "../lib/config.js";
import { stagesValidation, thresholdsValidation } from "../lib/options.js";

/**
 * E2E POST validasi via Next.js /api/v1 proxy.
 *
 * Setiap iterasi diawali PUT slot no-op untuk mengibar ulang
 * perlu_validasi — tanpa itu iterasi kedua+ hanya membaca cache.
 */
export const options = {
  stages: stagesValidation,
  thresholds: thresholdsValidation,
};

export function setup() {
  const cfg = getConfig();
  requireCredentials(cfg);
  const session = loginViaFe(cfg);
  const targets = discoverTargetsFe(cfg, session);

  if (!targets.jadwalSemesterId) {
    throw new Error("No jadwal semester found — run backend seed first");
  }
  if (!targets.jadwalKelasId) {
    throw new Error("No jadwal kelas found — run make seed-demo first");
  }

  // Ambil satu slot untuk re-arm (PUT no-op mengibar ulang perlu_validasi).
  const opts = feOpts(session);
  const jkRes = http.get(
    apiUrl(cfg.baseUrl, `/api/v1/jadwal-kelas/${targets.jadwalKelasId}`),
    { ...opts, tags: { name: "setup_jk_detail" } }
  );
  if (jkRes.status !== 200) {
    throw new Error(`Gagal muat jadwal kelas (${jkRes.status}): ${jkRes.body}`);
  }
  const slots = jkRes.json("slot_jadwal") || [];
  const slot = (Array.isArray(slots) ? slots : []).find((s) => s.id);
  if (!slot) {
    throw new Error("Jadwal kelas belum punya slot — run make seed-demo first");
  }

  return { cfg, ...targets, slotId: slot.id, mingguKe: slot.minggu_ke };
}

export default function (data) {
  const opts = feOpts(feSession(data.cfg));
  const jsId = data.jadwalSemesterId;

  // Re-arm: tulis ulang nilai slot yang sama (no-op data, flag
  // perlu_validasi kembali true) agar validasi menjalankan DeteksiKonflik penuh.
  const rearm = http.put(
    apiUrl(data.cfg.baseUrl, `/api/v1/slot/${data.slotId}`),
    JSON.stringify({ minggu_ke: data.mingguKe }),
    { ...opts, tags: { name: "validasi_rearm" } }
  );
  check(rearm, { "rearm slot 2xx": (r) => r.status >= 200 && r.status < 300 });

  const res = http.post(
    apiUrl(data.cfg.baseUrl, `/api/v1/jadwal-semester/${jsId}/validasi`),
    null,
    { ...opts, tags: { name: "validasi" } }
  );

  check(res, {
    "validasi 2xx": (r) => r.status >= 200 && r.status < 300,
  });

  sleep(3 + Math.random() * 2);
}
