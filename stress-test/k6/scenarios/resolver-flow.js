import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargets, login, sessionHeaders } from "../lib/auth.js";
import { apiUrl, getConfig, requireCredentials } from "../lib/config.js";
import { stagesResolver, thresholdsResolver } from "../lib/options.js";

/**
 * Alur Conflict Resolver (business process baru):
 *   list konflik terbuka → minta alternatif AI (selesaikan) → lihat resolusi
 *   → terima resolusi (menerapkan perubahan + deteksi ulang konflik).
 *
 * Konflik dijamin ada lewat POST demo-konflik (endpoint khusus demo —
 * hanya berlaku bila nama semester/tahun ajaran mengandung kata "demo").
 * HATI-HATI: skenario ini memutasi data demo (guru slot diacak, guru
 * ditugaskan ulang saat resolusi diterima).
 */
export const options = {
  stages: stagesResolver,
  thresholds: thresholdsResolver,
};

// State per-VU (module scope terisolasi per VU di k6).
let konflikCache = [];
let idxKonflik = 0;

const resolverWriteStatuses = [200, 201, 202, 204, 409];

function resolverOk(status) {
  return (status >= 200 && status < 300) || status === 409;
}

function muatKonflik(baseUrl, headers, jsId) {
  const res = http.get(
    apiUrl(baseUrl, `/api/v1/jadwal-semester/${jsId}/konflik`),
    { headers, tags: { name: "konflik_list" } }
  );
  check(res, { "konflik list 200": (r) => r.status === 200 });
  if (res.status !== 200) return [];
  const rows = res.json();
  return (Array.isArray(rows) ? rows : [])
    .map((k) => k.id)
    .filter(Boolean);
}

function pastikanKonflik(baseUrl, headers, jsId) {
  konflikCache = muatKonflik(baseUrl, headers, jsId);
  if (konflikCache.length === 0) {
    // Suntik konflik acak (guru bentrok, ruangan bentrok, hari libur,
    // kelebihan jam) lalu list ulang. Gagal (bukan semester demo) →
    // tetap coba konflik yang sudah ada di data.
    const suntik = http.post(
      apiUrl(baseUrl, `/api/v1/jadwal-semester/${jsId}/demo-konflik`),
      null,
      { headers, tags: { name: "demo_konflik" } }
    );
    check(suntik, {
      "demo-konflik 2xx": (r) => r.status >= 200 && r.status < 300,
    });
    konflikCache = muatKonflik(baseUrl, headers, jsId);
  }
  return konflikCache;
}

export function setup() {
  const cfg = getConfig();
  requireCredentials(cfg);
  const token = login(cfg);
  const targets = discoverTargets(cfg, token);

  if (!targets.jadwalSemesterId) {
    throw new Error("No jadwal semester found — run backend seed first");
  }

  const headers = sessionHeaders(token);
  // Populasi konflik dari data saat ini, lalu suntik demo bila belum ada.
  http.post(
    apiUrl(cfg.baseUrl, `/api/v1/jadwal-semester/${targets.jadwalSemesterId}/validasi`),
    null,
    { headers, tags: { name: "setup_validasi" } }
  );
  const ids = pastikanKonflik(cfg.baseUrl, headers, targets.jadwalSemesterId);
  if (ids.length === 0) {
    throw new Error(
      "No open conflicts. Pastikan semester target bernama 'demo' (untuk demo-konflik) atau datanya punya konflik."
    );
  }

  return { cfg, token, ...targets };
}

export default function (data) {
  const headers = sessionHeaders(data.token);
  const jsId = data.jadwalSemesterId;

  const ids = pastikanKonflik(data.cfg.baseUrl, headers, jsId);
  if (ids.length === 0) {
    sleep(1);
    return;
  }

  const konflikId = ids[idxKonflik % ids.length];
  idxKonflik++;

  // 1. Minta alternatif perbaikan (deterministik, tanpa ML).
  const selesai = http.post(
    apiUrl(data.cfg.baseUrl, `/api/v1/konflik/${konflikId}/selesaikan`),
    null,
    {
      headers,
      tags: { name: "resolver_selesaikan" },
      expectedStatuses: resolverWriteStatuses,
    }
  );
  check(selesai, {
    "selesaikan 2xx atau 409": (r) => resolverOk(r.status),
  });
  if (selesai.status === 409) {
    konflikCache = [];
    sleep(1 + Math.random());
    return;
  }
  if (!resolverOk(selesai.status)) return;

  let alternatif = [];
  try {
    alternatif = selesai.json("alternatif") || [];
  } catch {
    alternatif = [];
  }

  // 2. Lihat daftar resolusi tersimpan untuk konflik ini.
  const resolusi = http.get(
    apiUrl(data.cfg.baseUrl, `/api/v1/konflik/${konflikId}/resolusi`),
    { headers, tags: { name: "resolver_resolusi" } }
  );
  check(resolusi, {
    "resolusi list 2xx": (r) => r.status >= 200 && r.status < 300,
  });

  // 3. Terima alternatif peringkat pertama — menerapkan perubahan
  //    (reassign guru / ruangan / jam) lalu menjalankan ulang DeteksiKonflik.
  if (alternatif.length > 0 && alternatif[0].id) {
    const terima = http.post(
      apiUrl(data.cfg.baseUrl, `/api/v1/resolusi/${alternatif[0].id}/terima`),
      null,
      {
        headers,
        tags: { name: "resolver_terima" },
        expectedStatuses: resolverWriteStatuses,
      }
    );
    check(terima, {
      "terima 2xx atau 409": (r) => resolverOk(r.status),
    });
    // terima menghapus + menyimpan ulang konflik semester ini → cache usang.
    konflikCache = [];
  }

  sleep(1 + Math.random());
}
