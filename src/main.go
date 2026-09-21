package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/grafika-scheduling/backend/src/config"
	"github.com/grafika-scheduling/backend/src/database"
	"github.com/grafika-scheduling/backend/src/router"
	"github.com/grafika-scheduling/backend/pkg/mlclient"
)

func main() {
	migrateCmd := flag.String("migrate", "", "jalankan migrasi lalu keluar: init, up, down, version, force (force butuh arg versi di flag.Args)")
	flag.Parse()

	cfg := config.Load()
	dsn := cfg.DatabaseURL()

	if *migrateCmd != "" {
		if err := runMigrateMode(dsn, *migrateCmd, flag.Args()); err != nil {
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
	mlClient := mlclient.NewClient(cfg.MLServiceURL)

	app := router.New(db, mlClient)

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
