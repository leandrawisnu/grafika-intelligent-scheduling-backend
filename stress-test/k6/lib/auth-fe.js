import http from "k6/http";
import { check } from "k6";
import { apiUrl } from "./config.js";

const SESSION_COOKIE_NAMES = ["__Host-gis_session", "gis_session"];

/**
 * Per-VU session cache (module scope is isolated per VU in k6).
 * Cookie jars / setup() handoff are unreliable for authenticated iterations.
 */
let vuFeSession = null;

function cookieHeaderFromResponse(res) {
  for (const name of SESSION_COOKIE_NAMES) {
    const entry = res.cookies[name];
    if (entry && entry.length > 0 && entry[0].value) {
      return `${name}=${entry[0].value}`;
    }
  }
  return "";
}

function cookieHeaderFromJar(jar, baseUrl) {
  const urls = [baseUrl, `${baseUrl}/`, apiUrl(baseUrl, "/api/auth/login")];
  for (const url of urls) {
    const stored = jar.cookiesForURL(url);
    for (const name of SESSION_COOKIE_NAMES) {
      const entry = stored[name];
      if (entry && entry.length > 0 && entry[0].value) {
        return `${name}=${entry[0].value}`;
      }
    }
  }
  return "";
}

/**
 * Browser-like auth: POST /api/auth/login → session cookie → /api/v1/* via Next.js proxy.
 */
export function loginViaFe(cfg) {
  const jar = http.cookieJar();
  const res = http.post(
    apiUrl(cfg.baseUrl, "/api/auth/login"),
    JSON.stringify({ email: cfg.email, password: cfg.password }),
    {
      headers: { "Content-Type": "application/json" },
      tags: { name: "fe_auth_login" },
      jar,
    }
  );

  check(res, {
    "fe login 200": (r) => r.status === 200,
    "fe login has pengguna": (r) => Boolean(r.json("pengguna")),
  });

  if (res.status !== 200) {
    throw new Error(`FE login failed (${res.status}): ${res.body}`);
  }

  const cookieHeader =
    cookieHeaderFromResponse(res) || cookieHeaderFromJar(jar, cfg.baseUrl);
  if (!cookieHeader) {
    throw new Error(
      "FE login: no session cookie in response (expected __Host-gis_session or gis_session)"
    );
  }

  return { cookieHeader };
}

/** One login per VU — reuse for all iterations in that VU. */
export function feSession(cfg) {
  if (!vuFeSession) {
    vuFeSession = loginViaFe(cfg);
  }
  return vuFeSession;
}

export function feOpts(session) {
  return {
    headers: {
      "Content-Type": "application/json",
      Cookie: session.cookieHeader,
    },
  };
}

/**
 * Resolve jadwal semester / kelas IDs through FE proxy (session cookie).
 */
export function discoverTargetsFe(cfg, session) {
  const opts = feOpts(session);
  let jadwalSemesterId = cfg.jadwalSemesterId;
  let jadwalKelasId = cfg.jadwalKelasId;

  if (!jadwalSemesterId) {
    const listRes = http.get(apiUrl(cfg.baseUrl, "/api/v1/jadwal-semester"), {
      ...opts,
      tags: { name: "fe_setup_jadwal_list" },
    });

    if (listRes.status !== 200) {
      throw new Error(
        `FE discover jadwal-semester failed (${listRes.status}): ${listRes.body}`
      );
    }

    const list = listRes.json();
    const rows = Array.isArray(list) ? list : [];
    const pick = rows.find((j) => j.punya_kelas_aktif) || rows[0];
    jadwalSemesterId = pick?.id || "";
  }

  if (jadwalSemesterId && !jadwalKelasId) {
    const jkRes = http.get(
      apiUrl(
        cfg.baseUrl,
        `/api/v1/jadwal-semester/${jadwalSemesterId}/jadwal-kelas-aktif?ringkas=1`
      ),
      { ...opts, tags: { name: "fe_setup_jk_list" } }
    );

    if (jkRes.status === 200) {
      const jkList = jkRes.json();
      const rows = Array.isArray(jkList) ? jkList : [];
      jadwalKelasId = rows[0]?.id || "";
    }
  }

  return { jadwalSemesterId, jadwalKelasId };
}
