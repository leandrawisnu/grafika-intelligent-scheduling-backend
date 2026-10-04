/** Per-VU katalog ETag cache (module scope = one per VU). */

let katalogEtag = null;

export function katalogHeaders(baseOpts) {
  const headers = { ...baseOpts.headers };
  if (katalogEtag) {
    headers["If-None-Match"] = katalogEtag;
  }
  return { ...baseOpts, headers };
}

export function rememberKatalogEtag(res) {
  const etag = res.headers.Etag || res.headers.ETag;
  if (etag) {
    katalogEtag = etag;
  }
}

export function resetKatalogEtag() {
  katalogEtag = null;
}
