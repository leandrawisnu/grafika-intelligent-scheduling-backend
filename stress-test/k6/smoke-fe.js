import http from "k6/http";
import { check, sleep } from "k6";
import { runFeBootstrapBatch } from "./lib/bootstrap-smoke.js";
import { discoverTargetsFe, feOpts, feSession, loginViaFe } from "./lib/auth-fe.js";
import { apiUrl, getConfig, requireCredentials } from "./lib/config.js";
import { katalogHeaders } from "./lib/katalog-etag.js";
import { stagesSmoke, thresholdsSmokeFe } from "./lib/options.js";
import { bootstrapSummary } from "./lib/summary.js";

/**
 * E2E smoke: login via Next.js BFF, API via /api/v1 proxy (same-origin cookie).
 * K6_BASE_URL = frontend origin (https://domain or http://127.0.0.1:3000).
 */
export const options = {
  stages: stagesSmoke,
  thresholds: thresholdsSmokeFe,
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
  const kOpts = katalogHeaders(opts);

  runFeBootstrapBatch(data.cfg.baseUrl, data.jadwalSemesterId, opts, kOpts);

  sleep(1);
}

export function handleSummary(data) {
  return bootstrapSummary(data, "fe-smoke");
}
