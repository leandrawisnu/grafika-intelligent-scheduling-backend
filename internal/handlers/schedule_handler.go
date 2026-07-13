package handlers

import (
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/internal/dto"
	"github.com/grafika-scheduling/backend/internal/models"
	"github.com/grafika-scheduling/backend/internal/services"
	"github.com/grafika-scheduling/backend/pkg/mlclient"
	"gorm.io/gorm"
)

type PengelolaJadwal struct {
	layananJadwal  *services.LayananJadwal
	layananKonflik *services.LayananKonflik
	mlClient       *mlclient.Client
	db             *gorm.DB
}

func NewPengelolaJadwal(db *gorm.DB, mlClient *mlclient.Client) *PengelolaJadwal {
	return &PengelolaJadwal{
		layananJadwal:  services.NewLayananJadwal(db),
		layananKonflik: services.NewLayananKonflik(db),
		mlClient:       mlClient,
		db:             db,
	}
}

func (h *PengelolaJadwal) DaftarkanRute(r fiber.Router) {
	// Jadwal Semester
	r.Post("/jadwal-semester", h.BuatJadwalSemester)
	r.Get("/jadwal-semester", h.DaftarJadwalSemester)
	r.Get("/jadwal-semester/:id", h.AmbilJadwalSemester)
	r.Put("/jadwal-semester/:id/status", h.TransisiStatus)
	r.Post("/jadwal-semester/:id/publikasi", h.Publikasi)
	r.Post("/jadwal-semester/:id/batalkan-publikasi", h.BatalkanPublikasi)

	// Jurusan dalam Jadwal Semester
	r.Post("/jadwal-semester/:id/jurusan", h.TambahJurusan)
	r.Delete("/jadwal-semester/:id/jurusan/:jurusanId", h.HapusJurusan)

	// Cek kesiapan aktivasi
	r.Get("/jadwal-semester/:id/kesiapan", h.CekKesiapan)

	// Jadwal Kelas
	r.Post("/jadwal-semester/:id/jadwal-kelas", h.BuatJadwalKelas)
	r.Get("/jadwal-kelas/:id", h.AmbilJadwalKelas)
	r.Get("/jadwal-semester/:id/jadwal-kelas-aktif", h.SemuaJadwalKelasAktif)

	// Slot
	r.Post("/jadwal-kelas/:id/slot", h.TambahSlot)
	r.Post("/jadwal-kelas/:id/slot/massal", h.TambahSlotMassal)
	r.Put("/slot/:slotId", h.PerbaruiSlot)
	r.Delete("/slot/:slotId", h.HapusSlot)

	// Penempatan Guru
	r.Get("/jadwal-semester/:id/slot-belum-diplot", h.SlotBelumDiplot)
	r.Put("/slot/:slotId/tugaskan-guru", h.TugaskanGuru)
	r.Post("/jadwal-semester/:id/tugaskan-guru/massal", h.TugaskanMassal)
	r.Get("/jadwal-semester/:id/ketersediaan-guru", h.KetersediaanGuru)

	// Konflik
	r.Get("/jadwal-semester/:id/konflik", h.DaftarKonflik)
	r.Post("/jadwal-semester/:id/validasi", h.Validasi)
	r.Post("/jadwal-semester/:id/prediksi-konflik", h.PrediksiKonflik)

	// AI
	r.Post("/konflik/:id/selesaikan", h.SelesaikanKonflik)
	r.Get("/konflik/:id/resolusi", h.DaftarResolusi)
	r.Post("/konflik/:id/jelaskan", h.JelaskanKonflik)
	r.Post("/resolusi/:id/terima", h.TerimaResolusi)
	r.Post("/ai/tanya", h.AITanya)
}

// ========== Jadwal Semester ==========

