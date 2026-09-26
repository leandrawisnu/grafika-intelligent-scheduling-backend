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

func Dari(c *fiber.Ctx) (models.Pengguna, bool) {
	akun, ok := c.Locals(LocalPengguna).(models.Pengguna)
	return akun, ok
}

func SaringSemester(c *fiber.Ctx, js *models.JadwalSemester) {
	id, ok := Terbatas(c)
	if !ok || js == nil {
		return
	}
	jurusan := make([]models.JadwalSemesterJurusan, 0)
	for _, item := range js.Jurusan {
		if item.JurusanID == id {
			jurusan = append(jurusan, item)
		}
	}
	js.Jurusan = jurusan
	kelas := make([]models.JadwalKelas, 0)
	for _, item := range js.JadwalKelas {
		if item.JurusanID == id {
			kelas = append(kelas, item)
		}
	}
	js.JadwalKelas = kelas
}

func (l *Lingkup) kelas(c *fiber.Ctx, jurusanID uuid.UUID) error {
	if c.Method() == fiber.MethodPost {
		var body struct {
			JurusanID string `json:"jurusan_id"`
		}
		if json.Unmarshal(c.Body(), &body) != nil || !samaUUID(body.JurusanID, jurusanID) {
			return ErrAksesDitolak
		}
		return nil
	}
	var kelas models.Kelas
	if err := l.db.Select("jurusan_id").First(&kelas, "id = ?", c.Params("id")).Error; err != nil {
		return nil
	}
	if kelas.JurusanID != jurusanID {
		return ErrAksesDitolak
	}
	if c.Method() == fiber.MethodPut {
		var body struct {
			JurusanID string `json:"jurusan_id"`
		}
		if json.Unmarshal(c.Body(), &body) == nil && body.JurusanID != "" && !samaUUID(body.JurusanID, jurusanID) {
			return ErrAksesDitolak
		}
	}
	return nil
}

func (l *Lingkup) kelasDiBody(c *fiber.Ctx, jurusanID uuid.UUID) error {
	var body struct {
		KelasID string `json:"kelas_id"`
	}
	if json.Unmarshal(c.Body(), &body) != nil {
		return ErrAksesDitolak
	}
	var kelas models.Kelas
	if err := l.db.Select("jurusan_id").First(&kelas, "id = ?", body.KelasID).Error; err != nil {
		return ErrAksesDitolak
	}
	if kelas.JurusanID != jurusanID {
		return ErrAksesDitolak
	}
	return nil
}

func (l *Lingkup) jadwalKelas(id string, jurusanID uuid.UUID) error {
	var jk models.JadwalKelas
	if err := l.db.Select("jurusan_id").First(&jk, "id = ?", id).Error; err != nil {
		return ErrAksesDitolak
	}
	if jk.JurusanID != jurusanID {
		return ErrAksesDitolak
	}
	return nil
}

func (l *Lingkup) slot(slotID string, jurusanID uuid.UUID) error {
	var slot models.SlotJadwal
	if err := l.db.Select("jadwal_kelas_id").First(&slot, "id = ?", slotID).Error; err != nil {
		return ErrAksesDitolak
	}
	return l.jadwalKelas(slot.JadwalKelasID.String(), jurusanID)
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
