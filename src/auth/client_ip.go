package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ClientIP returns the originating client IP, proxy-aware.
//
// Behind a single trusted reverse proxy (Caddy), the last
// entry of X-Forwarded-For is the one appended by the proxy
// and reflects the real peer. A client can only prepend
// entries, so the last entry cannot be spoofed — provided
// the backend is reachable only through the proxy.
//
// Falls back to the direct remote address when the header
// is absent (local dev, direct access).
func ClientIP(c *fiber.Ctx) string {
	if xff := c.Get("X-Forwarded-For"); xff != "" {
		entries := strings.Split(xff, ",")
		if last := strings.TrimSpace(entries[len(entries)-1]); last != "" {
			return last
		}
	}
	return c.IP()
}
