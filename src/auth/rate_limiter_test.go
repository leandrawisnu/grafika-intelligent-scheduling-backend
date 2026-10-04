package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/src/models"
)

// newTestApp wires the limiter the way router.New does:
// a first middleware simulates WajibSesi by injecting the
// user named in the X-Test-User header into locals.
func newTestApp(limiter *RateLimiter) *fiber.App {
	app := fiber.New()
	app.Post("/heavy",
		func(c *fiber.Ctx) error {
			id := uuid.NewSHA1(uuid.NameSpaceOID, []byte(c.Get("X-Test-User")))
			c.Locals(LocalPengguna, models.Pengguna{BaseModel: models.BaseModel{ID: id}})
			return c.Next()
		},
		limiter.Handler(),
		func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusOK)
		},
	)
	return app
}

func postHeavy(app *fiber.App, user string) (*http.Response, error) {
	req := httptest.NewRequest("POST", "/heavy", nil)
	req.Header.Set("X-Test-User", user)
	return app.Test(req)
}

func TestRateLimiterAllowsUpToLimitThenRejects(t *testing.T) {
	app := newTestApp(NewRateLimiter(3, time.Minute))

	for i := 1; i <= 3; i++ {
		res, err := postHeavy(app, "alice")
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		if res.StatusCode != fiber.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i, res.StatusCode)
		}
	}

	res, err := postHeavy(app, "alice")
	if err != nil {
		t.Fatalf("request 4: %v", err)
	}
	if res.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("request 4: status = %d, want 429", res.StatusCode)
	}
	if res.Header.Get("Retry-After") == "" {
		t.Fatal("429 response missing Retry-After header")
	}
}

func TestRateLimiterKeysPerUser(t *testing.T) {
	app := newTestApp(NewRateLimiter(2, time.Minute))

	// Exhaust alice's budget.
	for i := 0; i < 2; i++ {
		res, err := postHeavy(app, "alice")
		if err != nil {
			t.Fatalf("alice request %d: %v", i, err)
		}
		if res.StatusCode != fiber.StatusOK {
			t.Fatalf("alice request %d: status = %d, want 200", i, res.StatusCode)
		}
	}

	// Bob has an independent budget.
	res, err := postHeavy(app, "bob")
	if err != nil {
		t.Fatalf("bob: %v", err)
	}
	if res.StatusCode != fiber.StatusOK {
		t.Fatalf("bob: status = %d, want 200 (independent budget)", res.StatusCode)
	}

	// Alice is still locked out.
	res, err = postHeavy(app, "alice")
	if err != nil {
		t.Fatalf("alice: %v", err)
	}
	if res.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("alice: status = %d, want 429", res.StatusCode)
	}
}

func TestRateLimiterFallsBackToPerIPKey(t *testing.T) {
	// No user in locals — requests are keyed by client IP.
	app := fiber.New()
	limiter := NewRateLimiter(2, time.Minute)
	app.Post("/heavy",
		limiter.Handler(),
		func(c *fiber.Ctx) error { return c.SendStatus(fiber.StatusOK) },
	)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("POST", "/heavy", nil)
		res, err := app.Test(req)
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		if res.StatusCode != fiber.StatusOK {
			t.Fatalf("request %d: status = %d, want 200", i, res.StatusCode)
		}
	}

	req := httptest.NewRequest("POST", "/heavy", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatalf("request 3: %v", err)
	}
	if res.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("request 3: status = %d, want 429 (per-IP fallback)", res.StatusCode)
	}
}

func TestRateLimiterDisabledWhenLimitNotPositive(t *testing.T) {
	app := newTestApp(NewRateLimiter(0, time.Minute))

	for i := 0; i < 10; i++ {
		res, err := postHeavy(app, "alice")
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		if res.StatusCode != fiber.StatusOK {
			t.Fatalf("request %d: status = %d, want 200 (limiter disabled)", i, res.StatusCode)
		}
	}
}
