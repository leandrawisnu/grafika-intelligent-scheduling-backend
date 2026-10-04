import http from "k6/http";
import { check } from "k6";
import { apiUrl } from "./config.js";
import { rememberKatalogEtag } from "./katalog-etag.js";
import { recordResponseSize } from "./smoke-metrics.js";

export function runKatalog(baseUrl, requestOpts) {
  const res = http.get(apiUrl(baseUrl, "/api/v1/katalog"), {
    ...requestOpts,
    tags: { name: "katalog" },
  });
  recordResponseSize(res, "katalog");
  check(res, {
    "katalog status 2xx/304": (r) =>
      (r.status >= 200 && r.status < 300) || r.status === 304,
  });
  rememberKatalogEtag(res);
  return res;
}

/**
 * Shared bootstrap batch for smoke / fe-smoke (katalog via runKatalog separately).
 */
export function runBootstrapBatch(baseUrl, jsId, requestOpts) {
  const tagged = (name, extra = {}) => ({ ...requestOpts, tags: { name, ...extra } });

  const responses = http.batch([
    ["GET", apiUrl(baseUrl, "/api/v1/jadwal-semester"), null, tagged("jadwal_list")],
    [
      "GET",
      apiUrl(baseUrl, `/api/v1/jadwal-semester/${jsId}`),
      null,
      tagged("jadwal_detail"),
    ],
    [
      "GET",
      apiUrl(baseUrl, `/api/v1/jadwal-semester/${jsId}/jadwal-kelas-aktif?ringkas=1`),
      null,
      tagged("jk_aktif_ringkas"),
    ],
    [
      "GET",
      apiUrl(baseUrl, `/api/v1/jadwal-semester/${jsId}/ringkasan`),
      null,
      tagged("ringkasan"),
    ],
  ]);

  const names = ["jadwal_list", "jadwal_detail", "jk_aktif_ringkas", "ringkasan"];

  for (let i = 0; i < responses.length; i++) {
    const res = responses[i];
    const name = names[i];
    recordResponseSize(res, name);
    check(res, {
      [`${name} status 2xx`]: (r) => r.status >= 200 && r.status < 300,
    });
  }

  return responses;
}

/** FE-only: auth/sesi BFF + bootstrap API batch. */
export function runFeBootstrapBatch(baseUrl, jsId, requestOpts, katalogOpts) {
  runKatalog(baseUrl, katalogOpts || requestOpts);

  const sesi = http.get(apiUrl(baseUrl, "/api/auth/sesi"), {
    ...requestOpts,
    tags: { name: "auth_sesi" },
  });
  recordResponseSize(sesi, "auth_sesi");
  check(sesi, { "auth_sesi status 2xx": (r) => r.status >= 200 && r.status < 300 });

  return runBootstrapBatch(baseUrl, jsId, requestOpts);
}

/** API-direct: auth/sesi + bootstrap batch. */
export function runApiBootstrapBatch(baseUrl, jsId, headers, katalogHeaders) {
  runKatalog(baseUrl, { headers: katalogHeaders || headers });

  const sesi = http.get(apiUrl(baseUrl, "/api/v1/auth/sesi"), {
    headers,
    tags: { name: "auth_sesi" },
  });
  recordResponseSize(sesi, "auth_sesi");
  check(sesi, { "auth_sesi status 2xx": (r) => r.status >= 200 && r.status < 300 });

  return runBootstrapBatch(baseUrl, jsId, { headers });
}
