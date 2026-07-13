package database

import (
	"log"

	"github.com/grafika-scheduling/backend/internal/models"
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

	if err := db.AutoMigrate(
		&models.TahunAjaran{},
		&models.Semester{},
		&models.Jurusan{},
		&models.Guru{},
		&models.MataPelajaran{},
		&models.Kelas{},
		&models.Ruangan{},
		&models.Hari{},
		&models.JamPelajaran{},
		&models.HariLiburGuru{},
		&models.KualifikasiGuru{},
		&models.JadwalSemester{},
		&models.JadwalSemesterJurusan{},
		&models.JadwalKelas{},
		&models.SlotJadwal{},
		&models.Konflik{},
		&models.ResolusiAI{},
		&models.LogAuditJadwal{},
	); err != nil {
		log.Fatalf("Gagal auto-migrate: %v", err)
	}

	return db
}
