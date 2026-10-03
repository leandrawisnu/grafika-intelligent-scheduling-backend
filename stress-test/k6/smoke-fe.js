import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargetsFe, feOpts, feSession, loginViaFe } from "./lib/auth-fe.js";
import { apiUrl, getConfig, requireCredentials } from "./lib/config.js";
import { stagesSmoke, thresholdsSmoke } from "./lib/options.js";

/**
 * E2E smoke: login via Next.js BFF, API via /api/v1 proxy (same-origin cookie).
 * K6_BASE_URL = frontend origin (https://domain or http://127.0.0.1:3000).
 */
export const options = {
  stages: stagesSmoke,
  thresholds: thresholdsSmoke,
};

export function setup() {
  const cfg = getConfig();
  requireCredentials(cfg);

  const loginPage = http.get(apiUrl(cfg.baseUrl, "/login"));
  check(loginPage, { "login page ok": (r) => r.status === 200 });

  const session = loginViaFe(cfg);
  const targets = discoverTargetsFe(cfg, session);

  if (!targets.jadwalSemesterId) {
    throw new Error("No jadwal semester found — run backend seed first");
  }

  return { cfg, ...targets };
}

export default function (data) {
  const opts = feOpts(feSession(data.cfg));
  const jsId = data.jadwalSemesterId;

  const responses = http.batch([
    ["GET", apiUrl(data.cfg.baseUrl, "/api/v1/katalog"), null, { ...opts, tags: { name: "katalog" } }],
    ["GET", apiUrl(data.cfg.baseUrl, "/api/auth/sesi"), null, { ...opts, tags: { name: "auth_sesi" } }],
    ["GET", apiUrl(data.cfg.baseUrl, "/api/v1/jadwal-semester"), null, { ...opts, tags: { name: "jadwal_list" } }],
    [
      "GET",
      apiUrl(data.cfg.baseUrl, `/api/v1/jadwal-semester/${jsId}`),
      null,
      { ...opts, tags: { name: "jadwal_detail" } },
    ],
    [
      "GET",
      apiUrl(data.cfg.baseUrl, `/api/v1/jadwal-semester/${jsId}/jadwal-kelas-aktif?ringkas=1`),
      null,
      { ...opts, tags: { name: "jk_aktif_ringkas" } },
    ],
    [
      "GET",
      apiUrl(data.cfg.baseUrl, `/api/v1/jadwal-semester/${jsId}/ringkasan`),
      null,
      { ...opts, tags: { name: "ringkasan" } },
    ],
    [
      "GET",
      apiUrl(data.cfg.baseUrl, `/api/v1/jadwal-semester/${jsId}/konflik`),
      null,
      { ...opts, tags: { name: "konflik" } },
    ],
  ]);

  for (const res of responses) {
    check(res, {
      "status 2xx": (r) => r.status >= 200 && r.status < 300,
      "not auth error": (r) =>
        r.status !== 401 || !String(r.body).includes("Sesi tidak berlaku"),
    });
  }

  sleep(1);
}
