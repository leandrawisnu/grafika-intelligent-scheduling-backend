import { Trend } from "k6/metrics";

export const responseSizeBytes = new Trend("response_size_bytes", true);

export function recordResponseSize(res, tagName) {
  const bytes = res.body ? res.body.length : 0;
  responseSizeBytes.add(bytes, { name: tagName });
  return bytes;
}
