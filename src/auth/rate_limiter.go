package auth

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/src/models"
)

// RateLimiter is an in-memory fixed-window limiter for
// authenticated requests on heavy endpoints. Requests are
// keyed per user (UUID) when a session is present, falling
// back to per client IP.
//
// It is per-process state: it resets on restart and is not
// shared across replicas.
type RateLimiter struct {
	limiter *Penghitung
}

// NewRateLimiter creates a limiter. limit <= 0 (or window <= 0)
// disables it: NewRateLimiter returns nil and its Handler
// passes every request through.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	if limit <= 0 || window <= 0 {
		return nil
	}
	return &RateLimiter{limiter: BaruPenghitung(limit, window)}
}

// Handler returns a Fiber middleware that rejects requests
// over the limit with 429 and a Retry-After header.
// It must run after WajibSesi so the user UUID is
// available in locals.
func (r *RateLimiter) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if r == nil {
			return c.Next()
		}
		key := "ip:" + ClientIP(c)
		if akun, ok := c.Locals(LocalPengguna).(models.Pengguna); ok && akun.ID != uuid.Nil {
			key = "user:" + akun.ID.String()
		}
		if r.limiter.TerlaluSering(key) {
			c.Set("Retry-After", strconv.Itoa(int(r.limiter.jendela.Seconds())))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Terlalu banyak permintaan. Coba lagi nanti.",
			})
		}
		r.limiter.Catat(key)
		return c.Next()
	}
}
