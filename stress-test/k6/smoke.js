import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargets, login, sessionHeaders } from "./lib/auth.js";
import { apiUrl, getConfig, requireCredentials } from "./lib/config.js";
import { stagesSmoke, thresholdsSmoke } from "./lib/options.js";

export const options = {
  stages: stagesSmoke,
  thresholds: thresholdsSmoke,
};

export function setup() {
  const cfg = getConfig();
  requireCredentials(cfg);

  const health = http.get(apiUrl(cfg.baseUrl, "/health"));
  check(health, { "health ok": (r) => r.status === 200 });

  const token = login(cfg);
  const targets = discoverTargets(cfg, token);

  if (!targets.jadwalSemesterId) {
    throw new Error("No jadwal semester found — run backend seed first");
  }

  return { cfg, token, ...targets };
}

export default function (data) {
  const headers = sessionHeaders(data.token);
  const jsId = data.jadwalSemesterId;

  const responses = http.batch([
    ["GET", apiUrl(data.cfg.baseUrl, "/api/v1/katalog"), null, { headers, tags: { name: "katalog" } }],
    ["GET", apiUrl(data.cfg.baseUrl, "/api/v1/auth/sesi"), null, { headers, tags: { name: "auth_sesi" } }],
    ["GET", apiUrl(data.cfg.baseUrl, "/api/v1/jadwal-semester"), null, { headers, tags: { name: "jadwal_list" } }],
    [
      "GET",
      apiUrl(data.cfg.baseUrl, `/api/v1/jadwal-semester/${jsId}`),
      null,
      { headers, tags: { name: "jadwal_detail" } },
    ],
    [
      "GET",
      apiUrl(data.cfg.baseUrl, `/api/v1/jadwal-semester/${jsId}/jadwal-kelas-aktif?ringkas=1`),
      null,
      { headers, tags: { name: "jk_aktif_ringkas" } },
    ],
    [
      "GET",
      apiUrl(data.cfg.baseUrl, `/api/v1/jadwal-semester/${jsId}/ringkasan`),
      null,
      { headers, tags: { name: "ringkasan" } },
    ],
    [
      "GET",
      apiUrl(data.cfg.baseUrl, `/api/v1/jadwal-semester/${jsId}/konflik`),
      null,
      { headers, tags: { name: "konflik" } },
    ],
  ]);

  for (const res of responses) {
    check(res, { "status 2xx": (r) => r.status >= 200 && r.status < 300 });
  }

  sleep(1);
}