func (h *PengelolaJadwal) BuatJadwalSemester(c *fiber.Ctx) error {
	var req dto.BuatJadwalSemesterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	semID, _ := uuid.Parse(req.SemesterID)
	js, err := h.layananJadwal.BuatJadwalSemester(semID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(js)
}

func (h *PengelolaJadwal) DaftarJadwalSemester(c *fiber.Ctx) error {
	list, err := h.layananJadwal.DaftarJadwalSemester()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(list)
}

func (h *PengelolaJadwal) AmbilJadwalSemester(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	js, err := h.layananJadwal.AmbilJadwalSemester(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "tidak ditemukan"})
	}
	return c.JSON(js)
}

func (h *PengelolaJadwal) TransisiStatus(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var req dto.TransisiStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	js, err := h.layananJadwal.TransisiStatus(id, req.Status)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(js)
}

func (h *PengelolaJadwal) Publikasi(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	js, err := h.layananJadwal.Publikasikan(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(js)
}

func (h *PengelolaJadwal) BatalkanPublikasi(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	js, err := h.layananJadwal.BatalkanPublikasi(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(js)
}

// ========== Jurusan dalam Jadwal Semester ==========

func (h *PengelolaJadwal) TambahJurusan(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var req dto.TambahJurusanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	if err := h.layananJadwal.TambahJurusanKeSemester(id, req.JurusanIDs); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	js, _ := h.layananJadwal.AmbilJadwalSemester(id)
	return c.JSON(js)
}

func (h *PengelolaJadwal) HapusJurusan(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	jurID, _ := uuid.Parse(c.Params("jurusanId"))
	if err := h.layananJadwal.HapusJurusanDariSemester(id, jurID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "dihapus"})
}

func (h *PengelolaJadwal) CekKesiapan(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	ok, msg := h.layananJadwal.CekKesiapanAktivasi(id)
	return c.JSON(fiber.Map{"siap": ok, "pesan": msg})
}

// ========== Jadwal Kelas ==========

func (h *PengelolaJadwal) BuatJadwalKelas(c *fiber.Ctx) error {
	jsID, _ := uuid.Parse(c.Params("id"))
	var req dto.BuatJadwalKelasRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	kelasID, _ := uuid.Parse(req.KelasID)
	jk, err := h.layananJadwal.BuatJadwalKelas(jsID, kelasID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(jk)
}

func (h *PengelolaJadwal) AmbilJadwalKelas(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	jk, err := h.layananJadwal.AmbilJadwalKelas(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "tidak ditemukan"})
	}
	return c.JSON(jk)
}

func (h *PengelolaJadwal) SemuaJadwalKelasAktif(c *fiber.Ctx) error {
	jsID, _ := uuid.Parse(c.Params("id"))
	list, err := h.layananJadwal.AmbilSemuaJadwalKelasAktif(jsID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(list)
}

// ========== Slot ==========

func (h *PengelolaJadwal) TambahSlot(c *fiber.Ctx) error {
	jkID, _ := uuid.Parse(c.Params("id"))
	var slot models.SlotJadwal
	if err := c.BodyParser(&slot); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	hasil, err := h.layananJadwal.TambahSlot(jkID, &slot)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(hasil)
}

func (h *PengelolaJadwal) TambahSlotMassal(c *fiber.Ctx) error {
	jkID, _ := uuid.Parse(c.Params("id"))
	var req dto.SlotMassalRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}

	slots := make([]models.SlotJadwal, len(req.Slots))
	for i, s := range req.Slots {
		kelasID, _ := uuid.Parse(s.KelasID)
		mapelID, _ := uuid.Parse(s.MataPelajaranID)
		hariID, _ := uuid.Parse(s.HariID)
		jamID, _ := uuid.Parse(s.JamPelajaranID)
		ruangID := uuid.Nil
		guruID := uuid.Nil
		if s.RuanganID != "" {
			ruangID, _ = uuid.Parse(s.RuanganID)
		}
		if s.GuruID != "" {
			guruID, _ = uuid.Parse(s.GuruID)
		}
		slots[i] = models.SlotJadwal{
			KelasID:         kelasID,
			MataPelajaranID: mapelID,
			HariID:          hariID,
			JamPelajaranID:  jamID,
			RuanganID:       ruangID,
			GuruID:          guruID,
			MingguKe:        s.MingguKe,
			Terkunci:        s.Terkunci,
		}
	}

	hasil, err := h.layananJadwal.TambahSlotBulk(jkID, slots)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(hasil)
}

