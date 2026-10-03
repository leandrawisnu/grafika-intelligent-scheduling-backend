import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargetsFe, feOpts, loginViaFe } from "../lib/auth-fe.js";
import { apiUrl, getConfig, requireCredentials } from "../lib/config.js";
import { stagesBootstrap, thresholdsReadLight } from "../lib/options.js";

/**
 * E2E bootstrap: matches browser jadwal-context load through Next.js /api/v1 proxy.
 */
export const options = {
  stages: stagesBootstrap,
  thresholds: thresholdsReadLight,
};

export function setup() {
  const cfg = getConfig();
  requireCredentials(cfg);
  const jar = loginViaFe(cfg);
  const targets = discoverTargetsFe(cfg, jar);

  if (!targets.jadwalSemesterId) {
    throw new Error("No jadwal semester found — run backend seed first");
  }

  return { cfg, jar, ...targets };
}

export default function (data) {
  const opts = feOpts(data.jar);
  const jsId = data.jadwalSemesterId;

  const katalog = http.get(apiUrl(data.cfg.baseUrl, "/api/v1/katalog"), {
    ...opts,
    tags: { name: "katalog" },
  });
  check(katalog, { "katalog 200": (r) => r.status === 200 });

  const sesi = http.get(apiUrl(data.cfg.baseUrl, "/api/auth/sesi"), {
    ...opts,
    tags: { name: "auth_sesi" },
  });
  check(sesi, { "sesi 200": (r) => r.status === 200 });

  const list = http.get(apiUrl(data.cfg.baseUrl, "/api/v1/jadwal-semester"), {
    ...opts,
    tags: { name: "jadwal_list" },
  });
  check(list, { "jadwal list 200": (r) => r.status === 200 });

  const batch = http.batch([
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

  for (const res of batch) {
    check(res, { "bootstrap batch 2xx": (r) => r.status >= 200 && r.status < 300 });
  }

  sleep(0.5 + Math.random() * 0.5);
}
