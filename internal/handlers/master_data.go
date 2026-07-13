package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/internal/models"
	"gorm.io/gorm"
)

type PengelolaMaster struct {
	db *gorm.DB
}

func NewPengelolaMaster(db *gorm.DB) *PengelolaMaster {
	return &PengelolaMaster{db: db}
}

func (h *PengelolaMaster) DaftarkanRute(r fiber.Router) {
	r.Get("/hari", h.DaftarHari)

	r.Get("/tahun-ajaran", h.DaftarTahunAjaran)
	r.Post("/tahun-ajaran", h.BuatTahunAjaran)
	r.Get("/tahun-ajaran/:id", h.AmbilTahunAjaran)
	r.Put("/tahun-ajaran/:id", h.PerbaruiTahunAjaran)
	r.Delete("/tahun-ajaran/:id", h.HapusTahunAjaran)

	r.Get("/semester", h.DaftarSemester)
	r.Post("/semester", h.BuatSemester)
	r.Get("/semester/:id", h.AmbilSemester)
	r.Put("/semester/:id", h.PerbaruiSemester)
	r.Delete("/semester/:id", h.HapusSemester)

	r.Get("/jurusan", h.DaftarJurusan)
	r.Post("/jurusan", h.BuatJurusan)
	r.Get("/jurusan/:id", h.AmbilJurusan)
	r.Put("/jurusan/:id", h.PerbaruiJurusan)
	r.Delete("/jurusan/:id", h.HapusJurusan)

	r.Get("/guru", h.DaftarGuru)
	r.Post("/guru", h.BuatGuru)
	r.Get("/guru/:id", h.AmbilGuru)
	r.Put("/guru/:id", h.PerbaruiGuru)
	r.Delete("/guru/:id", h.HapusGuru)
	r.Get("/guru/:id/hari-libur", h.DaftarHariLiburGuru)
	r.Post("/guru/:id/hari-libur", h.BuatHariLiburGuru)
	r.Delete("/guru/:id/hari-libur/:liburId", h.HapusHariLiburGuru)
	r.Get("/guru/:id/kualifikasi", h.DaftarKualifikasiGuru)
	r.Post("/guru/:id/kualifikasi", h.BuatKualifikasiGuru)
	r.Delete("/guru/:id/kualifikasi/:kualId", h.HapusKualifikasiGuru)

	r.Get("/mata-pelajaran", h.DaftarMataPelajaran)
	r.Post("/mata-pelajaran", h.BuatMataPelajaran)
	r.Get("/mata-pelajaran/:id", h.AmbilMataPelajaran)
	r.Put("/mata-pelajaran/:id", h.PerbaruiMataPelajaran)
	r.Delete("/mata-pelajaran/:id", h.HapusMataPelajaran)

	r.Get("/kelas", h.DaftarKelas)
	r.Post("/kelas", h.BuatKelas)
	r.Get("/kelas/:id", h.AmbilKelas)
	r.Put("/kelas/:id", h.PerbaruiKelas)
	r.Delete("/kelas/:id", h.HapusKelas)

	r.Get("/ruangan", h.DaftarRuangan)
	r.Post("/ruangan", h.BuatRuangan)
	r.Get("/ruangan/:id", h.AmbilRuangan)
	r.Put("/ruangan/:id", h.PerbaruiRuangan)
	r.Delete("/ruangan/:id", h.HapusRuangan)

	r.Get("/jam-pelajaran", h.DaftarJamPelajaran)
	r.Post("/jam-pelajaran", h.BuatJamPelajaran)
	r.Get("/jam-pelajaran/:id", h.AmbilJamPelajaran)
	r.Put("/jam-pelajaran/:id", h.PerbaruiJamPelajaran)
	r.Delete("/jam-pelajaran/:id", h.HapusJamPelajaran)
}

// Hari
func (h *PengelolaMaster) DaftarHari(c *fiber.Ctx) error {
	var items []models.Hari
	h.db.Order("urutan_hari").Find(&items)
	return c.JSON(items)
}

// Tahun Ajaran
func (h *PengelolaMaster) DaftarTahunAjaran(c *fiber.Ctx) error {
	var items []models.TahunAjaran
	h.db.Order("dibuat_pada DESC").Find(&items)
	return c.JSON(items)
}

func (h *PengelolaMaster) BuatTahunAjaran(c *fiber.Ctx) error {
	var item models.TahunAjaran
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	if err := h.db.Create(&item).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(item)
}

func (h *PengelolaMaster) AmbilTahunAjaran(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.TahunAjaran
	if h.db.First(&item, "id = ?", id).Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "tidak ditemukan"})
	}
	return c.JSON(item)
}

func (h *PengelolaMaster) PerbaruiTahunAjaran(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.TahunAjaran
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	h.db.Model(&models.TahunAjaran{}).Where("id = ?", id).Updates(&item)
	h.db.First(&item, "id = ?", id)
	return c.JSON(item)
}

