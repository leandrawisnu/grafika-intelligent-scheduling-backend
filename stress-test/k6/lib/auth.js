import http from "k6/http";
import { check } from "k6";
import { apiUrl } from "./config.js";

export function sessionHeaders(token) {
  return {
    "Content-Type": "application/json",
    "X-GIS-Session": token,
  };
}

export function login(cfg) {
  const res = http.post(
    apiUrl(cfg.baseUrl, "/api/v1/auth/masuk"),
    JSON.stringify({ email: cfg.email, password: cfg.password }),
    {
      headers: { "Content-Type": "application/json" },
      tags: { name: "auth_masuk" },
    }
  );

  check(res, {
    "login status 200": (r) => r.status === 200,
    "login has token": (r) => Boolean(r.json("token")),
  });

  if (res.status !== 200) {
    throw new Error(`Login failed (${res.status}): ${res.body}`);
  }

  return res.json("token");
}

/**
 * Resolve jadwal semester / kelas IDs for scenarios (env override or API discovery).
 */
export function discoverTargets(cfg, token) {
  const headers = sessionHeaders(token);
  let jadwalSemesterId = cfg.jadwalSemesterId;
  let jadwalKelasId = cfg.jadwalKelasId;

  if (!jadwalSemesterId) {
    const listRes = http.get(apiUrl(cfg.baseUrl, "/api/v1/jadwal-semester"), {
      headers,
      tags: { name: "setup_jadwal_list" },
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
      { headers, tags: { name: "setup_jk_list" } }
    );

    if (jkRes.status === 200) {
      const jkList = jkRes.json();
      const rows = Array.isArray(jkList) ? jkList : [];
      jadwalKelasId = rows[0]?.id || "";
    }
  }

  return { jadwalSemesterId, jadwalKelasId };
}
