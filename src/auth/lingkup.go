package auth

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/src/models"
	"gorm.io/gorm"
)

var ErrAksesDitolak = errors.New("akses ditolak")

const LocalPengguna = "gis.pengguna"

type Lingkup struct {
	db *gorm.DB
}

func NewLingkup(db *gorm.DB) *Lingkup {
	return &Lingkup{db: db}
}

func (l *Lingkup) Periksa(c *fiber.Ctx, akun models.Pengguna) error {
	if akun.Peran == "admin" {
		return nil
	}
	if akun.Peran != "koor_jurusan" || akun.JurusanID == nil {
		return ErrAksesDitolak
	}
	jurusanID := *akun.JurusanID
	switch GerbangKoor(c.Method(), c.Path()) {
	case GerbangBaca:
		return nil
	case GerbangTolak:
		return ErrAksesDitolak
	case GerbangCekKelas:
		return l.kelas(c, jurusanID)
	case GerbangCekKelasBody:
		return l.kelasDiBody(c, jurusanID)
	case GerbangCekJadwalKelas:
		if menugaskanGuru(c.Body()) {
			return ErrAksesDitolak
		}
		return l.jadwalKelas(c.Params("id"), jurusanID)
	case GerbangCekSlot:
		if menugaskanGuru(c.Body()) {
			return ErrAksesDitolak
		}
		return l.slot(c.Params("slotId"), jurusanID)
	case GerbangCekJurusanBody:
		return jurusanDiBody(c.Body(), jurusanID)
	case GerbangCekJurusanParam:
		id, err := uuid.Parse(c.Params("jurusanId"))
		if err != nil || id != jurusanID {
			return ErrAksesDitolak
		}
		return nil
	case GerbangCekDokumenImpor:
		return l.dokumenImpor(c, jurusanID)
	default:
		return ErrAksesDitolak
	}
}

func Terbatas(c *fiber.Ctx) (uuid.UUID, bool) {
	akun, ok := Dari(c)
	if !ok || akun.Peran != "koor_jurusan" || akun.JurusanID == nil {
		return uuid.Nil, false
	}
	return *akun.JurusanID, true
}

// KodeJurusanPengguna resolve kode jurusan milik pengguna koor.
// Jurusan per-semester (N copy per kode), jadi pemanggil yang butuh filter
// lintas semester harus bandingkan kode, bukan UUID.
func KodeJurusanPengguna(db *gorm.DB, c *fiber.Ctx) (string, bool) {
	akun, ok := Dari(c)
	if !ok || akun.Peran != "koor_jurusan" || akun.JurusanID == nil {
		return "", false
	}
	if akun.Jurusan != nil && akun.Jurusan.Kode != "" {
		return akun.Jurusan.Kode, true
	}
	var j models.Jurusan
	if err := db.Select("kode").First(&j, "id = ?", *akun.JurusanID).Error; err != nil {
		return "", false
	}
	return j.Kode, true
}

func Dari(c *fiber.Ctx) (models.Pengguna, bool) {
	akun, ok := c.Locals(LocalPengguna).(models.Pengguna)
	return akun, ok
}

func SaringSemester(c *fiber.Ctx, js *models.JadwalSemester) {
	kode, ok := kodeJurusanC(c)
	if !ok || js == nil {
		return
	}
	jurusan := make([]models.JadwalSemesterJurusan, 0)
	for _, item := range js.Jurusan {
		if item.Jurusan != nil && item.Jurusan.Kode == kode {
			jurusan = append(jurusan, item)
		}
	}
	js.Jurusan = jurusan
	kelas := make([]models.JadwalKelas, 0)
	for _, item := range js.JadwalKelas {
		if item.Jurusan != nil && item.Jurusan.Kode == kode {
			kelas = append(kelas, item)
		} else if item.Jurusan == nil && item.Kelas != nil && item.Kelas.Jurusan != nil && item.Kelas.Jurusan.Kode == kode {
			kelas = append(kelas, item)
		}
	}
	js.JadwalKelas = kelas
}

// kodeJurusanC resolve kode jurusan milik koor dari pengguna.jurusan_id.
// Jurusan kini per-semester (N copy), jadi perbandingan memakai kode, bukan UUID.
func kodeJurusanC(c *fiber.Ctx) (string, bool) {
	akun, ok := Dari(c)
	if !ok || akun.Peran != "koor_jurusan" || akun.JurusanID == nil {
		return "", false
	}
	if akun.Jurusan != nil && akun.Jurusan.Kode != "" {
		return akun.Jurusan.Kode, true
	}
	return "", false
}

func (l *Lingkup) kodeJurusan(jurusanID uuid.UUID) (string, bool) {
	var j models.Jurusan
	if err := l.db.Select("kode").First(&j, "id = ?", jurusanID).Error; err != nil {
		return "", false
	}
	return j.Kode, true
}

func (l *Lingkup) kelas(c *fiber.Ctx, jurusanID uuid.UUID) error {
	kode, ok := l.kodeJurusan(jurusanID)
	if !ok {
		return ErrAksesDitolak
	}
	if c.Method() == fiber.MethodPost {
		var body struct {
			JurusanID string `json:"jurusan_id"`
		}
		if json.Unmarshal(c.Body(), &body) != nil {
			return ErrAksesDitolak
		}
		var target models.Jurusan
		if err := l.db.Select("kode").First(&target, "id = ?", body.JurusanID).Error; err != nil {
			return ErrAksesDitolak
		}
		if target.Kode != kode {
			return ErrAksesDitolak
		}
		return nil
	}
	var kelas models.Kelas
	if err := l.db.Preload("Jurusan").Select("jurusan_id").First(&kelas, "id = ?", c.Params("id")).Error; err != nil {
		return nil
	}
	if kelas.Jurusan == nil || kelas.Jurusan.Kode != kode {
		return ErrAksesDitolak
	}
	if c.Method() == fiber.MethodPut {
		var body struct {
			JurusanID string `json:"jurusan_id"`
		}
		if json.Unmarshal(c.Body(), &body) == nil && body.JurusanID != "" {
			var target models.Jurusan
			if err := l.db.Select("kode").First(&target, "id = ?", body.JurusanID).Error; err != nil {
				return ErrAksesDitolak
			}
			if target.Kode != kode {
				return ErrAksesDitolak
			}
		}
	}
	return nil
}

