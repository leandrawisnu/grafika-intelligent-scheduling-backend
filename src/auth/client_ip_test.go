package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestClientIPUsesLastForwardedForEntry(t *testing.T) {
	app := fiber.New()
	var got string
	app.Get("/", func(c *fiber.Ctx) error {
		got = ClientIP(c)
		return c.SendStatus(fiber.StatusOK)
	})

	// A client may prepend arbitrary entries; the trusted
	// proxy appends the real peer last.
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "spoofed.example, 10.0.0.5, 203.0.113.7")

	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if want := "203.0.113.7"; got != want {
		t.Fatalf("ClientIP() = %q, want %q", got, want)
	}
}

func TestClientIPSplitsAndTrimsEntries(t *testing.T) {
	app := fiber.New()
	var got string
	app.Get("/", func(c *fiber.Ctx) error {
		got = ClientIP(c)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", " 198.51.100.1 ,203.0.113.9 ")

	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if want := "203.0.113.9"; got != want {
		t.Fatalf("ClientIP() = %q, want %q (trimmed last entry)", got, want)
	}
}

func TestClientIPFallsBackWithoutHeader(t *testing.T) {
	app := fiber.New()
	var got string
	app.Get("/", func(c *fiber.Ctx) error {
		got = ClientIP(c)
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	if _, err := app.Test(req); err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	// app.Test serves over an internal conn, so the exact
	// remote address is not deterministic — it must simply
	// fall back to a non-empty remote IP.
	if got == "" {
		t.Fatal("ClientIP() = empty, want non-empty remote IP fallback")
	}
}
