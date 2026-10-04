package services

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/src/models"
	"gorm.io/gorm"
)

type LayananDemoKonflik struct {
	db *gorm.DB
}

func NewLayananDemoKonflik(db *gorm.DB) *LayananDemoKonflik {
	return &LayananDemoKonflik{db: db}
}

type HasilDemoKonflik struct {
	GuruBentrok      int `json:"guru_bentrok"`
	RuanganBentrok   int `json:"ruangan_bentrok"`
	GuruHariLibur    int `json:"guru_hari_libur"`
	GuruKelebihanJam int `json:"guru_kelebihan_jam"`
}

type kunciSel struct {
	HariID string
	JamID  string
}

// TambahKonflikAcak menyuntikkan konflik acak ke slot jadwal, lalu menjalankan
// ulang deteksi konflik. Hanya berlaku bila nama semester atau tahun ajaran
// mengandung kata "demo" — data asli tidak boleh diacak.
func (s *LayananDemoKonflik) TambahKonflikAcak(jsID uuid.UUID) (*HasilDemoKonflik, error) {
	var js models.JadwalSemester
	if err := s.db.Preload("Semester.TahunAjaran").First(&js, "id = ?", jsID).Error; err != nil {
		return nil, fmt.Errorf("jadwal semester tidak ditemukan")
	}
	label := strings.ToLower(js.Semester.Nama)
	if js.Semester.TahunAjaran != nil {
		label += " " + strings.ToLower(js.Semester.TahunAjaran.Nama)
	}
	if !strings.Contains(label, "demo") {
		return nil, fmt.Errorf("hanya untuk jadwal semester demo")
	}

	var slots []models.SlotJadwal
	if err := s.db.Joins("JOIN jadwal_kelas ON jadwal_kelas.id = slot_jadwal.jadwal_kelas_id").
		Where("jadwal_kelas.jadwal_semester_id = ? AND jadwal_kelas.is_active = ?", jsID, true).
		Find(&slots).Error; err != nil {
		return nil, fmt.Errorf("gagal memuat slot: %w", err)
	}
	if len(slots) == 0 {
		return nil, fmt.Errorf("jadwal ini belum punya slot")
	}

	var gurus []models.Guru
	s.db.Where("aktif = ?", true).Find(&gurus)
	var ruangans []models.Ruangan
	s.db.Where("aktif = ?", true).Find(&ruangans)
	var haris []models.Hari
	s.db.Where("akhir_pekan = ?", false).Find(&haris)

	sel := map[kunciSel][]models.SlotJadwal{}
	for _, slot := range slots {
		k := kunciSel{HariID: slot.HariID.String(), JamID: slot.JamPelajaranID.String()}
		sel[k] = append(sel[k], slot)
	}

	hasil := &HasilDemoKonflik{}

	// 1. Guru bentrok: satu guru mengajar dua kelas pada sel (hari, jam) yang sama.
	if len(gurus) > 0 {
		for _, k := range acakKunci(sel, 2) {
			grup := sel[k]
			if len(grup) < 2 {
				continue
			}
			guru := gurus[rand.Intn(len(gurus))]
			for i := 0; i < 2; i++ {
				s.db.Model(&models.SlotJadwal{}).Where("id = ?", grup[i].ID).
					Update("guru_id", guru.ID)
			}
			hasil.GuruBentrok++
		}
	}

	// 2. Ruangan bentrok: satu ruangan dipakai dua kelas pada sel yang sama.
	if len(ruangans) > 0 {
		for _, k := range acakKunci(sel, 1) {
			grup := sel[k]
			if len(grup) < 2 {
				continue
			}
			ruang := ruangans[rand.Intn(len(ruangans))]
			for i := 0; i < 2; i++ {
				s.db.Model(&models.SlotJadwal{}).Where("id = ?", grup[i].ID).
					Update("ruangan_id", ruang.ID)
			}
			hasil.RuanganBentrok++
		}
	}

	// 3. Guru hari libur: guru diplot ke slot pada hari liburnya.
	if len(gurus) > 0 && len(haris) > 0 {
		guru := gurus[rand.Intn(len(gurus))]
		hari := haris[rand.Intn(len(haris))]
		var ada int64
		s.db.Model(&models.HariLiburGuru{}).
			Where("guru_id = ? AND hari_id = ?", guru.ID, hari.ID).Count(&ada)
		if ada == 0 {
			s.db.Create(&models.HariLiburGuru{
				GuruID:     guru.ID,
				HariID:     hari.ID,
				SemesterID: js.SemesterID,
				Alasan:     "demo",
			})
		}
		for _, slot := range slots {
			if slot.HariID == hari.ID {
				s.db.Model(&models.SlotJadwal{}).Where("id = ?", slot.ID).
					Update("guru_id", guru.ID)
				hasil.GuruHariLibur++
				break
			}
		}
	}

	// 4. Guru kelebihan jam: guru dengan jam maksimal terendah diplot ke banyak slot.
	if len(gurus) > 0 {
		jumlah := map[uuid.UUID]int{}
		for _, slot := range slots {
			if slot.GuruID != uuid.Nil {
				jumlah[slot.GuruID]++
			}
		}
		target := gurus[0]
		for _, g := range gurus {
			if g.JamMaksimalPerMinggu < target.JamMaksimalPerMinggu {
				target = g
			}
		}
		butuh := int(target.JamMaksimalPerMinggu) - jumlah[target.ID] + 2
		if butuh < 2 {
			butuh = 2
		}
		ditambah := 0
		for _, slot := range slots {
			if butuh == 0 {
				break
			}
			if slot.GuruID == target.ID {
				continue
			}
			s.db.Model(&models.SlotJadwal{}).Where("id = ?", slot.ID).
				Update("guru_id", target.ID)
			butuh--
			ditambah++
		}
		if ditambah > 0 {
			hasil.GuruKelebihanJam++
		}
	}

	if _, err := NewLayananKonflik(s.db).DeteksiKonflik(jsID); err != nil {
		return hasil, fmt.Errorf("konflik disuntikkan tapi deteksi gagal: %w", err)
	}
	return hasil, nil
}

func acakKunci[V any](m map[kunciSel]V, n int) []kunciSel {
	kuncis := make([]kunciSel, 0, len(m))
	for k := range m {
		kuncis = append(kuncis, k)
	}
	rand.Shuffle(len(kuncis), func(i, j int) { kuncis[i], kuncis[j] = kuncis[j], kuncis[i] })
	if n > len(kuncis) {
		n = len(kuncis)
	}
	return kuncis[:n]
}
