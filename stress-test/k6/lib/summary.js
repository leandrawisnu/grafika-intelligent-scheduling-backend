import { smokeBootstrapTags } from "./options.js";

/**
 * k6 handleSummary — export per-endpoint p95 and avg response size.
 */
export function bootstrapSummary(data, label) {
  const endpoints = {};

  for (const name of smokeBootstrapTags) {
    const durKey = `http_req_duration{name:${name}}`;
    const sizeKey = `response_size_bytes{name:${name}}`;
    const dur = data.metrics[durKey]?.values;
    const size = data.metrics[sizeKey]?.values;

    endpoints[name] = {
      p95_ms: dur?.["p(95)"] ?? null,
      avg_ms: dur?.avg ?? null,
      avg_bytes: size?.avg ?? null,
    };
  }

  const overall = data.metrics.http_req_duration?.values;
  const payload = {
    label,
    overall_p95_ms: overall?.["p(95)"] ?? null,
    overall_avg_ms: overall?.avg ?? null,
    http_req_failed_rate: data.metrics.http_req_failed?.values?.rate ?? null,
    endpoints,
  };

  return {
    [`results/${label}-summary.json`]: JSON.stringify(payload, null, 2),
  };
}
