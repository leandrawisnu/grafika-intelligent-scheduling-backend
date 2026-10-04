package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/pkg/mlclient"
	"github.com/grafika-scheduling/backend/pkg/storage"
	"github.com/grafika-scheduling/backend/src/auth"
	"github.com/grafika-scheduling/backend/src/dto"
	"github.com/grafika-scheduling/backend/src/models"
	"github.com/grafika-scheduling/backend/src/services"
	"gorm.io/gorm"
)

const batasBerkasImporAI = 15 * 1024 * 1024

type PengelolaImporAI struct {
	db          *gorm.DB
	layanan     *services.LayananImporAI
	objek       *storage.Client
	rateLimiter fiber.Handler
}

func NewPengelolaImporAI(db *gorm.DB, mlClient *mlclient.Client, objek *storage.Client, rateLimiter fiber.Handler) *PengelolaImporAI {
	if rateLimiter == nil {
		// Pass-through when the limiter is disabled.
		rateLimiter = func(c *fiber.Ctx) error { return c.Next() }
	}
	return &PengelolaImporAI{
		db:          db,
		layanan:     services.NewLayananImporAI(db, mlClient, objek),
		objek:       objek,
		rateLimiter: rateLimiter,
	}
}

func (h *PengelolaImporAI) DaftarkanRute(r fiber.Router) {
	// Unggah/analisis, penerapan, dan ulangi = endpoint berat (parse LLM +
	// transaksi); polling status sengaja tidak dibatasi agar UI bisa
	// memantau progres tanpa terkena 429.
	r.Post("/dokumen-impor", h.rateLimiter, h.Buat)
	r.Get("/dokumen-impor/:id", h.Ambil)
	r.Post("/dokumen-impor/:id/terapkan", h.rateLimiter, h.Terapkan)
	r.Post("/dokumen-impor/:id/ulangi", h.rateLimiter, h.Ulangi)
}

func (h *PengelolaImporAI) Buat(c *fiber.Ctx) error {
	if h.objek == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Penyimpanan MinIO belum dikonfigurasi."})
	}

	berkas, err := c.FormFile("berkas")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Berkas wajib diunggah."})
	}
	if berkas.Size <= 0 || berkas.Size > batasBerkasImporAI {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Berkas maksimal 15 MB."})
	}

	target := strings.TrimSpace(c.FormValue("target"))
	if target == "" {
		target = "otomatis"
	}
	if target != "otomatis" && target != "jadwal" && target != "master" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "target harus otomatis, jadwal, atau master."})
	}

	var jadwalID *uuid.UUID
	var semesterID *uuid.UUID
	if raw := strings.TrimSpace(c.FormValue("jadwal_semester_id")); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "jadwal_semester_id tidak valid."})
		}
		var js models.JadwalSemester
		if err := h.db.Select("id", "semester_id").First(&js, "id = ?", id).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "jadwal semester tidak ditemukan."})
		}
		jadwalID = &id
		semesterID = &js.SemesterID
	}

	sumber, err := berkas.Open()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Gagal membaca berkas."})
	}
	defer sumber.Close()
	isi, err := io.ReadAll(io.LimitReader(sumber, batasBerkasImporAI+1))
	if err != nil || len(isi) == 0 || len(isi) > batasBerkasImporAI {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Berkas maksimal 15 MB."})
	}

	nama := bersihkanNamaBerkas(berkas.Filename)
	ext := strings.ToLower(filepath.Ext(nama))
	if len(ext) > 10 {
		ext = ""
	}
	kunci, err := storage.Key(storage.PrefixJadwal, uuid.NewString()+ext)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyiapkan penyimpanan berkas."})
	}
	ctx, batal := context.WithTimeout(c.UserContext(), 30*time.Second)
	defer batal()
	if err := h.objek.Put(ctx, kunci, bytes.NewReader(isi), int64(len(isi)), berkas.Header.Get("Content-Type")); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Gagal menyimpan berkas ke MinIO."})
	}

	akun, _ := auth.Dari(c)
	job := models.DokumenImpor{
		JadwalSemesterID: jadwalID,
		SemesterID:       semesterID,
		Target:           target,
		NamaBerkas:       nama,
		KunciBerkas:      kunci,
		MimeBerkas:       berkas.Header.Get("Content-Type"),
		UkuranBerkas:     int64(len(isi)),
		Status:           "memproses",
		Tahap:            "menunggu",
		DilakukanOleh:    akun.Email,
	}
	if err := h.db.Create(&job).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal membuat job impor."})
	}
	go h.layanan.JalankanAnalisis(job.ID)
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"id": job.ID})
}

