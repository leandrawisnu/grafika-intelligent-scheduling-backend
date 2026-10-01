package database

import (
	"log"
	"time"

	"github.com/grafika-scheduling/backend/src/models"
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

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.TahunAjaran{},
		&models.Semester{},
		&models.Jurusan{},
		&models.Pengguna{},
		&models.Sesi{},
		&models.Guru{},
		&models.MataPelajaran{},
		&models.Kelas{},
		&models.Ruangan{},
		&models.Hari{},
		&models.JamPelajaran{},
		&models.HariLiburGuru{},
		&models.JadwalSemester{},
		&models.JadwalSemesterJurusan{},
		&models.JadwalKelas{},
		&models.SlotJadwal{},
		&models.Konflik{},
		&models.ResolusiAI{},
		&models.LogAuditJadwal{},
		&models.SesiChatAI{},
		&models.PesanChatAI{},
		&models.FeedbackChatAI{},
		&models.MemoriChatbot{},
	)
}
