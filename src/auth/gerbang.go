package auth

import "strings"

const (
	GerbangBaca            = "baca"
	GerbangTolak           = "tolak"
	GerbangCekKelas        = "cek-kelas"
	GerbangCekKelasBody    = "cek-kelas-body"
	GerbangCekJadwalKelas  = "cek-jadwal-kelas"
	GerbangCekSlot         = "cek-slot"
	GerbangCekJurusanBody  = "cek-jurusan-body"
	GerbangCekJurusanParam = "cek-jurusan-param"
)

// GerbangKoor mengelompokkan mutasi koor jurusan.
// Baca lolos. Mutasi sekolah ditolak. Mutasi yang terikat jurusan dicek pemanggil.
func GerbangKoor(method, path string) string {
	if strings.Contains(path, "/auth/") {
		return GerbangBaca
	}
	method = strings.ToUpper(method)
	if method == "GET" || method == "HEAD" || method == "OPTIONS" {
		return GerbangBaca
	}
	p := path
	switch {
	case strings.Contains(p, "/tugaskan-guru"),
		strings.Contains(p, "/publikasi"),
		strings.Contains(p, "/batalkan-publikasi"),
		strings.Contains(p, "/status"),
		strings.Contains(p, "/validasi"),
		strings.Contains(p, "/prediksi-konflik"),
		strings.Contains(p, "/konflik"),
		strings.Contains(p, "/resolusi"),
		strings.Contains(p, "/ai/"),
		strings.Contains(p, "/pengguna"),
		strings.Contains(p, "/tahun-ajaran"),
		strings.Contains(p, "/semester"),
		strings.Contains(p, "/mata-pelajaran"),
		strings.Contains(p, "/ruangan"),
		strings.Contains(p, "/jam-pelajaran"),
		strings.Contains(p, "/guru"):
		return GerbangTolak
	case strings.Contains(p, "/jadwal-semester/") && strings.HasSuffix(strings.TrimRight(p, "/"), "/jurusan"):
		return GerbangCekJurusanBody
	case strings.Contains(p, "/jadwal-semester/") && strings.Contains(p, "/jurusan/"):
		return GerbangCekJurusanParam
	case strings.Contains(p, "/jadwal-kelas/") && strings.Contains(p, "/slot"):
		return GerbangCekJadwalKelas
	case strings.HasSuffix(p, "/jadwal-kelas"):
		return GerbangCekKelasBody
	case strings.Contains(p, "/slot/"):
		return GerbangCekSlot
	case strings.Contains(p, "/kelas"):
		return GerbangCekKelas
	default:
		return GerbangTolak
	}
}
