import http from "k6/http";
import { check } from "k6";
import { apiUrl } from "./config.js";

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

  return jar;
}

export function feOpts(jar) {
  return {
    jar,
    headers: { "Content-Type": "application/json" },
  };
}

/**
 * Resolve jadwal semester / kelas IDs through FE proxy (cookie jar).
 */
export function discoverTargetsFe(cfg, jar) {
  const opts = feOpts(jar);
  let jadwalSemesterId = cfg.jadwalSemesterId;
  let jadwalKelasId = cfg.jadwalKelasId;

  if (!jadwalSemesterId) {
    const listRes = http.get(apiUrl(cfg.baseUrl, "/api/v1/jadwal-semester"), {
      ...opts,
      tags: { name: "fe_setup_jadwal_list" },
    });

    if (listRes.status === 200) {
      const list = listRes.json();
      const rows = Array.isArray(list) ? list : [];
      const pick = rows.find((j) => j.punya_kelas_aktif) || rows[0];
      jadwalSemesterId = pick?.id || "";
    }
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
