package auth

import "testing"

func TestGerbangKoor(t *testing.T) {
	kasus := []struct {
		method string
		path   string
		ingin  string
	}{
		{"POST", "/api/v1/auth/keluar", GerbangBaca},
		{"GET", "/api/v1/auth/sesi", GerbangBaca},
		{"GET", "/api/v1/jurusan", GerbangBaca},
		{"POST", "/api/v1/jadwal-semester/abc/publikasi", GerbangTolak},
		{"POST", "/api/v1/jadwal-semester/abc/batalkan-publikasi", GerbangTolak},
		{"PUT", "/api/v1/slot/abc/tugaskan-guru", GerbangTolak},
		{"POST", "/api/v1/tahun-ajaran", GerbangTolak},
		{"POST", "/api/v1/jadwal-semester", GerbangTolak},
		{"POST", "/api/v1/kelas", GerbangCekKelas},
		{"PUT", "/api/v1/kelas/abc", GerbangCekKelas},
		{"POST", "/api/v1/jadwal-semester/abc/jadwal-kelas", GerbangCekKelasBody},
		{"POST", "/api/v1/jadwal-kelas/abc/slot", GerbangCekJadwalKelas},
		{"POST", "/api/v1/jadwal-kelas/abc/slot/massal", GerbangCekJadwalKelas},
		{"PUT", "/api/v1/slot/abc", GerbangCekSlot},
		{"DELETE", "/api/v1/slot/abc", GerbangCekSlot},
		{"POST", "/api/v1/jadwal-semester/abc/jurusan", GerbangCekJurusanBody},
		{"DELETE", "/api/v1/jadwal-semester/abc/jurusan/def", GerbangCekJurusanParam},
	}
	for _, k := range kasus {
		if dapat := GerbangKoor(k.method, k.path); dapat != k.ingin {
			t.Errorf("%s %s = %s, ingin %s", k.method, k.path, dapat, k.ingin)
		}
	}
}
