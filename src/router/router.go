package router

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/grafika-scheduling/backend/pkg/mlclient"
	"github.com/grafika-scheduling/backend/src/auth"
	"github.com/grafika-scheduling/backend/src/config"
	"github.com/grafika-scheduling/backend/src/handlers"
	"gorm.io/gorm"
)

func New(db *gorm.DB, mlClient *mlclient.Client, cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Grafika Scheduling",
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     originsBersih(cfg.CORSOrigins),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, X-GIS-Session",
		AllowCredentials: true,
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "layanan": "grafika"})
	})

	layananAuth := auth.NewLayanan(db)
	mw := auth.NewMiddleware(layananAuth, auth.NewLingkup(db))
	pengelolaAuth := handlers.NewPengelolaAuth(layananAuth)

	v1 := app.Group("/api/v1")
	v1.Post("/auth/masuk", pengelolaAuth.Masuk)
	v1.Use(mw.WajibSesi)
	v1.Post("/auth/keluar", pengelolaAuth.Keluar)
	v1.Get("/auth/sesi", pengelolaAuth.Sesi)

	pengelolaMaster := handlers.NewPengelolaMaster(db)
	pengelolaMaster.DaftarkanRute(v1)

	pengelolaJadwal := handlers.NewPengelolaJadwal(db, mlClient)
	pengelolaJadwal.DaftarkanRute(v1)

	return app
}

func originsBersih(raw string) string {
	bagian := strings.Split(raw, ",")
	bersih := make([]string, 0, len(bagian))
	for _, item := range bagian {
		item = strings.TrimSpace(item)
		if item == "" || item == "*" {
			continue
		}
		bersih = append(bersih, item)
	}
	if len(bersih) == 0 {
		return "http://localhost:3000"
	}
	return strings.Join(bersih, ",")
}
