package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/internal/models"
	"gorm.io/gorm"
)

type LayananKonflik struct {
	db *gorm.DB
}

func NewLayananKonflik(db *gorm.DB) *LayananKonflik {
	return &LayananKonflik{db: db}
}

func (s *LayananKonflik) DeteksiKonflik(jsID uuid.UUID) ([]models.Konflik, error) {
	// Ambil semua slot dari jadwal_kelas aktif dalam jadwal_semester ini
	var slots []models.SlotJadwal
	if err := s.db.Joins("JOIN jadwal_kelas ON jadwal_kelas.id = slot_jadwal.jadwal_kelas_id").
		Where("jadwal_kelas.jadwal_semester_id = ? AND jadwal_kelas.is_active = ?", jsID, true).
		Preload("Kelas").Preload("MataPelajaran").Preload("Hari").Preload("JamPelajaran").
		Preload("Ruangan").Preload("Guru").
		Find(&slots).Error; err != nil {
		return nil, fmt.Errorf("gagal memuat slot: %w", err)
	}

	var konflik []models.Konflik

	konflik = append(konflik, s.cekGuruBentrok(jsID, slots)...)
	konflik = append(konflik, s.cekRuanganBentrok(jsID, slots)...)
	konflik = append(konflik, s.cekKelasBentrok(jsID, slots)...)
	konflik = append(konflik, s.cekGuruKelebihanJam(jsID, slots)...)
	konflik = append(konflik, s.cekGuruHariLibur(jsID, slots)...)
	konflik = append(konflik, s.cekGuruTidakBerkualifikasi(jsID, slots)...)

	// Hapus konflik lama dan simpan yang baru
	s.db.Where("jadwal_semester_id = ?", jsID).Delete(&models.Konflik{})
	if len(konflik) > 0 {
		if err := s.db.Create(&konflik).Error; err != nil {
			return nil, fmt.Errorf("gagal menyimpan konflik: %w", err)
		}
	}

	// Perbarui status bebas konflik
	bebasKonflik := len(konflik) == 0
	s.db.Model(&models.JadwalSemester{}).Where("id = ?", jsID).
		Update("bebas_konflik", bebasKonflik)

	return konflik, nil
}

func (s *LayananKonflik) cekGuruBentrok(jsID uuid.UUID, slots []models.SlotJadwal) []models.Konflik {
	var konflik []models.Konflik
	type kunci struct {
		GuruID, HariID, JamPelajaranID string
		Minggu                         int16
	}
	terlihat := make(map[kunci][]models.SlotJadwal)

	for _, slot := range slots {
		if slot.GuruID == uuid.Nil {
			continue
		}
		k := kunci{slot.GuruID.String(), slot.HariID.String(), slot.JamPelajaranID.String(), slot.MingguKe}
		terlihat[k] = append(terlihat[k], slot)
	}

	for _, grup := range terlihat {
		if len(grup) < 2 {
			continue
		}
		for i := 1; i < len(grup); i++ {
			aID := grup[0].ID
			bID := grup[i].ID
			namaGuru := "Tidak Diketahui"
			if grup[0].Guru != nil {
				namaGuru = grup[0].Guru.NamaLengkap
			}
			konflik = append(konflik, models.Konflik{
				JadwalSemesterID: jsID,
				TipeKonflik:      "guru_bentrok",
				TingkatKeparahan: "kesalahan",
				SlotAID:          &aID,
				SlotBID:          &bID,
				GuruID:           &grup[0].GuruID,
				Deskripsi:        fmt.Sprintf("Guru %s mengajar dua kelas (%s & %s) pada jam %s hari %s", namaGuru, grup[0].Kelas.Nama, grup[i].Kelas.Nama, grup[0].JamPelajaran.WaktuMulai, grup[0].Hari.Nama),
				TerdeteksiPada:   time.Now(),
			})
		}
	}
	return konflik
}

func (s *LayananKonflik) cekRuanganBentrok(jsID uuid.UUID, slots []models.SlotJadwal) []models.Konflik {
	var konflik []models.Konflik
	type kunci struct {
		RuanganID, HariID, JamPelajaranID string
		Minggu                             int16
	}
	terlihat := make(map[kunci][]models.SlotJadwal)

	for _, slot := range slots {
		if slot.RuanganID == uuid.Nil {
			continue
		}
		k := kunci{slot.RuanganID.String(), slot.HariID.String(), slot.JamPelajaranID.String(), slot.MingguKe}
		terlihat[k] = append(terlihat[k], slot)
	}

	for _, grup := range terlihat {
		if len(grup) < 2 {
			continue
		}
		for i := 1; i < len(grup); i++ {
			aID := grup[0].ID
			bID := grup[i].ID
			namaRuangan := "Tidak Diketahui"
			if grup[0].Ruangan != nil {
				namaRuangan = grup[0].Ruangan.Nama
			}
			konflik = append(konflik, models.Konflik{
				JadwalSemesterID: jsID,
				TipeKonflik:      "ruangan_bentrok",
				TingkatKeparahan: "kesalahan",
				SlotAID:          &aID,
				SlotBID:          &bID,
				Deskripsi:        fmt.Sprintf("Ruangan %s digunakan oleh %s dan %s pada jam %s hari %s", namaRuangan, grup[0].Kelas.Nama, grup[i].Kelas.Nama, grup[0].JamPelajaran.WaktuMulai, grup[0].Hari.Nama),
				TerdeteksiPada:   time.Now(),
			})
		}
	}
	return konflik
}

