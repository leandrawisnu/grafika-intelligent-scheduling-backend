package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/grafika-scheduling/backend/pkg/mlclient"
	"github.com/grafika-scheduling/backend/pkg/timezone"
	"github.com/grafika-scheduling/backend/src/auth"
	"github.com/grafika-scheduling/backend/src/config"
	"github.com/grafika-scheduling/backend/src/database"
	"github.com/grafika-scheduling/backend/src/router"
)

func main() {
	migrateFlag := flag.Bool("migrate", false, "jalankan migrasi lalu keluar. Tanpa argumen sama dengan up. Argumen: init, up, down, version, force <versi>")
	flag.Parse()

	cfg := config.Load()
	time.Local = timezone.Loc
	dsn := cfg.DatabaseURL()

	if *migrateFlag {
		command := "up"
		args := flag.Args()
		if len(args) > 0 {
			command = args[0]
			args = args[1:]
		}
		if err := runMigrateMode(dsn, command, args); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		return
	}

	db := database.Connect(dsn)
	if cfg.AutoMigrate {
		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("Gagal auto-migrate: %v", err)
		}
	}
	if err := auth.NewLayanan(db).PastikanAdmin(cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Fatalf("akun admin: %v", err)
	}
	mlClient := mlclient.NewClient(cfg.MLServiceURL)

	app := router.New(db, mlClient, cfg)

	log.Printf("Grafika Scheduling mendengarkan %s\n", cfg.ListenAddr())
	if err := app.Listen(cfg.ListenAddr()); err != nil {
		log.Fatal(err)
	}
}

func runMigrateMode(dsn, command string, args []string) error {
	switch command {
	case "init":
		return database.RunInitSQL(dsn)
	default:
		migrationsPath := os.Getenv("MIGRATIONS_PATH")
		if err := database.RunMigrateCommand(dsn, migrationsPath, command, args); err != nil {
			return err
		}
		if command != "version" {
			fmt.Println("migrate selesai:", command)
		}
		return nil
	}
}