func (h *PengelolaJadwal) PerbaruiSlot(c *fiber.Ctx) error {
	slotID, _ := uuid.Parse(c.Params("slotId"))
	var perbaruan map[string]interface{}
	if err := c.BodyParser(&perbaruan); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	slot, err := h.layananJadwal.PerbaruiSlot(slotID, perbaruan)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(slot)
}

func (h *PengelolaJadwal) HapusSlot(c *fiber.Ctx) error {
	slotID, _ := uuid.Parse(c.Params("slotId"))
	if err := h.layananJadwal.HapusSlot(slotID); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "dihapus"})
}

// ========== Penempatan Guru ==========

func (h *PengelolaJadwal) SlotBelumDiplot(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	slots, err := h.layananJadwal.AmbilSlotBelumDiplot(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(slots)
}

func (h *PengelolaJadwal) TugaskanGuru(c *fiber.Ctx) error {
	slotID, _ := uuid.Parse(c.Params("slotId"))
	var req dto.TugaskanGuruRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	guruID, _ := uuid.Parse(req.GuruID)
	slot, err := h.layananJadwal.TugaskanGuru(slotID, guruID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(slot)
}

func (h *PengelolaJadwal) TugaskanMassal(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var req dto.TugaskanBulkRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	tugas := make(map[string]string, len(req.Tugas))
	for _, t := range req.Tugas {
		tugas[t.SlotID] = t.GuruID
	}
	if err := h.layananJadwal.TugaskanBulk(id, tugas); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"status": "ditugaskan"})
}