func (h *PengelolaImporAI) Ambil(c *fiber.Ctx) error {
	job, errResp := h.jobDenganAkses(c)
	if errResp != nil {
		return errResp
	}
	resp := fiber.Map{
		"id":                 job.ID,
		"status":             job.Status,
		"tahap":              job.Tahap,
		"target":             job.Target,
		"nama_berkas":        job.NamaBerkas,
		"jadwal_semester_id": job.JadwalSemesterID,
		"dilakukan_oleh":     job.DilakukanOleh,
		"created_at":         job.CreatedAt,
		"updated_at":         job.UpdatedAt,
	}
	if job.Pesan != nil {
		resp["pesan"] = *job.Pesan
	}
	if job.RencanaJSON != nil {
		resp["rencana"] = json.RawMessage(*job.RencanaJSON)
	}
	if job.HasilTerapkanJSON != nil {
		resp["hasil_terapkan"] = json.RawMessage(*job.HasilTerapkanJSON)
	}
	return c.JSON(resp)
}

func (h *PengelolaImporAI) Terapkan(c *fiber.Ctx) error {
	job, errResp := h.jobDenganAkses(c)
	if errResp != nil {
		return errResp
	}
	var req dto.TerapkanImporRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "format body salah"})
	}
	hasil, err := h.layanan.TerapkanRencana(job, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(hasil)
}

func (h *PengelolaImporAI) Ulangi(c *fiber.Ctx) error {
	job, errResp := h.jobDenganAkses(c)
	if errResp != nil {
		return errResp
	}
	if job.Status != "gagal" && job.Status != "siap" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Hanya dokumen berstatus gagal atau siap yang bisa dianalisis ulang."})
	}
	if err := h.db.Model(&models.DokumenImpor{}).Where("id = ?", job.ID).Updates(map[string]any{
		"status": "memproses", "tahap": "menunggu", "pesan": nil, "rencana_json": nil,
	}).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyiapkan ulang job impor."})
	}
	go h.layanan.JalankanAnalisis(job.ID)
	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{"id": job.ID})
}

func (h *PengelolaImporAI) jobDenganAkses(c *fiber.Ctx) (*models.DokumenImpor, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id tidak valid"})
	}
	var job models.DokumenImpor
	if err := h.db.First(&job, "id = ?", id).Error; err != nil {
		return nil, c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "dokumen impor tidak ditemukan"})
	}
	if !h.bolehLihat(c, &job) {
		return nil, c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "dokumen impor tidak ditemukan"})
	}
	return &job, nil
}

// bolehLihat: koor_jurusan hanya boleh melihat/menerapkan job untuk jurusannya.
func (h *PengelolaImporAI) bolehLihat(c *fiber.Ctx, job *models.DokumenImpor) bool {
	if _, terbatas := auth.Terbatas(c); !terbatas {
		return true
	}
	if job.JadwalSemesterID == nil {
		return false
	}
	kode, ok := auth.KodeJurusanPengguna(h.db, c)
	if !ok {
		return false
	}
	var jumlah int64
	h.db.Table("jadwal_semester_jurusan AS jsj").
		Joins("JOIN jurusan j ON j.id = jsj.jurusan_id").
		Where("jsj.jadwal_semester_id = ? AND j.kode = ?", *job.JadwalSemesterID, kode).
		Count(&jumlah)
	return jumlah > 0
}

func bersihkanNamaBerkas(nama string) string {
	nama = filepath.Base(strings.TrimSpace(nama))
	if nama == "." || nama == "/" || nama == "" {
		nama = "dokumen"
	}
	if len(nama) > 180 {
		nama = nama[len(nama)-180:]
	}
	return nama
}
