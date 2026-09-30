package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/src/models"
)

const batasUsulan = 3

// PerubahanSlot mengikuti aksi yang sudah diterapkan TerimaResolusi.
type PerubahanSlot struct {
	Action        string `json:"action"`
	SlotID        string `json:"slot_id"`
	NewTeacherID  string `json:"new_teacher_id,omitempty"`
	NewRoomID     string `json:"new_room_id,omitempty"`
	NewTimeSlotID string `json:"new_time_slot_id,omitempty"`
}

// UsulanCocok adalah satu perbaikan dari data jadwal yang sudah ada.
type UsulanCocok struct {
	Peringkat  int
	Label      string
	Penjelasan string
	Keyakinan  float64
	Perubahan  []PerubahanSlot
}

type konteksCocok struct {
	slots []models.SlotJadwal
	guru  []models.Guru
	ruang []models.Ruangan
	jam   []models.JamPelajaran
	libur map[string]bool
}

func (s *LayananKonflik) UsulkanPerbaikan(konflik models.Konflik) ([]UsulanCocok, error) {
	var slots []models.SlotJadwal
	if err := s.db.Joins("JOIN jadwal_kelas ON jadwal_kelas.id = slot_jadwal.jadwal_kelas_id").
		Where("jadwal_kelas.jadwal_semester_id = ? AND jadwal_kelas.is_active = ?", konflik.JadwalSemesterID, true).
		Preload("Kelas").Preload("MataPelajaran").Preload("Hari").Preload("JamPelajaran").
		Preload("Ruangan").Preload("Guru").
		Find(&slots).Error; err != nil {
		return nil, fmt.Errorf("gagal memuat slot: %w", err)
	}
	slots = slotUntukUsulan(slots, konflik)

	var guru []models.Guru
	if err := s.db.Where("aktif = ?", true).Order("nama_lengkap").Find(&guru).Error; err != nil {
		return nil, fmt.Errorf("gagal memuat guru: %w", err)
	}
	var ruang []models.Ruangan
	if err := s.db.Where("aktif = ?", true).Order("nama").Find(&ruang).Error; err != nil {
		return nil, fmt.Errorf("gagal memuat ruangan: %w", err)
	}
	var jam []models.JamPelajaran
	if err := s.db.Where("istirahat = ?", false).Order("jam_ke").Find(&jam).Error; err != nil {
		return nil, fmt.Errorf("gagal memuat jam: %w", err)
	}
	var libur []models.HariLiburGuru
	if err := s.db.Find(&libur).Error; err != nil {
		return nil, fmt.Errorf("gagal memuat hari libur guru: %w", err)
	}

	indeksLibur := make(map[string]bool, len(libur))
	for _, h := range libur {
		indeksLibur[h.GuruID.String()+"|"+h.HariID.String()] = true
	}

	return cocokkan(konteksCocok{
		slots: slots,
		guru:  guru,
		ruang: ruang,
		jam:   jam,
		libur: indeksLibur,
	}, konflik), nil
}

// slotUntukUsulan menyisakan slot yang bisa memengaruhi cocokkan: slot konflik
// (termasuk milik GuruID untuk guru_kelebihan_jam) dan slot yang berbagi
// hari+jam, guru, kelas, ruangan, atau mapel dengan slot konflik.
func slotUntukUsulan(semua []models.SlotJadwal, konflik models.Konflik) []models.SlotJadwal {
	inti := map[uuid.UUID]models.SlotJadwal{}
	for _, s := range semua {
		if idSama(s.ID, konflik.SlotAID) || idSama(s.ID, konflik.SlotBID) || idSama(s.GuruID, konflik.GuruID) {
			inti[s.ID] = s
		}
	}
	var out []models.SlotJadwal
	for _, s := range semua {
		if ikutUsulan(s, inti) {
			out = append(out, s)
		}
	}
	return out
}

func ikutUsulan(s models.SlotJadwal, inti map[uuid.UUID]models.SlotJadwal) bool {
	if _, ok := inti[s.ID]; ok {
		return true
	}
	for _, k := range inti {
		switch {
		case kunciSama(s.HariID, k.HariID) && kunciSama(s.JamPelajaranID, k.JamPelajaranID),
			kunciSama(s.GuruID, k.GuruID),
			kunciSama(s.KelasID, k.KelasID),
			kunciSama(s.RuanganID, k.RuanganID),
			kunciSama(s.MataPelajaranID, k.MataPelajaranID):
			return true
		}
	}
	return false
}

func idSama(id uuid.UUID, ptr *uuid.UUID) bool {
	return ptr != nil && kunciSama(id, *ptr)
}

// kunciSama menolak uuid.Nil agar slot tanpa guru/ruangan tidak saling cocok.
func kunciSama(a, b uuid.UUID) bool {
	return a != uuid.Nil && a == b
}

