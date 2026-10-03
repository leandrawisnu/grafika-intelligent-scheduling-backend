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

export const thresholdsSmoke = {
  http_req_failed: ["rate<0.05"],
  http_req_duration: ["p(95)<2000"],
};

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
  "http_req_duration{name:validasi}": ["p(95)<10000"],
};
