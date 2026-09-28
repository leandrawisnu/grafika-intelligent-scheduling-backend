package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/pkg/mlclient"
	"github.com/grafika-scheduling/backend/pkg/storage"
	"github.com/grafika-scheduling/backend/src/dto"
	"github.com/grafika-scheduling/backend/src/models"
	"github.com/grafika-scheduling/backend/src/services"
)

const batasBerkasImpor = 15 * 1024 * 1024

func (h *PengelolaJadwal) PratinjauImpor(c *fiber.Ctx) error {
	js, err := h.jadwalDraf(c)
	if err != nil {
		return err
	}
	if h.objek == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Penyimpanan MinIO belum dikonfigurasi."})
	}
	berkas, err := c.FormFile("berkas")
	if err != nil {
		return berkasDitolak(c)
	}
	ext, ok := ekstensiBerkas(berkas.Filename)
	if !ok || berkas.Size > batasBerkasImpor {
		return berkasDitolak(c)
	}
	src, err := berkas.Open()
	if err != nil {
		return berkasDitolak(c)
	}
	defer src.Close()
	isi, err := io.ReadAll(io.LimitReader(src, batasBerkasImpor+1))
	if err != nil || len(isi) == 0 || len(isi) > batasBerkasImpor {
		return berkasDitolak(c)
	}

	nama := uuid.New().String() + ext
	kunci, err := storage.Key(storage.PrefixJadwal, nama)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal menyimpan berkas ke MinIO."})
	}
	induk := c.UserContext()
	if induk == nil {
		induk = context.Background()
	}
	ctx, batal := context.WithTimeout(induk, 30*time.Second)
	defer batal()
	if err := h.objek.Put(ctx, kunci, bytes.NewReader(isi), int64(len(isi)), tipeBerkas(ext)); err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Gagal menyimpan berkas ke MinIO."})
	}

	katalog, err := h.layananJadwal.KatalogImpor(js.SemesterID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal memuat data master."})
	}
	barisML, err := h.mlClient.EkstrakJadwal(nama, isi, katalogML(katalog))
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": potong("Gagal membaca berkas: "+err.Error(), 300)})
	}
	sudah, err := h.layananJadwal.KunciSlotAktif(js.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Gagal memuat slot yang sudah ada."})
	}
	mentah := make([]services.BarisMentah, len(barisML))
	for i, b := range barisML {
		mentah[i] = services.BarisMentah{
			Hari: b.Hari, Jam: b.Jam, MataPelajaran: b.MataPelajaran,
			Kelas: b.Kelas, Guru: b.Guru, Ruangan: b.Ruangan,
		}
	}
	return c.JSON(fiber.Map{
		"berkas_key": kunci,
		"baris":      services.Cocokkan(mentah, katalog, sudah),
	})
}

func (h *PengelolaJadwal) SimpanImpor(c *fiber.Ctx) error {
	js, err := h.jadwalDraf(c)
	if err != nil {
		return err
	}
	var req dto.SimpanImporRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "format body salah"})
	}
	baris := make([]services.BarisImpor, 0, len(req.Baris))
	for _, b := range req.Baris {
		kelasID, err := uuid.Parse(b.KelasID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "kelas pada baris impor tidak valid"})
		}
		mapelID, err := uuid.Parse(b.MataPelajaranID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "mata pelajaran pada baris impor tidak valid"})
		}
		hariID, err := uuid.Parse(b.HariID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "hari pada baris impor tidak valid"})
		}
		jamID, err := uuid.Parse(b.JamPelajaranID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "jam pada baris impor tidak valid"})
		}
		guruID, err := uuidOpsional(b.GuruID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "guru pada baris impor tidak valid"})
		}
		ruangID, err := uuidOpsional(b.RuanganID)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ruangan pada baris impor tidak valid"})
		}
		baris = append(baris, services.BarisImpor{
			KelasID: kelasID, MataPelajaranID: mapelID, HariID: hariID, JamPelajaranID: jamID,
			GuruID: guruID, RuanganID: ruangID,
		})
	}
	jumlah, err := h.layananJadwal.SimpanImpor(js.ID, js.SemesterID, baris)
	if err != nil {
		status := fiber.StatusBadRequest
		if strings.Contains(err.Error(), "gagal menyimpan slot") {
			status = fiber.StatusInternalServerError
		}
		return c.Status(status).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"jumlah": jumlah})
}

func (h *PengelolaJadwal) jadwalDraf(c *fiber.Ctx) (*models.JadwalSemester, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return nil, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "id jadwal tidak valid"})
	}
	js, err := h.layananJadwal.AmbilJadwalSemester(id)
	if err != nil {
		return nil, c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "jadwal tidak ditemukan"})
	}
	return js, nil
}

func katalogML(k services.KatalogCocok) mlclient.KatalogEkstrak {
	out := mlclient.KatalogEkstrak{}
	for _, h := range k.Hari {
		tambahLabel(&out.Hari, h.Nama, "")
	}
	for _, j := range k.Jam {
		if j.Istirahat {
			continue
		}
		mulai := j.Mulai
		if len(mulai) >= 5 {
			mulai = mulai[:5]
		}
		out.Jam = append(out.Jam, fmt.Sprintf("%d %s", j.JamKe, mulai))
	}
	for _, m := range k.Mapel {
		tambahLabel(&out.MataPelajaran, m.Nama, m.Kode)
	}
	for _, kelas := range k.Kelas {
		tambahLabel(&out.Kelas, kelas.Nama, kelas.Kode)
	}
	for _, g := range k.Guru {
		tambahLabel(&out.Guru, g.Nama, g.Kode)
	}
	for _, r := range k.Ruangan {
		tambahLabel(&out.Ruangan, r.Nama, r.Kode)
	}
	return out
}

func tambahLabel(dst *[]string, nama, kode string) {
	if strings.TrimSpace(nama) != "" {
		*dst = append(*dst, nama)
	}
	if strings.TrimSpace(kode) != "" && !strings.EqualFold(kode, nama) {
		*dst = append(*dst, kode)
	}
}

func uuidOpsional(nilai string) (*uuid.UUID, error) {
	nilai = strings.TrimSpace(nilai)
	if nilai == "" {
		return nil, nil
	}
	id, err := uuid.Parse(nilai)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func ekstensiBerkas(nama string) (string, bool) {
	switch strings.ToLower(filepath.Ext(nama)) {
	case ".pdf":
		return ".pdf", true
	case ".xlsx":
		return ".xlsx", true
	default:
		return "", false
	}
}

func tipeBerkas(ext string) string {
	if ext == ".pdf" {
		return "application/pdf"
	}
	return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}

func berkasDitolak(c *fiber.Ctx) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"error": "Berkas harus PDF atau Excel .xlsx, maksimal 15 MB.",
	})
}

func potong(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
