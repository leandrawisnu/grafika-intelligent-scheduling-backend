import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargetsFe, feOpts, feSession, loginViaFe } from "../lib/auth-fe.js";
import { apiUrl, getConfig, requireCredentials } from "../lib/config.js";
import { stagesResolver, thresholdsResolver } from "../lib/options.js";

/**
 * E2E alur Conflict Resolver via Next.js /api/v1 proxy:
 *   list konflik → selesaikan (alternatif AI) → resolusi → terima.
 *
 * Konflik dijamin ada lewat POST demo-konflik (hanya untuk semester
 * bernama "demo"). HATI-HATI: skenario ini memutasi data demo.
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

function muatKonflik(baseUrl, opts, jsId) {
  const res = http.get(
    apiUrl(baseUrl, `/api/v1/jadwal-semester/${jsId}/konflik`),
    { ...opts, tags: { name: "konflik_list" } }
  );
  check(res, { "konflik list 200": (r) => r.status === 200 });
  if (res.status !== 200) return [];
  const rows = res.json();
  return (Array.isArray(rows) ? rows : [])
    .map((k) => k.id)
    .filter(Boolean);
}

function pastikanKonflik(baseUrl, opts, jsId) {
  konflikCache = muatKonflik(baseUrl, opts, jsId);
  if (konflikCache.length === 0) {
    const suntik = http.post(
      apiUrl(baseUrl, `/api/v1/jadwal-semester/${jsId}/demo-konflik`),
      null,
      { ...opts, tags: { name: "demo_konflik" } }
    );
    check(suntik, {
      "demo-konflik 2xx": (r) => r.status >= 200 && r.status < 300,
    });
    konflikCache = muatKonflik(baseUrl, opts, jsId);
  }
  return konflikCache;
}

export function setup() {
  const cfg = getConfig();
  requireCredentials(cfg);
  const session = loginViaFe(cfg);
  const targets = discoverTargetsFe(cfg, session);

  if (!targets.jadwalSemesterId) {
    throw new Error("No jadwal semester found — run backend seed first");
  }

  const opts = feOpts(session);
  http.post(
    apiUrl(cfg.baseUrl, `/api/v1/jadwal-semester/${targets.jadwalSemesterId}/validasi`),
    null,
    { ...opts, tags: { name: "setup_validasi" } }
  );
  const ids = pastikanKonflik(cfg.baseUrl, opts, targets.jadwalSemesterId);
  if (ids.length === 0) {
    throw new Error(
      "No open conflicts. Pastikan semester target bernama 'demo' (untuk demo-konflik) atau datanya punya konflik."
    );
  }

  return { cfg, ...targets };
}

export default function (data) {
  const opts = feOpts(feSession(data.cfg));
  const jsId = data.jadwalSemesterId;

  const ids = pastikanKonflik(data.cfg.baseUrl, opts, jsId);
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
      ...opts,
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

  // 2. Lihat daftar resolusi tersimpan.
  const resolusi = http.get(
    apiUrl(data.cfg.baseUrl, `/api/v1/konflik/${konflikId}/resolusi`),
    { ...opts, tags: { name: "resolver_resolusi" } }
  );
  check(resolusi, {
    "resolusi list 2xx": (r) => r.status >= 200 && r.status < 300,
  });

  // 3. Terima alternatif peringkat pertama (menerapkan perubahan
  //    + deteksi ulang konflik).
  if (alternatif.length > 0 && alternatif[0].id) {
    const terima = http.post(
      apiUrl(data.cfg.baseUrl, `/api/v1/resolusi/${alternatif[0].id}/terima`),
      null,
      {
        ...opts,
        tags: { name: "resolver_terima" },
        expectedStatuses: resolverWriteStatuses,
      }
    );
    check(terima, {
      "terima 2xx atau 409": (r) => resolverOk(r.status),
    });
    konflikCache = [];
  }

  sleep(1 + Math.random());
}
