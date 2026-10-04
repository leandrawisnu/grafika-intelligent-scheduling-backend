package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/src/models"
	"gorm.io/gorm"
)

type BarisImpor struct {
	KelasID         uuid.UUID
	MataPelajaranID uuid.UUID
	HariID          uuid.UUID
	JamPelajaranID  uuid.UUID
	RuanganID       *uuid.UUID
	GuruID          *uuid.UUID
}

type slotImpor struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	JadwalKelasID   uuid.UUID  `gorm:"column:jadwal_kelas_id"`
	KelasID         uuid.UUID  `gorm:"column:kelas_id"`
	MataPelajaranID uuid.UUID  `gorm:"column:mata_pelajaran_id"`
	HariID          uuid.UUID  `gorm:"column:hari_id"`
	JamPelajaranID  uuid.UUID  `gorm:"column:jam_pelajaran_id"`
	RuanganID       *uuid.UUID `gorm:"column:ruangan_id"`
	GuruID          *uuid.UUID `gorm:"column:guru_id"`
	MingguKe        int16      `gorm:"column:minggu_ke"`
	Terkunci        bool       `gorm:"column:terkunci"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (slotImpor) TableName() string { return "slot_jadwal" }

func (s *LayananJadwal) KatalogImpor(semesterID uuid.UUID) (KatalogCocok, error) {
	var hari []models.Hari
	if err := s.db.Order("urutan_hari").Find(&hari).Error; err != nil {
		return KatalogCocok{}, err
	}
	var jam []models.JamPelajaran
	if err := s.db.Order("jam_ke").Find(&jam).Error; err != nil {
		return KatalogCocok{}, err
	}
	var mapel []models.MataPelajaran
	if err := s.db.Find(&mapel).Error; err != nil {
		return KatalogCocok{}, err
	}
	var kelas []models.Kelas
	if err := s.db.Where("semester_id = ?", semesterID).Find(&kelas).Error; err != nil {
		return KatalogCocok{}, err
	}
	var guru []models.Guru
	if err := s.db.Find(&guru).Error; err != nil {
		return KatalogCocok{}, err
	}
	var ruang []models.Ruangan
	if err := s.db.Where("semester_id = ?", semesterID).Find(&ruang).Error; err != nil {
		return KatalogCocok{}, err
	}

	kat := KatalogCocok{}
	for _, h := range hari {
		kat.Hari = append(kat.Hari, EntriCocok{ID: h.ID.String(), Nama: h.Nama})
	}
	for _, j := range jam {
		kat.Jam = append(kat.Jam, JamCocok{
			ID: j.ID.String(), JamKe: int(j.JamKe), Mulai: j.WaktuMulai, Istirahat: j.Istirahat,
		})
	}
	for _, m := range mapel {
		kat.Mapel = append(kat.Mapel, EntriCocok{ID: m.ID.String(), Nama: m.Nama, Kode: m.Kode})
	}
	for _, k := range kelas {
		kat.Kelas = append(kat.Kelas, EntriCocok{ID: k.ID.String(), Nama: k.Nama, Kode: k.Kode})
	}
	for _, g := range guru {
		kat.Guru = append(kat.Guru, EntriCocok{ID: g.ID.String(), Nama: g.NamaLengkap, Kode: g.NIP})
	}
	for _, r := range ruang {
		kat.Ruangan = append(kat.Ruangan, EntriCocok{ID: r.ID.String(), Nama: r.Nama, Kode: r.Kode})
	}
	return kat, nil
}

func (s *LayananJadwal) KunciSlotAktif(jsID uuid.UUID) (map[string]struct{}, error) {
	var rows []struct {
		KelasID uuid.UUID `gorm:"column:kelas_id"`
		HariID  uuid.UUID `gorm:"column:hari_id"`
		JamID   uuid.UUID `gorm:"column:jam_pelajaran_id"`
	}
	err := s.db.Table("slot_jadwal").
		Select("slot_jadwal.kelas_id, slot_jadwal.hari_id, slot_jadwal.jam_pelajaran_id").
		Joins("JOIN jadwal_kelas ON jadwal_kelas.id = slot_jadwal.jadwal_kelas_id").
		Where("jadwal_kelas.jadwal_semester_id = ? AND jadwal_kelas.is_active = ?", jsID, true).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	kunci := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		kunci[kunciSlot(row.KelasID.String(), row.HariID.String(), row.JamID.String())] = struct{}{}
	}
	return kunci, nil
}

func (s *LayananJadwal) SimpanImpor(jsID, semesterID uuid.UUID, baris []BarisImpor) (int, error) {
	if len(baris) == 0 {
		return 0, fmt.Errorf("tidak ada baris untuk disimpan")
	}
	var jumlah int
	err := s.db.Transaction(func(tx *gorm.DB) error {
		dalam := &LayananJadwal{db: tx}
		jadwalKelas := map[uuid.UUID]uuid.UUID{}
		terpakai := map[string]struct{}{}
		ada, err := dalam.KunciSlotAktif(jsID)
		if err != nil {
			return err
		}
		for k := range ada {
			terpakai[k] = struct{}{}
		}
		sekarang := time.Now()
		for _, b := range baris {
			if b.KelasID == uuid.Nil || b.MataPelajaranID == uuid.Nil || b.HariID == uuid.Nil || b.JamPelajaranID == uuid.Nil {
				return fmt.Errorf("baris impor belum lengkap")
			}
			kunci := kunciSlot(b.KelasID.String(), b.HariID.String(), b.JamPelajaranID.String())
			if _, bentrok := terpakai[kunci]; bentrok {
				continue
			}
			if err := pastikanReferensi(tx, semesterID, b); err != nil {
				return err
			}
			jkID, ok := jadwalKelas[b.KelasID]
			if !ok {
				jkID, err = dalam.pastikanJadwalKelasAktif(jsID, b.KelasID)
				if err != nil {
					return err
				}
				jadwalKelas[b.KelasID] = jkID
			}
			slot := slotImpor{
				ID:              uuid.New(),
				JadwalKelasID:   jkID,
				KelasID:         b.KelasID,
				MataPelajaranID: b.MataPelajaranID,
				HariID:          b.HariID,
				JamPelajaranID:  b.JamPelajaranID,
				RuanganID:       b.RuanganID,
				GuruID:          b.GuruID,
				MingguKe:        1,
				CreatedAt:       sekarang,
				UpdatedAt:       sekarang,
			}
			if err := tx.Create(&slot).Error; err != nil {
				return fmt.Errorf("gagal menyimpan slot impor: %w", err)
			}
			terpakai[kunci] = struct{}{}
			jumlah++
		}
		if jumlah == 0 {
			return fmt.Errorf("tidak ada baris baru untuk disimpan")
		}
		if err := dalam.TandaiPerluValidasi(jsID); err != nil {
			return fmt.Errorf("gagal tandai jadwal perlu validasi: %w", err)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return jumlah, nil
}

func (s *LayananJadwal) pastikanJadwalKelasAktif(jsID, kelasID uuid.UUID) (uuid.UUID, error) {
	var jk models.JadwalKelas
	err := s.db.Where("jadwal_semester_id = ? AND kelas_id = ? AND is_active = ?", jsID, kelasID, true).
		First(&jk).Error
	if err == nil {
		return jk.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, err
	}
	dibuat, err := s.BuatJadwalKelas(jsID, kelasID)
	if err != nil {
		return uuid.Nil, err
	}
	return dibuat.ID, nil
}

func pastikanReferensi(tx *gorm.DB, semesterID uuid.UUID, b BarisImpor) error {
	var kelas models.Kelas
	if err := tx.First(&kelas, "id = ?", b.KelasID).Error; err != nil {
		return fmt.Errorf("kelas tidak ditemukan")
	}
	if kelas.SemesterID != semesterID {
		return fmt.Errorf("kelas %s tidak termasuk semester jadwal ini", kelas.Nama)
	}
	if err := tx.First(&models.MataPelajaran{}, "id = ?", b.MataPelajaranID).Error; err != nil {
		return fmt.Errorf("mata pelajaran tidak ditemukan")
	}
	if err := tx.First(&models.Hari{}, "id = ?", b.HariID).Error; err != nil {
		return fmt.Errorf("hari tidak ditemukan")
	}
	var jam models.JamPelajaran
	if err := tx.First(&jam, "id = ?", b.JamPelajaranID).Error; err != nil {
		return fmt.Errorf("jam pelajaran tidak ditemukan")
	}
	if jam.Istirahat {
		return fmt.Errorf("jam istirahat tidak dapat diimpor")
	}
	if b.GuruID != nil {
		if err := tx.First(&models.Guru{}, "id = ?", *b.GuruID).Error; err != nil {
			return fmt.Errorf("guru tidak ditemukan")
		}
	}
	if b.RuanganID != nil {
		var ruangan models.Ruangan
		if err := tx.First(&ruangan, "id = ?", *b.RuanganID).Error; err != nil {
			return fmt.Errorf("ruangan tidak ditemukan")
		}
		if ruangan.SemesterID != semesterID {
			return fmt.Errorf("ruangan %s tidak termasuk semester jadwal ini", ruangan.Nama)
		}
	}
	return nil
}
