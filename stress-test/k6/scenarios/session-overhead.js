import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargets, login, sessionHeaders } from "../lib/auth.js";
import { apiUrl, getConfig, requireCredentials } from "../lib/config.js";
import { stagesSession, thresholdsReadLight } from "../lib/options.js";

/**
 * Measures session middleware overhead: every request hits DB sesi lookup.
 */
export const options = {
  stages: stagesSession,
  thresholds: thresholdsReadLight,
};

export function setup() {
  const cfg = getConfig();
  requireCredentials(cfg);
  const token = login(cfg);
  const targets = discoverTargets(cfg, token);
  return { cfg, token, ...targets };
}

export default function (data) {
  const headers = sessionHeaders(data.token);

  const endpoints = [
    "/api/v1/auth/sesi",
    "/api/v1/katalog",
    "/api/v1/jadwal-semester",
  ];

  if (data.jadwalSemesterId) {
    endpoints.push(`/api/v1/jadwal-semester/${data.jadwalSemesterId}/ringkasan`);
  }

  const path = endpoints[Math.floor(Math.random() * endpoints.length)];
  const res = http.get(apiUrl(data.cfg.baseUrl, path), {
    headers,
    tags: { name: "session_read" },
  });

  check(res, { "session read 2xx": (r) => r.status >= 200 && r.status < 300 });
  sleep(0.2 + Math.random() * 0.3);
}
