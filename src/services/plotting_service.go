package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/src/dto"
	"github.com/grafika-scheduling/backend/src/models"
	"gorm.io/gorm"
)

type LayananPlotting struct {
	db *gorm.DB
}

func NewLayananPlotting(db *gorm.DB) *LayananPlotting {
	return &LayananPlotting{db: db}
}

func parseID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("id tidak valid: %s", raw)
	}
	return id, nil
}

// BuatPlotting memvalidasi cross-semester lalu menyimpan satu baris plotting.
// Aturan: kelas.semester_id dan ruangan.semester_id harus sama dengan semester path.
func (s *LayananPlotting) BuatPlotting(semesterID uuid.UUID, req dto.BuatPlottingRequest) (*models.Plotting, error) {
	kelasID, err := parseID(req.KelasID)
	if err != nil {
		return nil, err
	}
	hariID, err := parseID(req.HariID)
	if err != nil {
		return nil, err
	}
	jamID, err := parseID(req.JamPelajaranID)
	if err != nil {
		return nil, err
	}
	mapelID, err := parseID(req.MataPelajaranID)
	if err != nil {
		return nil, err
	}
	guruID, err := parseID(req.GuruID)
	if err != nil {
		return nil, err
	}
	var ruanganID *uuid.UUID
	if req.RuanganID != "" {
		r, err := parseID(req.RuanganID)
		if err != nil {
			return nil, err
		}
		ruanganID = &r
	}

	var kelas models.Kelas
	if err := s.db.Select("id", "semester_id").First(&kelas, "id = ?", kelasID).Error; err != nil {
		return nil, fmt.Errorf("kelas tidak ditemukan")
	}
	if kelas.SemesterID != semesterID {
		return nil, fmt.Errorf("kelas bukan milik semester ini")
	}
	if ruanganID != nil {
		var ruangan models.Ruangan
		if err := s.db.Select("id", "semester_id").First(&ruangan, "id = ?", *ruanganID).Error; err != nil {
			return nil, fmt.Errorf("ruangan tidak ditemukan")
		}
		if ruangan.SemesterID != semesterID {
			return nil, fmt.Errorf("ruangan bukan milik semester ini")
		}
	}
	type refGlobal struct {
		dest any
		id   uuid.UUID
	}
	for nama, ref := range map[string]refGlobal{
		"hari":           {&models.Hari{}, hariID},
		"jam_pelajaran":  {&models.JamPelajaran{}, jamID},
		"mata_pelajaran": {&models.MataPelajaran{}, mapelID},
		"guru":           {&models.Guru{}, guruID},
	} {
		if err := s.db.Select("id").First(ref.dest, "id = ?", ref.id).Error; err != nil {
			return nil, fmt.Errorf("%s tidak ditemukan", nama)
		}
	}

	item := &models.Plotting{
		SemesterID:      semesterID,
		KelasID:         kelasID,
		HariID:          hariID,
		JamPelajaranID:  jamID,
		MataPelajaranID: mapelID,
		GuruID:          guruID,
		RuanganID:       ruanganID,
	}
	if err := s.db.Create(item).Error; err != nil {
		return nil, err
	}
	s.db.Preload("Kelas").Preload("Hari").Preload("JamPelajaran").
		Preload("MataPelajaran").Preload("Guru").Preload("Ruangan").
		First(item, "id = ?", item.ID)
	return item, nil
}