func (h *PengelolaJadwal) KetersediaanGuru(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	avail, err := h.layananJadwal.AmbilKetersediaanGuru(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(avail)
}

// ========== Konflik ==========

func (h *PengelolaJadwal) DaftarKonflik(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var konflik []models.Konflik
	h.db.Where("jadwal_semester_id = ?", id).Order("terdeteksi_pada DESC").Find(&konflik)
	return c.JSON(konflik)
}

func (h *PengelolaJadwal) Validasi(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	konflik, err := h.layananKonflik.DeteksiKonflik(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	ringkasan := make([]dto.RingkasanKonflik, len(konflik))
	for i, k := range konflik {
		ringkasan[i] = dto.RingkasanKonflik{
			ID:            k.ID.String(),
			Tipe:          k.TipeKonflik,
			Keparahan:     k.TingkatKeparahan,
			Deskripsi:     k.Deskripsi,
			Terselesaikan: k.Terselesaikan,
			Terdeteksi:    k.TerdeteksiPada.Format(time.RFC3339),
		}
	}
	return c.JSON(dto.HasilValidasi{
		JumlahKonflik: len(konflik),
		Konflik:       ringkasan,
		Bersih:        len(konflik) == 0,
	})
}

func (h *PengelolaJadwal) PrediksiKonflik(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))

	jadwalKelas, err := h.layananJadwal.AmbilSemuaJadwalKelasAktif(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var slots []dto.SlotUntukML
	for _, jk := range jadwalKelas {
		for _, slot := range jk.SlotJadwal {
			namaKelas := ""
			namaMapel := ""
			namaHari := ""
			jam := ""
			namaRuang := ""
			namaGuru := ""
			if slot.Kelas != nil { namaKelas = slot.Kelas.Nama }
			if slot.MataPelajaran != nil { namaMapel = slot.MataPelajaran.Nama }
			if slot.Hari != nil { namaHari = slot.Hari.Nama }
			if slot.JamPelajaran != nil { jam = slot.JamPelajaran.WaktuMulai }
			if slot.Ruangan != nil { namaRuang = slot.Ruangan.Nama }
			if slot.Guru != nil { namaGuru = slot.Guru.NamaLengkap }

			slots = append(slots, dto.SlotUntukML{
				ID:       slot.ID.String(),
				Kelas:    namaKelas,
				Mapel:    namaMapel,
				Hari:     namaHari,
				Jam:      jam,
				Ruangan:  namaRuang,
				Guru:     namaGuru,
				MingguKe: int(slot.MingguKe),
			})
		}
	}

	var guru []models.Guru
	h.db.Find(&guru)
	guruML := make([]dto.GuruUntukML, len(guru))
	for i, g := range guru {
		var hariLibur []string
		h.db.Model(&models.HariLiburGuru{}).Select("h.nama").
			Joins("JOIN hari h ON h.id = hari_libur_guru.hari_id").
			Where("guru_id = ?", g.ID).Pluck("h.nama", &hariLibur)
		var mapel []string
		h.db.Model(&models.KualifikasiGuru{}).Select("mp.nama").
			Joins("JOIN mata_pelajaran mp ON mp.id = kualifikasi_guru.mata_pelajaran_id").
			Where("guru_id = ?", g.ID).Pluck("mp.nama", &mapel)
		guruML[i] = dto.GuruUntukML{
			ID:              g.ID.String(),
			Nama:            g.NamaLengkap,
			JamMaksPerMinggu: g.JamMaksimalPerMinggu,
			HariLibur:       hariLibur,
			MataPelajaran:   mapel,
		}
	}

	req := dto.PrediksiKonflikRequest{
		JadwalSemesterID: id.String(),
		Slots:            slots,
		Guru:             guruML,
	}

	var hasil dto.HasilPrediksiML
	if err := h.mlClient.POST("/predict/conflicts", req, &hasil); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Layanan ML tidak tersedia: " + err.Error()})
	}
	return c.JSON(hasil)
}

// ========== Resolusi AI ==========

func (h *PengelolaJadwal) SelesaikanKonflik(c *fiber.Ctx) error {
	konflikID, _ := uuid.Parse(c.Params("id"))

	var konflik models.Konflik
	if h.db.First(&konflik, "id = ?", konflikID).Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "konflik tidak ditemukan"})
	}

	js, _ := h.layananJadwal.AmbilJadwalSemester(konflik.JadwalSemesterID)

	req := dto.MLResolveRequest{
		KonflikID: konflikID.String(),
		Konteks: mustMarshal(map[string]interface{}{
			"konflik": map[string]interface{}{
				"tipe":      konflik.TipeKonflik,
				"deskripsi": konflik.Deskripsi,
			},
			"jadwal_kelas": js.JadwalKelas,
		}),
	}

	var hasil dto.MLResolveResponse
	if err := h.mlClient.POST("/resolve/conflicts", req, &hasil); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Layanan ML tidak tersedia: " + err.Error()})
	}

	for _, alt := range hasil.Alternatif {
		perubahanJSON, _ := json.Marshal(alt.Perubahan)
		h.db.Create(&models.ResolusiAI{
			KonflikID:           konflikID,
			JadwalSemesterID:    konflik.JadwalSemesterID,
			Peringkat:           int16(alt.Peringkat),
			SkorKeyakinan:       alt.Keyakinan,
			UsulanPerubahanJSON: string(perubahanJSON),
			Penjelasan:          alt.Penjelasan,
		})
	}

	return c.JSON(hasil)
}

func (h *PengelolaJadwal) DaftarResolusi(c *fiber.Ctx) error {
	konflikID, _ := uuid.Parse(c.Params("id"))
	var resolusi []models.ResolusiAI
	h.db.Where("konflik_id = ?", konflikID).Order("peringkat").Find(&resolusi)
	return c.JSON(resolusi)
}

