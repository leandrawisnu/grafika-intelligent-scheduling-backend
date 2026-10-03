/**
 * Shared k6 config from environment variables.
 * Copy .env.example → .env and source before run (see scripts/run.sh).
 */

export function getConfig() {
  const baseUrl = (__ENV.K6_BASE_URL || "http://127.0.0.1:6060").replace(/\/$/, "");

  return {
    baseUrl,
    email: __ENV.K6_USER || "",
    password: __ENV.K6_PASSWORD || "",
    jadwalSemesterId: __ENV.K6_JADWAL_SEMESTER_ID || "",
    jadwalKelasId: __ENV.K6_JADWAL_KELAS_ID || "",
  };
}

export function apiUrl(baseUrl, path) {
  const normalized = path.startsWith("/") ? path : `/${path}`;
  return `${baseUrl}${normalized}`;
}

export function requireCredentials(cfg) {
  if (!cfg.email || !cfg.password) {
    throw new Error("Set K6_USER and K6_PASSWORD (see stress-test/.env.example)");
  }
}
