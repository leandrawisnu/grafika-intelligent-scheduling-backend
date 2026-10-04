package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/grafika-scheduling/backend/pkg/mlclient"
	"github.com/grafika-scheduling/backend/pkg/storage"
	"github.com/grafika-scheduling/backend/pkg/timezone"
	"github.com/grafika-scheduling/backend/src/auth"
	"github.com/grafika-scheduling/backend/src/config"
	"github.com/grafika-scheduling/backend/src/database"
	"github.com/grafika-scheduling/backend/src/models"
	"github.com/grafika-scheduling/backend/src/router"
	"gorm.io/gorm"
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
	bersihkanDokumenImporMacet(db)
	if err := auth.NewLayanan(db).PastikanAdmin(cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Fatalf("akun admin: %v", err)
	}
	objek, err := siapMinIO(cfg)
	if err != nil {
		log.Fatalf("minio: %v", err)
	}
	mlClient := mlclient.NewClient(cfg.MLServiceURL)

	app := router.New(db, mlClient, objek, cfg)

	log.Printf("Grafika Scheduling mendengarkan %s\n", cfg.ListenAddr())
	if err := app.Listen(cfg.ListenAddr()); err != nil {
		log.Fatal(err)
	}
}

// bersihkanDokumenImporMacet menandai job impor AI yang tertinggal "memproses"
// (mis. server berhenti di tengah proses) supaya user bisa menjalankan ulangi.
func bersihkanDokumenImporMacet(db *gorm.DB) {
	if !db.Migrator().HasTable(&models.DokumenImpor{}) {
		return
	}
	hasil := db.Model(&models.DokumenImpor{}).
		Where("status = ?", "memproses").
		Updates(map[string]any{
			"status": "gagal",
			"pesan":  "Server berhenti saat memproses dokumen ini. Gunakan Ulangi analisis.",
		})
	if hasil.RowsAffected > 0 {
		log.Printf("%d dokumen impor macet ditandai gagal", hasil.RowsAffected)
	}
}

func siapMinIO(cfg *config.Config) (*storage.Client, error) {
	if !cfg.MinIOConfigured() {
		log.Printf("MinIO nonaktif (MINIO_ENDPOINT kosong)")
		return nil, nil
	}
	klien, err := storage.New(storage.Options{
		Endpoint:  cfg.MinIOEndpoint,
		AccessKey: cfg.MinIOAccessKey,
		SecretKey: cfg.MinIOSecretKey,
		Bucket:    cfg.MinIOBucket,
		UseSSL:    cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, err
	}
	ctx, batal := context.WithTimeout(context.Background(), 15*time.Second)
	defer batal()
	if err := klien.EnsureBucket(ctx); err != nil {
		return nil, err
	}
	log.Printf("MinIO bucket %s siap", cfg.MinIOBucket)
	return klien, nil
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