func (h *PengelolaJadwal) TerimaResolusi(c *fiber.Ctx) error {
	resolusiID, _ := uuid.Parse(c.Params("id"))

	var resolusi models.ResolusiAI
	if h.db.First(&resolusi, "id = ?", resolusiID).Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "resolusi tidak ditemukan"})
	}

	var perubahan []map[string]interface{}
	json.Unmarshal([]byte(resolusi.UsulanPerubahanJSON), &perubahan)

	for _, ubah := range perubahan {
		aksi, _ := ubah["action"].(string)
		slotIDStr, _ := ubah["slot_id"].(string)
		slotID, _ := uuid.Parse(slotIDStr)

		switch aksi {
		case "reassign_teacher":
			guruIDStr, _ := ubah["new_teacher_id"].(string)
			guruID, _ := uuid.Parse(guruIDStr)
			h.layananJadwal.TugaskanGuru(slotID, guruID)
		case "change_room":
			ruangIDStr, _ := ubah["new_room_id"].(string)
			ruangID, _ := uuid.Parse(ruangIDStr)
			h.layananJadwal.PerbaruiSlot(slotID, map[string]interface{}{"ruangan_id": ruangID})
		case "change_time_slot":
			jamIDStr, _ := ubah["new_time_slot_id"].(string)
			jamID, _ := uuid.Parse(jamIDStr)
			h.layananJadwal.PerbaruiSlot(slotID, map[string]interface{}{"jam_pelajaran_id": jamID})
		case "swap_teachers":
			slotBIDStr, _ := ubah["swap_with_slot_id"].(string)
			slotBID, _ := uuid.Parse(slotBIDStr)
			var slotA, slotB models.SlotJadwal
			h.db.First(&slotA, "id = ?", slotID)
			h.db.First(&slotB, "id = ?", slotBID)
			h.layananJadwal.TugaskanGuru(slotID, slotB.GuruID)
			h.layananJadwal.TugaskanGuru(slotBID, slotA.GuruID)
		}
	}

	h.db.Model(&resolusi).Update("diterima", true)
	h.db.Model(&models.Konflik{}).Where("id = ?", resolusi.KonflikID).
		Updates(map[string]interface{}{
			"terselesaikan":      true,
			"diselesaikan_oleh": "ai",
			"terselesaikan_pada": time.Now(),
		})

	h.layananKonflik.DeteksiKonflik(resolusi.JadwalSemesterID)

	return c.JSON(fiber.Map{"status": "diterapkan"})
}

func (h *PengelolaJadwal) JelaskanKonflik(c *fiber.Ctx) error {
	konflikID, _ := uuid.Parse(c.Params("id"))

	var konflik models.Konflik
	if h.db.First(&konflik, "id = ?", konflikID).Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "konflik tidak ditemukan"})
	}

	req := map[string]interface{}{
		"conflict_id": konflikID.String(),
		"konteks": map[string]string{
			"tipe":      konflik.TipeKonflik,
			"deskripsi": konflik.Deskripsi,
		},
	}

	var hasil map[string]interface{}
	if err := h.mlClient.POST("/explain/resolution", req, &hasil); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Layanan ML tidak tersedia: " + err.Error()})
	}
	return c.JSON(hasil)
}

func (h *PengelolaJadwal) AITanya(c *fiber.Ctx) error {
	var req dto.AIQueryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}

	reqML := dto.MLQueryRequest{
		Pertanyaan:    req.Pertanyaan,
		JadwalSemesterID: req.JadwalSemesterID,
	}

	var hasil dto.MLQueryResponse
	if err := h.mlClient.POST("/query/nl", reqML, &hasil); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Layanan ML tidak tersedia: " + err.Error()})
	}
	return c.JSON(hasil)
}

func mustMarshal(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return json.RawMessage(data)
}