func (h *PengelolaMaster) HapusTahunAjaran(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	h.db.Delete(&models.TahunAjaran{}, "id = ?", id)
	return c.JSON(fiber.Map{"status": "dihapus"})
}

// Semester
func (h *PengelolaMaster) DaftarSemester(c *fiber.Ctx) error {
	var items []models.Semester
	q := h.db.Order("dibuat_pada DESC").Preload("TahunAjaran")
	if ayID := c.Query("tahun_ajaran_id"); ayID != "" {
		q = q.Where("tahun_ajaran_id = ?", ayID)
	}
	q.Find(&items)
	return c.JSON(items)
}

func (h *PengelolaMaster) BuatSemester(c *fiber.Ctx) error {
	var item models.Semester
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	if err := h.db.Create(&item).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(item)
}

func (h *PengelolaMaster) AmbilSemester(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.Semester
	if h.db.Preload("TahunAjaran").First(&item, "id = ?", id).Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "tidak ditemukan"})
	}
	return c.JSON(item)
}

func (h *PengelolaMaster) PerbaruiSemester(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.Semester
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	h.db.Model(&models.Semester{}).Where("id = ?", id).Updates(&item)
	h.db.Preload("TahunAjaran").First(&item, "id = ?", id)
	return c.JSON(item)
}

func (h *PengelolaMaster) HapusSemester(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	h.db.Delete(&models.Semester{}, "id = ?", id)
	return c.JSON(fiber.Map{"status": "dihapus"})
}

// Jurusan
func (h *PengelolaMaster) DaftarJurusan(c *fiber.Ctx) error {
	var items []models.Jurusan
	h.db.Order("kode").Find(&items)
	return c.JSON(items)
}

func (h *PengelolaMaster) BuatJurusan(c *fiber.Ctx) error {
	var item models.Jurusan
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	h.db.Create(&item)
	return c.Status(201).JSON(item)
}

func (h *PengelolaMaster) AmbilJurusan(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.Jurusan
	if h.db.First(&item, "id = ?", id).Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "tidak ditemukan"})
	}
	return c.JSON(item)
}

func (h *PengelolaMaster) PerbaruiJurusan(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.Jurusan
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	h.db.Model(&models.Jurusan{}).Where("id = ?", id).Updates(&item)
	h.db.First(&item, "id = ?", id)
	return c.JSON(item)
}

func (h *PengelolaMaster) HapusJurusan(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	h.db.Delete(&models.Jurusan{}, "id = ?", id)
	return c.JSON(fiber.Map{"status": "dihapus"})
}

// Guru
func (h *PengelolaMaster) DaftarGuru(c *fiber.Ctx) error {
	var items []models.Guru
	h.db.Order("nama_lengkap").Find(&items)
	return c.JSON(items)
}

func (h *PengelolaMaster) BuatGuru(c *fiber.Ctx) error {
	var item models.Guru
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	h.db.Create(&item)
	return c.Status(201).JSON(item)
}

func (h *PengelolaMaster) AmbilGuru(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.Guru
	if h.db.First(&item, "id = ?", id).Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "tidak ditemukan"})
	}
	return c.JSON(item)
}

func (h *PengelolaMaster) PerbaruiGuru(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.Guru
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	h.db.Model(&models.Guru{}).Where("id = ?", id).Updates(&item)
	h.db.First(&item, "id = ?", id)
	return c.JSON(item)
}

func (h *PengelolaMaster) HapusGuru(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	h.db.Delete(&models.Guru{}, "id = ?", id)
	return c.JSON(fiber.Map{"status": "dihapus"})
}

// Hari Libur Guru
func (h *PengelolaMaster) DaftarHariLiburGuru(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var items []models.HariLiburGuru
	h.db.Where("guru_id = ?", id).Preload("Hari").Find(&items)
	return c.JSON(items)
}

func (h *PengelolaMaster) BuatHariLiburGuru(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.HariLiburGuru
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	item.GuruID = id
	h.db.Create(&item)
	return c.Status(201).JSON(item)
}

func (h *PengelolaMaster) HapusHariLiburGuru(c *fiber.Ctx) error {
	h.db.Delete(&models.HariLiburGuru{}, "id = ?", c.Params("liburId"))
	return c.JSON(fiber.Map{"status": "dihapus"})
}

// Kualifikasi Guru
func (h *PengelolaMaster) DaftarKualifikasiGuru(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var items []models.KualifikasiGuru
	h.db.Where("guru_id = ?", id).Preload("MataPelajaran").Find(&items)
	return c.JSON(items)
}

func (h *PengelolaMaster) BuatKualifikasiGuru(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.KualifikasiGuru
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	item.GuruID = id
	h.db.Create(&item)
	return c.Status(201).JSON(item)
}

func (h *PengelolaMaster) HapusKualifikasiGuru(c *fiber.Ctx) error {
	h.db.Delete(&models.KualifikasiGuru{}, "id = ?", c.Params("kualId"))
	return c.JSON(fiber.Map{"status": "dihapus"})
}

