import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargetsFe, feOpts, loginViaFe } from "../lib/auth-fe.js";
import { apiUrl, getConfig, requireCredentials } from "../lib/config.js";
import { stagesGridRead, thresholdsReadHeavy } from "../lib/options.js";

/**
 * E2E heavy read: jadwal kelas grid via Next.js /api/v1 proxy.
 */
export const options = {
  stages: stagesGridRead,
  thresholds: thresholdsReadHeavy,
};

export function setup() {
  const cfg = getConfig();
  requireCredentials(cfg);
  const jar = loginViaFe(cfg);
  const targets = discoverTargetsFe(cfg, jar);

  if (!targets.jadwalKelasId) {
    throw new Error("No jadwal kelas found — run make seed-slots first");
  }

  return { cfg, jar, ...targets };
}

export default function (data) {
  const opts = feOpts(data.jar);
  const jkId = data.jadwalKelasId;

  const grid = http.get(apiUrl(data.cfg.baseUrl, `/api/v1/jadwal-kelas/${jkId}`), {
    ...opts,
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
      { ...opts, tags: { name: "jk_aktif_full" } }
    );
    check(full, { "jk aktif full 200": (r) => r.status === 200 });
  }

  sleep(1 + Math.random());
}
