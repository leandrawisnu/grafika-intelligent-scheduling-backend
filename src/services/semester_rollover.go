package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/src/dto"
	"github.com/grafika-scheduling/backend/src/models"
	"gorm.io/gorm"
)

type LayananRollover struct {
	db *gorm.DB
}

func NewLayananRollover(db *gorm.DB) *LayananRollover {
	return &LayananRollover{db: db}
}

// SalinSemester menyalin snapshot master (jurusan, ruangan, kelas, plotting) dari
// semester sumber ke semester target baru, lalu membuat jadwal_semester + jadwal_kelas
// kosong (tanpa slot). Semua dalam satu transaksi.
func (s *LayananRollover) SalinSemester(sumberID uuid.UUID, req dto.SalinSemesterRequest) (*models.Semester, error) {
	tahunAjaranID, err := uuid.Parse(req.TahunAjaranID)
	if err != nil {
		return nil, fmt.Errorf("tahun_ajaran_id tidak valid")
	}
	if req.SemesterKe != 1 && req.SemesterKe != 2 {
		return nil, fmt.Errorf("semester_ke harus 1 atau 2")
	}
	if req.Nama == "" {
		return nil, fmt.Errorf("nama semester target wajib diisi")
	}

	var sumber models.Semester
	if err := s.db.First(&sumber, "id = ?", sumberID).Error; err != nil {
		return nil, fmt.Errorf("semester sumber tidak ditemukan")
	}
	if sumber.TanggalMulai.Time.IsZero() || sumber.TanggalSelesai.Time.IsZero() {
		return nil, fmt.Errorf("semester sumber tanggalnya tidak lengkap")
	}
	var bentrok int64
	s.db.Model(&models.Semester{}).
		Where("tahun_ajaran_id = ? AND semester_ke = ?", tahunAjaranID, req.SemesterKe).
		Count(&bentrok)
	if bentrok > 0 {
		return nil, fmt.Errorf("semester target sudah ada")
	}

	var target models.Semester
	err = s.db.Transaction(func(tx *gorm.DB) error {
		target = models.Semester{
			TahunAjaranID:  tahunAjaranID,
			Nama:           req.Nama,
			SemesterKe:     req.SemesterKe,
			TanggalMulai:   sumber.TanggalMulai,
			TanggalSelesai: sumber.TanggalSelesai,
		}
		if err := tx.Create(&target).Error; err != nil {
			return err
		}

		// 1. Jurusan: UUID baru, kode sama.
		var jurusans []models.Jurusan
		if err := tx.Where("semester_id = ?", sumberID).Find(&jurusans).Error; err != nil {
			return err
		}
		if len(jurusans) == 0 {
			return fmt.Errorf("semester sumber tidak punya master jurusan")
		}
		petaJurusan := make(map[uuid.UUID]uuid.UUID, len(jurusans))
		for _, j := range jurusans {
			baru := models.Jurusan{Kode: j.Kode, Nama: j.Nama, SemesterID: target.ID}
			if err := tx.Create(&baru).Error; err != nil {
				return err
			}
			petaJurusan[j.ID] = baru.ID
		}

		// 2. Ruangan: UUID baru, kode sama.
		var ruangans []models.Ruangan
		if err := tx.Where("semester_id = ?", sumberID).Find(&ruangans).Error; err != nil {
			return err
		}
		petaRuangan := make(map[uuid.UUID]uuid.UUID, len(ruangans))
		for _, r := range ruangans {
			baru := models.Ruangan{
				Kode: r.Kode, Nama: r.Nama, Kapasitas: r.Kapasitas,
				TipeRuangan: r.TipeRuangan, Aktif: r.Aktif, SemesterID: target.ID,
			}
			if err := tx.Create(&baru).Error; err != nil {
				return err
			}
			petaRuangan[r.ID] = baru.ID
		}

		// 3. Kelas: UUID baru, remap jurusan_id.
		var kelass []models.Kelas
		if err := tx.Where("semester_id = ?", sumberID).Find(&kelass).Error; err != nil {
			return err
		}
		petaKelas := make(map[uuid.UUID]uuid.UUID, len(kelass))
		petaKelasJurusan := make(map[uuid.UUID]uuid.UUID, len(kelass))
		for _, k := range kelass {
			jurBaru, ok := petaJurusan[k.JurusanID]
			if !ok {
				return fmt.Errorf("jurusan kelas %s tidak ikut tersalin", k.Kode)
			}
			baru := models.Kelas{
				Kode: k.Kode, Nama: k.Nama, Tingkat: k.Tingkat,
				JurusanID: jurBaru, SemesterID: target.ID,
			}
			if err := tx.Create(&baru).Error; err != nil {
				return err
			}
			petaKelas[k.ID] = baru.ID
			petaKelasJurusan[baru.ID] = jurBaru
		}

		// 4. Plotting: remap kelas_id + ruangan_id (per-semester);
		//    guru/mapel/hari/jam tetap (global).
		var plottings []models.Plotting
		if err := tx.Where("semester_id = ?", sumberID).Find(&plottings).Error; err != nil {
			return err
		}
		for _, p := range plottings {
			kelasBaru, ok := petaKelas[p.KelasID]
			if !ok {
				return fmt.Errorf("kelas plotting tidak ikut tersalin")
			}
			baru := models.Plotting{
				SemesterID: target.ID, KelasID: kelasBaru,
				HariID: p.HariID, JamPelajaranID: p.JamPelajaranID,
				MataPelajaranID: p.MataPelajaranID, GuruID: p.GuruID,
			}
			if p.RuanganID != nil {
				ruangBaru, ok := petaRuangan[*p.RuanganID]
				if !ok {
					return fmt.Errorf("ruangan plotting tidak ikut tersalin")
				}
				baru.RuanganID = &ruangBaru
			}
			if err := tx.Create(&baru).Error; err != nil {
				return err
			}
		}

		// 5. Jadwal semester kosong.
		js := models.JadwalSemester{SemesterID: target.ID, Status: "draf"}
		if err := tx.Create(&js).Error; err != nil {
			return err
		}
		seenJurusan := map[uuid.UUID]bool{}
		for _, idBaru := range petaJurusan {
			if seenJurusan[idBaru] {
				continue
			}
			seenJurusan[idBaru] = true
			if err := tx.Create(&models.JadwalSemesterJurusan{
				JadwalSemesterID: js.ID, JurusanID: idBaru,
			}).Error; err != nil {
				return err
			}
		}
		for kelasBaru, jurBaru := range petaKelasJurusan {
			if err := tx.Create(&models.JadwalKelas{
				JadwalSemesterID: js.ID, JurusanID: jurBaru,
				KelasID: kelasBaru, Versi: 1, IsActive: true,
			}).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.db.Preload("TahunAjaran").First(&target, "id = ?", target.ID)
	return &target, nil
}
