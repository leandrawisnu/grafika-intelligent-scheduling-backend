import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargets, login, sessionHeaders } from "../lib/auth.js";
import { apiUrl, getConfig, requireCredentials } from "../lib/config.js";
import { stagesGridRead, thresholdsReadHeavy } from "../lib/options.js";

/**
 * Heavy read: full jadwal kelas grid with deep GORM preloads.
 */
export const options = {
  stages: stagesGridRead,
  thresholds: thresholdsReadHeavy,
};

export function setup() {
  const cfg = getConfig();
  requireCredentials(cfg);
  const token = login(cfg);
  const targets = discoverTargets(cfg, token);

  if (!targets.jadwalKelasId) {
    throw new Error("No jadwal kelas found — run make seed-demo first");
  }

  return { cfg, token, ...targets };
}

export default function (data) {
  const headers = sessionHeaders(data.token);
  const jkId = data.jadwalKelasId;

  const grid = http.get(apiUrl(data.cfg.baseUrl, `/api/v1/jadwal-kelas/${jkId}`), {
    headers,
    tags: { name: "jadwal_kelas_grid" },
  });
  check(grid, {
    "grid 200": (r) => r.status === 200,
    "grid has body": (r) => r.body && r.body.length > 100,
  });

  if (data.jadwalSemesterId) {
    const full = http.get(
      apiUrl(
        data.cfg.baseUrl,
        `/api/v1/jadwal-semester/${data.jadwalSemesterId}/jadwal-kelas-aktif`
      ),
      { headers, tags: { name: "jk_aktif_full" } }
    );
    check(full, { "jk aktif full 200": (r) => r.status === 200 });
  }

  sleep(1 + Math.random());
}
