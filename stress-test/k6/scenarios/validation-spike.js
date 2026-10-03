import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargets, login, sessionHeaders } from "../lib/auth.js";
import { apiUrl, getConfig, requireCredentials } from "../lib/config.js";
import { stagesValidation, thresholdsValidation } from "../lib/options.js";

/**
 * CPU + DB heavy: POST validasi (conflict detection on all active slots).
 * Low VU by design — avoid running on production data without backup.
 */
export const options = {
  stages: stagesValidation,
  thresholds: thresholdsValidation,
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

  const res = http.post(
    apiUrl(data.cfg.baseUrl, `/api/v1/jadwal-semester/${jsId}/validasi`),
    null,
    { headers, tags: { name: "validasi" } }
  );

  check(res, {
    "validasi 2xx": (r) => r.status >= 200 && r.status < 300,
  });

  sleep(3 + Math.random() * 2);
}
