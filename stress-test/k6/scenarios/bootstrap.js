import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargets, login, sessionHeaders } from "../lib/auth.js";
import { apiUrl, getConfig, requireCredentials } from "../lib/config.js";
import { stagesBootstrap, thresholdsReadLight } from "../lib/options.js";

/**
 * Simulates app bootstrap: catalog + jadwal load (matches frontend jadwal-context).
 */
export const options = {
  stages: stagesBootstrap,
  thresholds: thresholdsReadLight,
};

export function setup() {
  const cfg = getConfig();
  requireCredentials(cfg);
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

  const katalog = http.get(apiUrl(data.cfg.baseUrl, "/api/v1/katalog"), {
    headers,
    tags: { name: "katalog" },
  });
  check(katalog, { "katalog 200": (r) => r.status === 200 });

  const sesi = http.get(apiUrl(data.cfg.baseUrl, "/api/v1/auth/sesi"), {
    headers,
    tags: { name: "auth_sesi" },
  });
  check(sesi, { "sesi 200": (r) => r.status === 200 });

  const list = http.get(apiUrl(data.cfg.baseUrl, "/api/v1/jadwal-semester"), {
    headers,
    tags: { name: "jadwal_list" },
  });
  check(list, { "jadwal list 200": (r) => r.status === 200 });

  const batch = http.batch([
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

  for (const res of batch) {
    check(res, { "bootstrap batch 2xx": (r) => r.status >= 200 && r.status < 300 });
  }

  sleep(0.5 + Math.random() * 0.5);
}
