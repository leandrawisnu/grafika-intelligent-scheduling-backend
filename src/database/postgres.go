package database

import (
	"log"

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
	return db
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
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