func (s *LayananKonflik) cekKelasBentrok(jsID uuid.UUID, slots []models.SlotJadwal) []models.Konflik {
	var konflik []models.Konflik
	type kunci struct {
		KelasID, HariID, JamPelajaranID string
		Minggu                           int16
	}
	terlihat := make(map[kunci][]models.SlotJadwal)

	for _, slot := range slots {
		k := kunci{slot.KelasID.String(), slot.HariID.String(), slot.JamPelajaranID.String(), slot.MingguKe}
		terlihat[k] = append(terlihat[k], slot)
	}

	for _, grup := range terlihat {
		if len(grup) < 2 {
			continue
		}
		for i := 1; i < len(grup); i++ {
			aID := grup[0].ID
			bID := grup[i].ID
			konflik = append(konflik, models.Konflik{
				JadwalSemesterID: jsID,
				TipeKonflik:      "kelas_bentrok",
				TingkatKeparahan: "kesalahan",
				SlotAID:          &aID,
				SlotBID:          &bID,
				Deskripsi:        fmt.Sprintf("Kelas %s memiliki dua mata pelajaran (%s & %s) pada jam %s hari %s", grup[0].Kelas.Nama, grup[0].MataPelajaran.Nama, grup[i].MataPelajaran.Nama, grup[0].JamPelajaran.WaktuMulai, grup[0].Hari.Nama),
				TerdeteksiPada:   time.Now(),
			})
		}
	}
	return konflik
}

func (s *LayananKonflik) cekGuruKelebihanJam(jsID uuid.UUID, slots []models.SlotJadwal) []models.Konflik {
	var konflik []models.Konflik
	jamMingguan := make(map[string]float64)
	maksMingguan := make(map[string]float64)
	namaGuru := make(map[string]string)

	for _, slot := range slots {
		if slot.GuruID == uuid.Nil {
			continue
		}
		gid := slot.GuruID.String()
		jamMingguan[gid] += 1.0
		if slot.Guru != nil {
			maksMingguan[gid] = slot.Guru.JamMaksimalPerMinggu
			namaGuru[gid] = slot.Guru.NamaLengkap
		}
	}

	for gid, jam := range jamMingguan {
		maks := maksMingguan[gid]
		if jam > maks {
			parsed, _ := uuid.Parse(gid)
			konflik = append(konflik, models.Konflik{
				JadwalSemesterID: jsID,
				TipeKonflik:      "guru_kelebihan_jam",
				TingkatKeparahan: "peringatan",
				GuruID:           &parsed,
				Deskripsi:        fmt.Sprintf("Guru %s memiliki total %.0f jam mengajar per minggu, melebihi batas maksimal %.0f jam", namaGuru[gid], jam, maks),
				TerdeteksiPada:   time.Now(),
			})
		}
	}
	return konflik
}

func (s *LayananKonflik) cekGuruHariLibur(jsID uuid.UUID, slots []models.SlotJadwal) []models.Konflik {
	var konflik []models.Konflik

	for _, slot := range slots {
		if slot.GuruID == uuid.Nil {
			continue
		}
		var count int64
		s.db.Model(&models.HariLiburGuru{}).
			Where("guru_id = ? AND hari_id = ?", slot.GuruID, slot.HariID).
			Count(&count)
		if count > 0 {
			namaGuru := "Tidak Diketahui"
			namaHari := "Tidak Diketahui"
			if slot.Guru != nil {
				namaGuru = slot.Guru.NamaLengkap
			}
			if slot.Hari != nil {
				namaHari = slot.Hari.Nama
			}
			konflik = append(konflik, models.Konflik{
				JadwalSemesterID: jsID,
				TipeKonflik:      "guru_hari_libur",
				TingkatKeparahan: "kesalahan",
				SlotAID:          &slot.ID,
				GuruID:           &slot.GuruID,
				Deskripsi:        fmt.Sprintf("Guru %s dijadwalkan pada hari %s yang merupakan hari liburnya", namaGuru, namaHari),
				TerdeteksiPada:   time.Now(),
			})
		}
	}
	return konflik
}

func (s *LayananKonflik) cekGuruTidakBerkualifikasi(jsID uuid.UUID, slots []models.SlotJadwal) []models.Konflik {
	var konflik []models.Konflik

	for _, slot := range slots {
		if slot.GuruID == uuid.Nil {
			continue
		}
		var count int64
		s.db.Model(&models.KualifikasiGuru{}).
			Where("guru_id = ? AND mata_pelajaran_id = ?", slot.GuruID, slot.MataPelajaranID).
			Count(&count)
		if count == 0 {
			namaGuru := "Tidak Diketahui"
			namaMapel := "Tidak Diketahui"
			if slot.Guru != nil {
				namaGuru = slot.Guru.NamaLengkap
			}
			if slot.MataPelajaran != nil {
				namaMapel = slot.MataPelajaran.Nama
			}
			konflik = append(konflik, models.Konflik{
				JadwalSemesterID: jsID,
				TipeKonflik:      "guru_tidak_berkualifikasi",
				TingkatKeparahan: "peringatan",
				SlotAID:          &slot.ID,
				GuruID:           &slot.GuruID,
				Deskripsi:        fmt.Sprintf("Guru %s tidak memiliki kualifikasi untuk mengajar %s", namaGuru, namaMapel),
				TerdeteksiPada:   time.Now(),
			})
		}
	}
	return konflik
}
