package database

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Gagal mengambil pool database: %v", err)
	}
	// Batas pool mencegah 50 VU membuka koneksi tanpa batas (Postgres memutus, klien dapat EOF).
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	return db
}

// AutoMigrate deprecated: schema dikelola eksklusif via migrate CLI
// (make migrate-init / migrate-up) dan database/migrations/*.sql.
// Jangan dipakai — dipertahankan agar kode lama yang memanggilnya tetap kompilasi.
func AutoMigrate(db *gorm.DB) error {
	_ = db
	return nil
}
