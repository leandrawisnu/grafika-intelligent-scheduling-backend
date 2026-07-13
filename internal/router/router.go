package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/grafika-scheduling/backend/internal/handlers"
	"github.com/grafika-scheduling/backend/pkg/mlclient"
	"gorm.io/gorm"
)

func New(db *gorm.DB, mlClient *mlclient.Client) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "Grafika Scheduling",
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		AllowMethods:  "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:  "*",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "layanan": "grafika"})
	})

	v1 := app.Group("/api/v1")

	pengelolaMaster := handlers.NewPengelolaMaster(db)
	pengelolaMaster.DaftarkanRute(v1)

	pengelolaJadwal := handlers.NewPengelolaJadwal(db, mlClient)
	pengelolaJadwal.DaftarkanRute(v1)

	return app
}
