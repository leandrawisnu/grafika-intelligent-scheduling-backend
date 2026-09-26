package handlers

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/grafika-scheduling/backend/src/auth"
	"github.com/grafika-scheduling/backend/src/models"
)

type PengelolaAuth struct {
	layanan *auth.Layanan
}

func NewPengelolaAuth(layanan *auth.Layanan) *PengelolaAuth {
	return &PengelolaAuth{layanan: layanan}
}

func (h *PengelolaAuth) Masuk(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	token, akun, err := h.layanan.Masuk(c.IP(), req.Email, req.Password)
	c.Set("Cache-Control", "no-store")
	if errors.Is(err, auth.ErrTerlaluSering) {
		return c.Status(429).JSON(fiber.Map{"error": "Terlalu banyak percobaan. Coba lagi nanti."})
	}
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Email atau kata sandi salah."})
	}
	return c.JSON(fiber.Map{
		"token":    token,
		"pengguna": auth.Publik(akun),
	})
}

func (h *PengelolaAuth) Keluar(c *fiber.Ctx) error {
	h.layanan.HapusToken(auth.TokenDari(c))
	c.Set("Cache-Control", "no-store")
	return c.JSON(fiber.Map{"status": "keluar"})
}

func (h *PengelolaAuth) Sesi(c *fiber.Ctx) error {
	akun, _ := c.Locals(auth.LocalPengguna).(models.Pengguna)
	c.Set("Cache-Control", "no-store")
	return c.JSON(auth.Publik(akun))
}
