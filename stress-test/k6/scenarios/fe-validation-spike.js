import http from "k6/http";
import { check, sleep } from "k6";
import { discoverTargetsFe, feOpts, loginViaFe } from "../lib/auth-fe.js";
import { apiUrl, getConfig, requireCredentials } from "../lib/config.js";
import { stagesValidation, thresholdsValidation } from "../lib/options.js";

/**
 * E2E POST validasi via Next.js /api/v1 proxy.
 */
export const options = {
  stages: stagesValidation,
  thresholds: thresholdsValidation,
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

  const res = http.post(
    apiUrl(data.cfg.baseUrl, `/api/v1/jadwal-semester/${jsId}/validasi`),
    null,
    { ...opts, tags: { name: "validasi" } }
  );

  check(res, {
    "validasi 2xx": (r) => r.status >= 200 && r.status < 300,
  });

  sleep(3 + Math.random() * 2);
}