func cocokkan(k konteksCocok, konflik models.Konflik) []UsulanCocok {
	switch konflik.TipeKonflik {
	case "guru_bentrok":
		return k.usulGuru(k.slotDariID(konflik.SlotBID), k.slotDariID(konflik.SlotAID))
	case "guru_hari_libur":
		return k.usulGuru(k.slotDariID(konflik.SlotAID))
	case "guru_kelebihan_jam":
		return k.usulGuru(k.slotGuru(konflik.GuruID)...)
	case "ruangan_bentrok":
		return k.usulRuangan(k.slotDariID(konflik.SlotBID), k.slotDariID(konflik.SlotAID))
	case "kelas_bentrok":
		return k.usulJam(k.slotDariID(konflik.SlotBID), k.slotDariID(konflik.SlotAID))
	default:
		return nil
	}
}

func (k konteksCocok) usulGuru(slots ...*models.SlotJadwal) []UsulanCocok {
	var usulan []UsulanCocok
	terpakai := map[string]bool{}
	for _, slot := range slots {
		if slot == nil || slot.Terkunci {
			continue
		}
		for _, g := range k.kandidatGuru(*slot) {
			kunci := slot.ID.String() + "|" + g.ID.String()
			if terpakai[kunci] {
				continue
			}
			terpakai[kunci] = true
			sama := k.ajarMapel(g.ID, slot.MataPelajaranID, slot.ID)
			keyakinan := 0.6
			alasan := fmt.Sprintf("%s kosong pada %s jam ke-%d dan jam minggunya masih muat.", g.NamaLengkap, namaHari(*slot), nomorJam(*slot))
			if sama {
				keyakinan = 1
				alasan = fmt.Sprintf("%s sudah mengajar %s dan kosong pada %s jam ke-%d.", g.NamaLengkap, namaMapel(*slot), namaHari(*slot), nomorJam(*slot))
			}
			usulan = append(usulan, UsulanCocok{
				Peringkat:  len(usulan) + 1,
				Label:      "Pindah ke " + g.NamaLengkap,
				Penjelasan: alasan,
				Keyakinan:  keyakinan,
				Perubahan: []PerubahanSlot{{
					Action:       "reassign_teacher",
					SlotID:       slot.ID.String(),
					NewTeacherID: g.ID.String(),
				}},
			})
			if len(usulan) >= batasUsulan {
				return usulan
			}
		}
	}
	return usulan
}

func (k konteksCocok) usulRuangan(slots ...*models.SlotJadwal) []UsulanCocok {
	var usulan []UsulanCocok
	for _, slot := range slots {
		if slot == nil || slot.Terkunci {
			continue
		}
		for _, r := range k.kandidatRuang(*slot) {
			usulan = append(usulan, UsulanCocok{
				Peringkat:  len(usulan) + 1,
				Label:      "Pindah ke " + r.Nama,
				Penjelasan: fmt.Sprintf("Ruangan %s bertipe sama dan kosong pada %s jam ke-%d.", r.Nama, namaHari(*slot), nomorJam(*slot)),
				Keyakinan:  1,
				Perubahan: []PerubahanSlot{{
					Action:    "change_room",
					SlotID:    slot.ID.String(),
					NewRoomID: r.ID.String(),
				}},
			})
			if len(usulan) >= batasUsulan {
				return usulan
			}
		}
	}
	return usulan
}

func (k konteksCocok) usulJam(slots ...*models.SlotJadwal) []UsulanCocok {
	var usulan []UsulanCocok
	for _, slot := range slots {
		if slot == nil || slot.Terkunci {
			continue
		}
		for _, j := range k.kandidatJam(*slot) {
			usulan = append(usulan, UsulanCocok{
				Peringkat:  len(usulan) + 1,
				Label:      fmt.Sprintf("Pindah ke %s jam ke-%d", namaHari(*slot), j.JamKe),
				Penjelasan: fmt.Sprintf("Kelas, guru, dan ruangan slot ini kosong pada jam ke-%d.", j.JamKe),
				Keyakinan:  1,
				Perubahan: []PerubahanSlot{{
					Action:        "change_time_slot",
					SlotID:        slot.ID.String(),
					NewTimeSlotID: j.ID.String(),
				}},
			})
			if len(usulan) >= batasUsulan {
				return usulan
			}
		}
	}
	return usulan
}

func (k konteksCocok) kandidatGuru(slot models.SlotJadwal) []models.Guru {
	var sama, lain []models.Guru
	for _, g := range k.guru {
		if !k.guruLayak(slot, g) {
			continue
		}
		if k.ajarMapel(g.ID, slot.MataPelajaranID, slot.ID) {
			sama = append(sama, g)
		} else {
			lain = append(lain, g)
		}
	}
	return append(sama, lain...)
}

func (k konteksCocok) guruLayak(slot models.SlotJadwal, g models.Guru) bool {
	if !g.Aktif || g.ID == slot.GuruID {
		return false
	}
	if k.libur[g.ID.String()+"|"+slot.HariID.String()] {
		return false
	}
	if k.sibukGuru(g.ID, slot.HariID, slot.JamPelajaranID, slot.MingguKe, slot.ID) {
		return false
	}
	return k.jamMinggu(g.ID)+1 <= g.JamMaksimalPerMinggu
}