func (l *Lingkup) kelasDiBody(c *fiber.Ctx, jurusanID uuid.UUID) error {
	kode, ok := l.kodeJurusan(jurusanID)
	if !ok {
		return ErrAksesDitolak
	}
	var body struct {
		KelasID string `json:"kelas_id"`
	}
	if json.Unmarshal(c.Body(), &body) != nil {
		return ErrAksesDitolak
	}
	var kelas models.Kelas
	if err := l.db.Preload("Jurusan").Select("jurusan_id").First(&kelas, "id = ?", body.KelasID).Error; err != nil {
		return ErrAksesDitolak
	}
	if kelas.Jurusan == nil || kelas.Jurusan.Kode != kode {
		return ErrAksesDitolak
	}
	return nil
}

func (l *Lingkup) jadwalKelas(id string, jurusanID uuid.UUID) error {
	kode, ok := l.kodeJurusan(jurusanID)
	if !ok {
		return ErrAksesDitolak
	}
	var jk models.JadwalKelas
	if err := l.db.Preload("Jurusan").Preload("Kelas.Jurusan").Select("jurusan_id", "kelas_id").First(&jk, "id = ?", id).Error; err != nil {
		return ErrAksesDitolak
	}
	if jk.Jurusan != nil && jk.Jurusan.Kode == kode {
		return nil
	}
	if jk.Kelas != nil && jk.Kelas.Jurusan != nil && jk.Kelas.Jurusan.Kode == kode {
		return nil
	}
	return ErrAksesDitolak
}

func (l *Lingkup) slot(slotID string, jurusanID uuid.UUID) error {
	var slot models.SlotJadwal
	if err := l.db.Select("jadwal_kelas_id").First(&slot, "id = ?", slotID).Error; err != nil {
		return ErrAksesDitolak
	}
	return l.jadwalKelas(slot.JadwalKelasID.String(), jurusanID)
}

// dokumenImpor membatasi job impor AI koor_jurusan ke semester milik jurusannya.
// Saat create, jadwal_semester_id dibaca dari multipart; saat operasi lain,
// kepemilikan diperiksa dari baris job.
func (l *Lingkup) dokumenImpor(c *fiber.Ctx, jurusanID uuid.UUID) error {
	kode, ok := l.kodeJurusan(jurusanID)
	if !ok {
		return ErrAksesDitolak
	}
	id := c.Params("id")
	if id == "" {
		form, err := c.MultipartForm()
		if err != nil {
			return ErrAksesDitolak
		}
		nilai := ""
		if daftar := form.Value["jadwal_semester_id"]; len(daftar) > 0 {
			nilai = strings.TrimSpace(daftar[0])
		}
		if nilai == "" {
			return ErrAksesDitolak
		}
		return l.semesterMilikJurusan(nilai, kode)
	}

	var job models.DokumenImpor
	if err := l.db.Select("jadwal_semester_id").First(&job, "id = ?", id).Error; err != nil {
		return ErrAksesDitolak
	}
	if job.JadwalSemesterID == nil {
		return ErrAksesDitolak
	}
	return l.semesterMilikJurusan(job.JadwalSemesterID.String(), kode)
}

func (l *Lingkup) semesterMilikJurusan(jadwalSemesterID, kode string) error {
	var jumlah int64
	err := l.db.Table("jadwal_semester_jurusan AS jsj").
		Joins("JOIN jurusan j ON j.id = jsj.jurusan_id").
		Where("jsj.jadwal_semester_id = ? AND j.kode = ?", jadwalSemesterID, kode).
		Count(&jumlah).Error
	if err != nil || jumlah == 0 {
		return ErrAksesDitolak
	}
	return nil
}

func jurusanDiBody(body []byte, jurusanID uuid.UUID) error {
	var payload struct {
		JurusanIDs []string `json:"jurusan_ids"`
	}
	if json.Unmarshal(body, &payload) != nil || len(payload.JurusanIDs) == 0 {
		return ErrAksesDitolak
	}
	for _, id := range payload.JurusanIDs {
		if !samaUUID(id, jurusanID) {
			return ErrAksesDitolak
		}
	}
	return nil
}

func samaUUID(raw string, id uuid.UUID) bool {
	parsed, err := uuid.Parse(strings.TrimSpace(raw))
	return err == nil && parsed == id
}

func menugaskanGuru(body []byte) bool {
	if len(body) == 0 {
		return false
	}
	var v any
	if json.Unmarshal(body, &v) != nil {
		return false
	}
	return cariGuru(v)
}

func cariGuru(v any) bool {
	switch t := v.(type) {
	case map[string]any:
		if raw, ok := t["guru_id"]; ok && isiGuru(raw) {
			return true
		}
		for _, child := range t {
			if cariGuru(child) {
				return true
			}
		}
	case []any:
		for _, child := range t {
			if cariGuru(child) {
				return true
			}
		}
	}
	return false
}

func isiGuru(v any) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	s = strings.TrimSpace(s)
	return s != "" && s != uuid.Nil.String()
}
