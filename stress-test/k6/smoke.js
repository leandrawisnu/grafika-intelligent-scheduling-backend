import http from "k6/http";
import { check, sleep } from "k6";
import { runApiBootstrapBatch } from "./lib/bootstrap-smoke.js";
import { discoverTargets, login, sessionHeaders } from "./lib/auth.js";
import { apiUrl, getConfig, requireCredentials } from "./lib/config.js";
import { katalogHeaders } from "./lib/katalog-etag.js";
import { stagesSmoke, thresholdsSmokeBootstrap } from "./lib/options.js";
import { bootstrapSummary } from "./lib/summary.js";

export const options = {
  stages: stagesSmoke,
  thresholds: thresholdsSmokeBootstrap,
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
  const kHdr = katalogHeaders({ headers }).headers;

  runApiBootstrapBatch(data.cfg.baseUrl, data.jadwalSemesterId, headers, kHdr);

  sleep(1);
}

export function handleSummary(data) {
  return bootstrapSummary(data, "smoke-api");
}
