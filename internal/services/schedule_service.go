package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/internal/models"
	"gorm.io/gorm"
)

type LayananJadwal struct {
	db *gorm.DB
}

func NewLayananJadwal(db *gorm.DB) *LayananJadwal {
	return &LayananJadwal{db: db}
}

// ========== Jadwal Semester ==========

func (s *LayananJadwal) BuatJadwalSemester(semesterID uuid.UUID) (*models.JadwalSemester, error) {
	js := &models.JadwalSemester{SemesterID: semesterID, Status: "draf"}
	if err := s.db.Create(js).Error; err != nil {
		return nil, fmt.Errorf("gagal buat jadwal semester: %w", err)
	}
	return js, nil
}

func (s *LayananJadwal) DaftarJadwalSemester() ([]models.JadwalSemester, error) {
	var list []models.JadwalSemester
	if err := s.db.Preload("Semester.TahunAjaran").Preload("Jurusan.Jurusan").Order("dibuat_pada DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (s *LayananJadwal) AmbilJadwalSemester(id uuid.UUID) (*models.JadwalSemester, error) {
	var js models.JadwalSemester
	if err := s.db.Preload("Semester.TahunAjaran").
		Preload("Jurusan.Jurusan").
		Preload("JadwalKelas", func(db *gorm.DB) *gorm.DB {
			return db.Order("versi DESC")
		}).
		Preload("JadwalKelas.Kelas").
		Preload("JadwalKelas.SlotJadwal").
		Preload("JadwalKelas.SlotJadwal.Kelas").
		Preload("JadwalKelas.SlotJadwal.MataPelajaran").
		Preload("JadwalKelas.SlotJadwal.Hari").
		Preload("JadwalKelas.SlotJadwal.JamPelajaran").
		Preload("JadwalKelas.SlotJadwal.Ruangan").
		Preload("JadwalKelas.SlotJadwal.Guru").
		First(&js, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &js, nil
}

func (s *LayananJadwal) TambahJurusanKeSemester(id uuid.UUID, jurusanIDs []string) error {
	for _, jid := range jurusanIDs {
		jurID, _ := uuid.Parse(jid)
		s.db.FirstOrCreate(&models.JadwalSemesterJurusan{}, map[string]interface{}{
			"jadwal_semester_id": id,
			"jurusan_id":        jurID,
		})
	}
	return nil
}

func (s *LayananJadwal) HapusJurusanDariSemester(id, jurusanID uuid.UUID) error {
	return s.db.Where("jadwal_semester_id = ? AND jurusan_id = ?", id, jurusanID).
		Delete(&models.JadwalSemesterJurusan{}).Error
}

func (s *LayananJadwal) CekKesiapanAktivasi(id uuid.UUID) (bool, string) {
	var jurusans []models.JadwalSemesterJurusan
	s.db.Where("jadwal_semester_id = ?", id).Find(&jurusans)

	if len(jurusans) == 0 {
		return false, "Belum ada jurusan yang terdaftar"
	}

	for _, j := range jurusans {
		var count int64
		s.db.Model(&models.JadwalKelas{}).
			Where("jadwal_semester_id = ? AND jurusan_id = ? AND is_active = ?", id, j.JurusanID, true).
			Count(&count)
		if count == 0 {
			var jur models.Jurusan
			s.db.First(&jur, "id = ?", j.JurusanID)
			return false, fmt.Sprintf("Jurusan %s belum memiliki jadwal kelas yang aktif", jur.Nama)
		}
	}
	return true, ""
}

// ========== Jadwal Kelas ==========

func (s *LayananJadwal) BuatJadwalKelas(jsID, kelasID uuid.UUID) (*models.JadwalKelas, error) {
	var kelas models.Kelas
	if err := s.db.First(&kelas, "id = ?", kelasID).Error; err != nil {
		return nil, fmt.Errorf("kelas tidak ditemukan")
	}

	// Versi terakhir
	var last models.JadwalKelas
	versi := 1
	if err := s.db.Where("jadwal_semester_id = ? AND kelas_id = ?", jsID, kelasID).
		Order("versi DESC").First(&last).Error; err == nil {
		versi = last.Versi + 1
	}

	jk := &models.JadwalKelas{
		JadwalSemesterID: jsID,
		JurusanID:        kelas.JurusanID,
		KelasID:          kelasID,
		Versi:            versi,
		IsActive:         true,
	}
	if err := s.db.Create(jk).Error; err != nil {
		return nil, fmt.Errorf("gagal buat jadwal kelas: %w", err)
	}

	// Nonaktifkan versi lama
	if versi > 1 {
		s.db.Model(&models.JadwalKelas{}).
			Where("jadwal_semester_id = ? AND kelas_id = ? AND id != ?", jsID, kelasID, jk.ID).
			Update("is_active", false)
	}

	return jk, nil
}

func (s *LayananJadwal) AmbilJadwalKelas(id uuid.UUID) (*models.JadwalKelas, error) {
	var jk models.JadwalKelas
	if err := s.db.Preload("Kelas").Preload("SlotJadwal").
		Preload("SlotJadwal.MataPelajaran").Preload("SlotJadwal.Hari").
		Preload("SlotJadwal.JamPelajaran").Preload("SlotJadwal.Ruangan").
		Preload("SlotJadwal.Guru").First(&jk, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &jk, nil
}

func (s *LayananJadwal) AmbilSemuaJadwalKelasAktif(jsID uuid.UUID) ([]models.JadwalKelas, error) {
	var list []models.JadwalKelas
	if err := s.db.Where("jadwal_semester_id = ? AND is_active = ?", jsID, true).
		Preload("Kelas").Preload("SlotJadwal").
		Preload("SlotJadwal.Kelas").Preload("SlotJadwal.MataPelajaran").
		Preload("SlotJadwal.Hari").Preload("SlotJadwal.JamPelajaran").
		Preload("SlotJadwal.Ruangan").Preload("SlotJadwal.Guru").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ========== Slot ==========

func (s *LayananJadwal) TambahSlot(jkID uuid.UUID, slot *models.SlotJadwal) (*models.SlotJadwal, error) {
	slot.JadwalKelasID = jkID
	if err := s.db.Create(slot).Error; err != nil {
		return nil, fmt.Errorf("gagal tambah slot: %w", err)
	}
	return slot, nil
}

func (s *LayananJadwal) TambahSlotBulk(jkID uuid.UUID, slots []models.SlotJadwal) ([]models.SlotJadwal, error) {
	for i := range slots {
		slots[i].JadwalKelasID = jkID
	}
	if err := s.db.Create(&slots).Error; err != nil {
		return nil, fmt.Errorf("gagal tambah slot massal: %w", err)
	}
	return slots, nil
}

func (s *LayananJadwal) PerbaruiSlot(slotID uuid.UUID, data map[string]interface{}) (*models.SlotJadwal, error) {
	if err := s.db.Model(&models.SlotJadwal{}).Where("id = ?", slotID).Updates(data).Error; err != nil {
		return nil, fmt.Errorf("gagal perbarui slot: %w", err)
	}
	var slot models.SlotJadwal
	s.db.First(&slot, "id = ?", slotID)
	return &slot, nil
}

func (s *LayananJadwal) HapusSlot(slotID uuid.UUID) error {
	return s.db.Delete(&models.SlotJadwal{}, "id = ?", slotID).Error
}

// ========== Penempatan Guru ==========

func (s *LayananJadwal) AmbilSlotBelumDiplot(jsID uuid.UUID) ([]models.SlotJadwal, error) {
	var slots []models.SlotJadwal
	if err := s.db.Joins("JOIN jadwal_kelas ON jadwal_kelas.id = slot_jadwal.jadwal_kelas_id").
		Where("jadwal_kelas.jadwal_semester_id = ? AND jadwal_kelas.is_active = ? AND slot_jadwal.guru_id IS NULL", jsID, true).
		Preload("Kelas").Preload("MataPelajaran").Preload("Hari").
		Preload("JamPelajaran").Preload("Ruangan").
		Find(&slots).Error; err != nil {
		return nil, err
	}
	return slots, nil
}

func (s *LayananJadwal) TugaskanGuru(slotID, guruID uuid.UUID) (*models.SlotJadwal, error) {
	if err := s.db.Model(&models.SlotJadwal{}).Where("id = ?", slotID).
		Update("guru_id", guruID).Error; err != nil {
		return nil, fmt.Errorf("gagal tugaskan guru: %w", err)
	}
	var slot models.SlotJadwal
	s.db.Preload("Guru").First(&slot, "id = ?", slotID)
	return &slot, nil
}

func (s *LayananJadwal) TugaskanBulk(jsID uuid.UUID, tugas map[string]string) error {
	for slotID, guruID := range tugas {
		sID, _ := uuid.Parse(slotID)
		gID, _ := uuid.Parse(guruID)
		s.db.Model(&models.SlotJadwal{}).Where("id = ?", sID).Update("guru_id", gID)
	}
	return nil
}

func (s *LayananJadwal) AmbilKetersediaanGuru(jsID uuid.UUID) ([]map[string]interface{}, error) {
	var guru []models.Guru
	s.db.Find(&guru)

	var jadwalKelas []models.JadwalKelas
	s.db.Where("jadwal_semester_id = ? AND is_active = ?", jsID, true).Find(&jadwalKelas)
	jkIDs := make([]uuid.UUID, len(jadwalKelas))
	for i, jk := range jadwalKelas {
		jkIDs[i] = jk.ID
	}

	hasil := make([]map[string]interface{}, 0, len(guru))
	for _, g := range guru {
		var jumlahSlot int64
		if len(jkIDs) > 0 {
			s.db.Model(&models.SlotJadwal{}).
				Where("jadwal_kelas_id IN ? AND guru_id = ?", jkIDs, g.ID).Count(&jumlahSlot)
		}

		var hariLibur []string
		s.db.Model(&models.HariLiburGuru{}).Select("h.nama").
			Joins("JOIN hari h ON h.id = hari_libur_guru.hari_id").
			Where("hari_libur_guru.guru_id = ?", g.ID).
			Pluck("h.nama", &hariLibur)

		hasil = append(hasil, map[string]interface{}{
			"id":                      g.ID,
			"nama":                    g.NamaLengkap,
			"nip":                     g.NIP,
			"jumlah_slot_ditugaskan":  jumlahSlot,
			"jam_maksimal_per_minggu": g.JamMaksimalPerMinggu,
			"hari_libur":              hariLibur,
		})
	}
	return hasil, nil
}

// ========== Status ==========

func (s *LayananJadwal) TransisiStatus(id uuid.UUID, status string) (*models.JadwalSemester, error) {
	valid := map[string]bool{"penempatan": true, "tinjauan": true, "draf": true}
	if !valid[status] {
		return nil, fmt.Errorf("status '%s' tidak valid", status)
	}
	if err := s.db.Model(&models.JadwalSemester{}).Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return nil, err
	}
	return s.AmbilJadwalSemester(id)
}

func (s *LayananJadwal) Publikasikan(id uuid.UUID) (*models.JadwalSemester, error) {
	ok, msg := s.CekKesiapanAktivasi(id)
	if !ok {
		return nil, fmt.Errorf("gagal publikasi: %s", msg)
	}

	if err := s.db.Model(&models.JadwalSemester{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": "dipublikasikan", "bebas_konflik": true}).Error; err != nil {
		return nil, err
	}
	return s.AmbilJadwalSemester(id)
}

func (s *LayananJadwal) BatalkanPublikasi(id uuid.UUID) (*models.JadwalSemester, error) {
	if err := s.db.Model(&models.JadwalSemester{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": "tinjauan", "bebas_konflik": false}).Error; err != nil {
		return nil, err
	}
	return s.AmbilJadwalSemester(id)
}
