package services

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	StatusSiap         = "siap"
	StatusPerluPilihan = "perlu_pilihan"
	StatusSudahAda     = "sudah_ada"
)

type BarisMentah struct {
	Hari          string `json:"hari"`
	Jam           string `json:"jam"`
	MataPelajaran string `json:"mata_pelajaran"`
	Kelas         string `json:"kelas"`
	Guru          string `json:"guru"`
	Ruangan       string `json:"ruangan"`
}

type EntriCocok struct {
	ID   string
	Nama string
	Kode string
}

type JamCocok struct {
	ID        string
	JamKe     int
	Mulai     string
	Istirahat bool
}

type KatalogCocok struct {
	Hari    []EntriCocok
	Jam     []JamCocok
	Mapel   []EntriCocok
	Kelas   []EntriCocok
	Guru    []EntriCocok
	Ruangan []EntriCocok
}

type HasilCocok struct {
	Hari            string `json:"hari"`
	Jam             string `json:"jam"`
	MataPelajaran   string `json:"mata_pelajaran"`
	Kelas           string `json:"kelas"`
	Guru            string `json:"guru"`
	Ruangan         string `json:"ruangan"`
	HariID          string `json:"hari_id,omitempty"`
	JamPelajaranID  string `json:"jam_pelajaran_id,omitempty"`
	MataPelajaranID string `json:"mata_pelajaran_id,omitempty"`
	KelasID         string `json:"kelas_id,omitempty"`
	GuruID          string `json:"guru_id,omitempty"`
	RuanganID       string `json:"ruangan_id,omitempty"`
	Status          string `json:"status"`
}

var (
	reWaktu = regexp.MustCompile(`(\d{1,2})[:.](\d{2})`)
	reJamKe = regexp.MustCompile(`(?i)^(?:jam\s*)?(?:ke\s*-?\s*)?(\d{1,2})$`)
)

var kanonHari = map[string]string{
	"senin": "senin", "monday": "senin",
	"selasa": "selasa", "tuesday": "selasa",
	"rabu": "rabu", "wednesday": "rabu",
	"kamis": "kamis", "thursday": "kamis",
	"jumat": "jumat", "friday": "jumat",
	"sabtu": "sabtu", "saturday": "sabtu",
	"minggu": "minggu", "ahad": "minggu", "sunday": "minggu",
}

func norm(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	spasi := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			spasi = false
			continue
		}
		if !spasi {
			b.WriteByte(' ')
			spasi = true
		}
	}
	return strings.TrimSpace(b.String())
}

func normHari(s string) string {
	rapat := strings.ReplaceAll(norm(s), " ", "")
	if kanon, ok := kanonHari[rapat]; ok {
		return kanon
	}
	return rapat
}

func jamMenit(s string) (string, bool) {
	m := reWaktu.FindStringSubmatch(s)
	if m == nil {
		return "", false
	}
	jam, _ := strconv.Atoi(m[1])
	menit, _ := strconv.Atoi(m[2])
	if jam > 23 || menit > 59 {
		return "", false
	}
	return fmt.Sprintf("%02d:%02d", jam, menit), true
}

func potongWaktu(s string) string {
	if mulai, ok := jamMenit(s); ok {
		return mulai
	}
	return ""
}

func cocokNama(q string, entri []EntriCocok) string {
	q = norm(q)
	if q == "" {
		return ""
	}
	var tepat []string
	var mirip []string
	for _, e := range entri {
		nama := norm(e.Nama)
		kode := norm(e.Kode)
		if nama == q || (kode != "" && kode == q) {
			tepat = append(tepat, e.ID)
			continue
		}
		if len(q) < 3 {
			continue
		}
		if (nama != "" && (strings.Contains(nama, q) || strings.Contains(q, nama))) ||
			(kode != "" && (strings.Contains(kode, q) || strings.Contains(q, kode))) {
			mirip = append(mirip, e.ID)
		}
	}
	if len(tepat) == 1 {
		return tepat[0]
	}
	if len(tepat) > 1 {
		return ""
	}
	if len(mirip) == 1 {
		return mirip[0]
	}
	return ""
}

