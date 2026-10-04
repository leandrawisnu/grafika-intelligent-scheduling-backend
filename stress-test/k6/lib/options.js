/** Shared load profiles and SLO thresholds for GIS API tests. */

export const stagesSmoke = [
  { duration: "30s", target: 1 },
];

export const stagesBootstrap = [
  { duration: "30s", target: 5 },
  { duration: "2m", target: 20 },
  { duration: "1m", target: 50 },
  { duration: "30s", target: 0 },
];

export const stagesGridRead = [
  { duration: "30s", target: 3 },
  { duration: "3m", target: 10 },
  { duration: "30s", target: 0 },
];

export const stagesSession = [
  { duration: "30s", target: 10 },
  { duration: "3m", target: 40 },
  { duration: "30s", target: 0 },
];

export const stagesValidation = [
  { duration: "10s", target: 1 },
  { duration: "2m", target: 3 },
  { duration: "10s", target: 0 },
];

export const stagesResolver = [
  { duration: "30s", target: 1 },
  { duration: "2m", target: 3 },
  { duration: "30s", target: 0 },
];

export const stagesPlotting = [
  { duration: "30s", target: 3 },
  { duration: "2m", target: 10 },
  { duration: "30s", target: 0 },
];

// Rollover = operasi admin sekali per semester → benchmark 1 VU.
export const stagesRollover = [{ duration: "1m", target: 1 }];

/** Tags used by smoke / fe-smoke bootstrap batch. */
export const smokeBootstrapTags = [
  "katalog",
  "auth_sesi",
  "jadwal_list",
  "jadwal_detail",
  "jk_aktif_ringkas",
  "ringkasan",
];

export const thresholdsSmoke = {
  http_req_failed: ["rate<0.05"],
  http_req_duration: ["p(95)<2000"],
};

/** API-direct smoke: per-endpoint latency (abortOnFail false = report only). */
export const thresholdsSmokeBootstrap = {
  http_req_failed: ["rate<0.05"],
  http_req_duration: ["p(95)<2000"],
  ...perEndpointThresholds(smokeBootstrapTags, 2000),
};

/**
 * FE proxy smoke: looser overall SLO; per-endpoint thresholds for diagnosis.
 */
export const thresholdsSmokeFe = {
  http_req_failed: ["rate<0.05"],
  http_req_duration: ["p(95)<2000"],
  ...perEndpointThresholds(smokeBootstrapTags, 2000),
};

function perEndpointThresholds(tags, p95Ms, overrides = {}) {
  const out = {};
  for (const name of tags) {
    const limit = overrides[name] ?? p95Ms;
    out[`http_req_duration{name:${name}}`] = [
      { threshold: `p(95)<${limit}`, abortOnFail: false },
    ];
    out[`response_size_bytes{name:${name}}`] = [
      { threshold: "avg<5000000", abortOnFail: false },
    ];
  }
  return out;
}

export const thresholdsReadLight = {
  http_req_failed: ["rate<0.01"],
  http_req_duration: ["p(95)<500", "p(99)<1500"],
};

export const thresholdsReadHeavy = {
  http_req_failed: ["rate<0.02"],
  http_req_duration: ["p(95)<2000", "p(99)<5000"],
};

export const thresholdsValidation = {
  http_req_failed: ["rate<0.05"],
  http_req_duration: ["p(95)<10000"],
  ...perEndpointThresholds(
    ["validasi", "validasi_rearm"],
    10000,
    { validasi_rearm: 2000 }
  ),
};

export const thresholdsResolver = {
  http_req_failed: ["rate<0.05"],
  http_req_duration: ["p(95)<2000"],
  ...perEndpointThresholds(
    ["konflik_list", "resolver_selesaikan", "resolver_resolusi", "resolver_terima"],
    3000
  ),
};

export const thresholdsPlotting = {
  http_req_failed: ["rate<0.02"],
  http_req_duration: ["p(95)<1000"],
  ...perEndpointThresholds(
    ["plotting_list", "plotting_create", "plotting_update", "plotting_delete"],
    1500
  ),
};

export const thresholdsRollover = {
  http_req_failed: ["rate<0.01"],
  ...perEndpointThresholds(["rollover_salin"], 15000),
};
