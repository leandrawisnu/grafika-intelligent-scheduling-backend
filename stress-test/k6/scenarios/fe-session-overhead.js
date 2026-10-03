import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargetsFe, feOpts, loginViaFe } from "../lib/auth-fe.js";
import { apiUrl, getConfig, requireCredentials } from "../lib/config.js";
import { stagesSession, thresholdsReadLight } from "../lib/options.js";

/**
 * E2E session overhead: cookie + Next.js middleware + proxy + Go session lookup.
 */
export const options = {
  stages: stagesSession,
  thresholds: thresholdsReadLight,
};

export function setup() {
  const cfg = getConfig();
  requireCredentials(cfg);
  const session = loginViaFe(cfg);
  const targets = discoverTargetsFe(cfg, session);
  return { cfg, session, ...targets };
}

export default function (data) {
  const opts = feOpts(data.session);

  const endpoints = [
    "/api/auth/sesi",
    "/api/v1/katalog",
    "/api/v1/jadwal-semester",
  ];

  if (data.jadwalSemesterId) {
    endpoints.push(`/api/v1/jadwal-semester/${data.jadwalSemesterId}/ringkasan`);
  }

  const path = endpoints[Math.floor(Math.random() * endpoints.length)];
  const res = http.get(apiUrl(data.cfg.baseUrl, path), {
    ...opts,
    tags: { name: "fe_session_read" },
  });

  check(res, { "fe session read 2xx": (r) => r.status >= 200 && r.status < 300 });
  sleep(0.2 + Math.random() * 0.3);
}