// Mata Pelajaran
func (h *PengelolaMaster) DaftarMataPelajaran(c *fiber.Ctx) error {
	var items []models.MataPelajaran
	h.db.Order("kode").Find(&items)
	return c.JSON(items)
}

func (h *PengelolaMaster) BuatMataPelajaran(c *fiber.Ctx) error {
	var item models.MataPelajaran
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	h.db.Create(&item)
	return c.Status(201).JSON(item)
}

func (h *PengelolaMaster) AmbilMataPelajaran(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.MataPelajaran
	if h.db.Preload("Jurusan").First(&item, "id = ?", id).Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "tidak ditemukan"})
	}
	return c.JSON(item)
}

func (h *PengelolaMaster) PerbaruiMataPelajaran(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.MataPelajaran
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	h.db.Model(&models.MataPelajaran{}).Where("id = ?", id).Updates(&item)
	h.db.Preload("Jurusan").First(&item, "id = ?", id)
	return c.JSON(item)
}

func (h *PengelolaMaster) HapusMataPelajaran(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	h.db.Delete(&models.MataPelajaran{}, "id = ?", id)
	return c.JSON(fiber.Map{"status": "dihapus"})
}

// Kelas
func (h *PengelolaMaster) DaftarKelas(c *fiber.Ctx) error {
	var items []models.Kelas
	q := h.db.Order("kode").Preload("Jurusan").Preload("Semester")
	if semID := c.Query("semester_id"); semID != "" {
		q = q.Where("semester_id = ?", semID)
	}
	q.Find(&items)
	return c.JSON(items)
}

func (h *PengelolaMaster) BuatKelas(c *fiber.Ctx) error {
	var item models.Kelas
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	h.db.Create(&item)
	return c.Status(201).JSON(item)
}

func (h *PengelolaMaster) AmbilKelas(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.Kelas
	if h.db.Preload("Jurusan").Preload("Semester").First(&item, "id = ?", id).Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "tidak ditemukan"})
	}
	return c.JSON(item)
}

func (h *PengelolaMaster) PerbaruiKelas(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.Kelas
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	h.db.Model(&models.Kelas{}).Where("id = ?", id).Updates(&item)
	h.db.Preload("Jurusan").Preload("Semester").First(&item, "id = ?", id)
	return c.JSON(item)
}

func (h *PengelolaMaster) HapusKelas(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	h.db.Delete(&models.Kelas{}, "id = ?", id)
	return c.JSON(fiber.Map{"status": "dihapus"})
}

// Ruangan
func (h *PengelolaMaster) DaftarRuangan(c *fiber.Ctx) error {
	var items []models.Ruangan
	h.db.Order("kode").Find(&items)
	return c.JSON(items)
}

func (h *PengelolaMaster) BuatRuangan(c *fiber.Ctx) error {
	var item models.Ruangan
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	h.db.Create(&item)
	return c.Status(201).JSON(item)
}

func (h *PengelolaMaster) AmbilRuangan(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.Ruangan
	if h.db.First(&item, "id = ?", id).Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "tidak ditemukan"})
	}
	return c.JSON(item)
}

func (h *PengelolaMaster) PerbaruiRuangan(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.Ruangan
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	h.db.Model(&models.Ruangan{}).Where("id = ?", id).Updates(&item)
	h.db.First(&item, "id = ?", id)
	return c.JSON(item)
}

func (h *PengelolaMaster) HapusRuangan(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	h.db.Delete(&models.Ruangan{}, "id = ?", id)
	return c.JSON(fiber.Map{"status": "dihapus"})
}

// Jam Pelajaran
func (h *PengelolaMaster) DaftarJamPelajaran(c *fiber.Ctx) error {
	var items []models.JamPelajaran
	h.db.Order("jam_ke").Find(&items)
	return c.JSON(items)
}

func (h *PengelolaMaster) BuatJamPelajaran(c *fiber.Ctx) error {
	var item models.JamPelajaran
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	h.db.Create(&item)
	return c.Status(201).JSON(item)
}

func (h *PengelolaMaster) AmbilJamPelajaran(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.JamPelajaran
	if h.db.First(&item, "id = ?", id).Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "tidak ditemukan"})
	}
	return c.JSON(item)
}

func (h *PengelolaMaster) PerbaruiJamPelajaran(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	var item models.JamPelajaran
	if err := c.BodyParser(&item); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "format body salah"})
	}
	h.db.Model(&models.JamPelajaran{}).Where("id = ?", id).Updates(&item)
	h.db.First(&item, "id = ?", id)
	return c.JSON(item)
}

func (h *PengelolaMaster) HapusJamPelajaran(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	h.db.Delete(&models.JamPelajaran{}, "id = ?", id)
	return c.JSON(fiber.Map{"status": "dihapus"})
}