func cocokHari(q string, hari []EntriCocok) string {
	ingin := normHari(q)
	if ingin == "" {
		return ""
	}
	var ketemu []string
	for _, h := range hari {
		if normHari(h.Nama) == ingin {
			ketemu = append(ketemu, h.ID)
		}
	}
	if len(ketemu) == 1 {
		return ketemu[0]
	}
	return ""
}

func cocokJam(q string, jam []JamCocok) (string, bool) {
	q = strings.TrimSpace(q)
	if q == "" || norm(q) == "istirahat" {
		return "", norm(q) == "istirahat"
	}
	if _, adaWaktu := jamMenit(q); adaWaktu {
		mulai, _ := jamMenit(q)
		var ketemu []JamCocok
		for _, j := range jam {
			if potongWaktu(j.Mulai) == mulai {
				ketemu = append(ketemu, j)
			}
		}
		if len(ketemu) == 1 {
			return ketemu[0].ID, ketemu[0].Istirahat
		}
		return "", false
	}
	m := reJamKe.FindStringSubmatch(norm(q))
	if m == nil {
		return "", false
	}
	ke, _ := strconv.Atoi(m[1])
	var ketemu []JamCocok
	for _, j := range jam {
		if j.JamKe == ke {
			ketemu = append(ketemu, j)
		}
	}
	if len(ketemu) == 1 {
		return ketemu[0].ID, ketemu[0].Istirahat
	}
	return "", false
}

func barisKosong(b BarisMentah) bool {
	return norm(b.Hari) == "" && norm(b.Jam) == "" && norm(b.MataPelajaran) == "" &&
		norm(b.Kelas) == "" && norm(b.Guru) == "" && norm(b.Ruangan) == ""
}

func buangBaris(b BarisMentah, jamIstirahat bool) bool {
	if barisKosong(b) || jamIstirahat {
		return true
	}
	mapel := norm(b.MataPelajaran)
	return mapel == "istirahat" || mapel == "rehat"
}

func kunciSlot(kelasID, hariID, jamID string) string {
	return kelasID + "|" + hariID + "|" + jamID
}

// Cocokkan memasangkan teks ekstraksi ke id master.
// sudah berisi kunci kelas|hari|jam yang sudah tersimpan. Baris kedua dengan kunci yang sama ditandai sudah_ada.
func Cocokkan(baris []BarisMentah, katalog KatalogCocok, sudah map[string]struct{}) []HasilCocok {
	if sudah == nil {
		sudah = map[string]struct{}{}
	}
	terpakai := map[string]struct{}{}
	for k, v := range sudah {
		terpakai[k] = v
	}
	hasil := make([]HasilCocok, 0, len(baris))
	for _, b := range baris {
		jamID, istirahat := cocokJam(b.Jam, katalog.Jam)
		if buangBaris(b, istirahat) {
			continue
		}
		h := HasilCocok{
			Hari:            b.Hari,
			Jam:             b.Jam,
			MataPelajaran:   b.MataPelajaran,
			Kelas:           b.Kelas,
			Guru:            b.Guru,
			Ruangan:         b.Ruangan,
			HariID:          cocokHari(b.Hari, katalog.Hari),
			JamPelajaranID:  jamID,
			MataPelajaranID: cocokNama(b.MataPelajaran, katalog.Mapel),
			KelasID:         cocokNama(b.Kelas, katalog.Kelas),
			GuruID:          cocokNama(b.Guru, katalog.Guru),
			RuanganID:       cocokNama(b.Ruangan, katalog.Ruangan),
		}
		if h.KelasID == "" || h.MataPelajaranID == "" || h.HariID == "" || h.JamPelajaranID == "" {
			h.Status = StatusPerluPilihan
			hasil = append(hasil, h)
			continue
		}
		kunci := kunciSlot(h.KelasID, h.HariID, h.JamPelajaranID)
		if _, ada := terpakai[kunci]; ada {
			h.Status = StatusSudahAda
			hasil = append(hasil, h)
			continue
		}
		h.Status = StatusSiap
		terpakai[kunci] = struct{}{}
		hasil = append(hasil, h)
	}
	return hasil
}
