package auth

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/grafika-scheduling/backend/src/models"
)

const HeaderSesi = "X-GIS-Session"

type Middleware struct {
	layanan *Layanan
	lingkup *Lingkup
}

func NewMiddleware(layanan *Layanan, lingkup *Lingkup) *Middleware {
	return &Middleware{layanan: layanan, lingkup: lingkup}
}

func (m *Middleware) WajibSesi(c *fiber.Ctx) error {
	token := TokenDari(c)
	akun, err := m.layanan.PenggunaDariToken(token)
	if err != nil {
		c.Set("Cache-Control", "no-store")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Sesi tidak berlaku."})
	}
	if err := m.lingkup.Periksa(c, akun); err != nil {
		c.Set("Cache-Control", "no-store")
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Akses ditolak."})
	}
	c.Locals(LocalPengguna, akun)
	c.Set("Cache-Control", "no-store")
	return c.Next()
}

func TokenDari(c *fiber.Ctx) string {
	if h := strings.TrimSpace(c.Get(HeaderSesi)); h != "" {
		return h
	}
	if v := c.Cookies("__Host-gis_session"); v != "" {
		return v
	}
	return c.Cookies("gis_session")
}

func Publik(akun models.Pengguna) fiber.Map {
	out := fiber.Map{
		"id":    akun.ID,
		"email": akun.Email,
		"peran": akun.Peran,
		"aktif": akun.Aktif,
	}
	if akun.JurusanID != nil {
		out["jurusan_id"] = akun.JurusanID
	}
	return out
}