func (k konteksCocok) kandidatRuang(slot models.SlotJadwal) []models.Ruangan {
	tipe := k.tipeRuang(slot)
	var out []models.Ruangan
	for _, r := range k.ruang {
		if !r.Aktif || r.ID == slot.RuanganID {
			continue
		}
		if tipe != "" && r.TipeRuangan != tipe {
			continue
		}
		if k.sibukRuang(r.ID, slot.HariID, slot.JamPelajaranID, slot.MingguKe, slot.ID) {
			continue
		}
		out = append(out, r)
	}
	return out
}

func (k konteksCocok) kandidatJam(slot models.SlotJadwal) []models.JamPelajaran {
	var out []models.JamPelajaran
	for _, j := range k.jam {
		if j.Istirahat || j.ID == slot.JamPelajaranID {
			continue
		}
		if k.sibukKelas(slot.KelasID, slot.HariID, j.ID, slot.MingguKe, slot.ID) {
			continue
		}
		if slot.GuruID != uuid.Nil && k.sibukGuru(slot.GuruID, slot.HariID, j.ID, slot.MingguKe, slot.ID) {
			continue
		}
		if slot.RuanganID != uuid.Nil && k.sibukRuang(slot.RuanganID, slot.HariID, j.ID, slot.MingguKe, slot.ID) {
			continue
		}
		out = append(out, j)
	}
	return out
}

func (k konteksCocok) ajarMapel(guruID, mapelID, kecuali uuid.UUID) bool {
	for _, slot := range k.slots {
		if slot.ID == kecuali || slot.GuruID != guruID {
			continue
		}
		if slot.MataPelajaranID == mapelID {
			return true
		}
	}
	return false
}

func (k konteksCocok) sibukGuru(guruID, hariID, jamID uuid.UUID, minggu int16, kecuali uuid.UUID) bool {
	for _, slot := range k.slots {
		if slot.ID == kecuali || slot.GuruID != guruID {
			continue
		}
		if slot.HariID == hariID && slot.JamPelajaranID == jamID && slot.MingguKe == minggu {
			return true
		}
	}
	return false
}

func (k konteksCocok) sibukRuang(ruangID, hariID, jamID uuid.UUID, minggu int16, kecuali uuid.UUID) bool {
	if ruangID == uuid.Nil {
		return false
	}
	for _, slot := range k.slots {
		if slot.ID == kecuali || slot.RuanganID != ruangID {
			continue
		}
		if slot.HariID == hariID && slot.JamPelajaranID == jamID && slot.MingguKe == minggu {
			return true
		}
	}
	return false
}

func (k konteksCocok) sibukKelas(kelasID, hariID, jamID uuid.UUID, minggu int16, kecuali uuid.UUID) bool {
	for _, slot := range k.slots {
		if slot.ID == kecuali || slot.KelasID != kelasID {
			continue
		}
		if slot.HariID == hariID && slot.JamPelajaranID == jamID && slot.MingguKe == minggu {
			return true
		}
	}
	return false
}

func (k konteksCocok) jamMinggu(guruID uuid.UUID) float64 {
	var n float64
	for _, slot := range k.slots {
		if slot.GuruID == guruID {
			n++
		}
	}
	return n
}

func (k konteksCocok) tipeRuang(slot models.SlotJadwal) string {
	if slot.Ruangan != nil && slot.Ruangan.TipeRuangan != "" {
		return slot.Ruangan.TipeRuangan
	}
	for _, r := range k.ruang {
		if r.ID == slot.RuanganID {
			return r.TipeRuangan
		}
	}
	return ""
}

func (k konteksCocok) slotDariID(id *uuid.UUID) *models.SlotJadwal {
	if id == nil {
		return nil
	}
	for i := range k.slots {
		if k.slots[i].ID == *id {
			return &k.slots[i]
		}
	}
	return nil
}

func (k konteksCocok) slotGuru(id *uuid.UUID) []*models.SlotJadwal {
	if id == nil {
		return nil
	}
	var out []*models.SlotJadwal
	for i := range k.slots {
		if k.slots[i].GuruID == *id && !k.slots[i].Terkunci {
			out = append(out, &k.slots[i])
		}
	}
	return out
}

func namaHari(slot models.SlotJadwal) string {
	if slot.Hari != nil && slot.Hari.Nama != "" {
		return slot.Hari.Nama
	}
	return "hari itu"
}

func namaMapel(slot models.SlotJadwal) string {
	if slot.MataPelajaran != nil && slot.MataPelajaran.Nama != "" {
		return slot.MataPelajaran.Nama
	}
	return "mata pelajaran ini"
}

func nomorJam(slot models.SlotJadwal) int16 {
	if slot.JamPelajaran != nil {
		return slot.JamPelajaran.JamKe
	}
	return 0
}
