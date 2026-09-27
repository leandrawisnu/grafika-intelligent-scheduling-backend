package services

import "testing"

func katalogUji() KatalogCocok {
	return KatalogCocok{
		Hari: []EntriCocok{{ID: "h-sen", Nama: "Senin"}, {ID: "h-sel", Nama: "Selasa"}},
		Jam: []JamCocok{
			{ID: "j1", JamKe: 1, Mulai: "07:00:00"},
			{ID: "j2", JamKe: 2, Mulai: "07:45:00"},
			{ID: "j-rehat", JamKe: 3, Mulai: "08:30:00", Istirahat: true},
		},
		Mapel: []EntriCocok{
			{ID: "m-mtk", Nama: "Matematika", Kode: "MTK"},
			{ID: "m-bind", Nama: "Bahasa Indonesia", Kode: "BIND"},
		},
		Kelas: []EntriCocok{
			{ID: "k-1", Nama: "XII RPL 1", Kode: "12RPL1"},
			{ID: "k-2", Nama: "XII RPL 2", Kode: "12RPL2"},
		},
		Guru:    []EntriCocok{{ID: "g-budi", Nama: "Budi Santoso", Kode: "198001"}},
		Ruangan: []EntriCocok{{ID: "r-lab", Nama: "Lab RPL", Kode: "LAB1"}},
	}
}

func TestCocokkanSiapDanGuruKosong(t *testing.T) {
	hasil := Cocokkan([]BarisMentah{{
		Hari: "Senin", Jam: "1", MataPelajaran: "MTK", Kelas: "XII RPL 1", Ruangan: "Lab RPL",
	}}, katalogUji(), nil)
	if len(hasil) != 1 {
		t.Fatalf("jumlah = %d", len(hasil))
	}
	dapat := hasil[0]
	if dapat.Status != StatusSiap || dapat.HariID != "h-sen" || dapat.JamPelajaranID != "j1" ||
		dapat.MataPelajaranID != "m-mtk" || dapat.KelasID != "k-1" || dapat.RuanganID != "r-lab" || dapat.GuruID != "" {
		t.Fatalf("hasil = %+v", dapat)
	}
}

func TestCocokkanJamWaktuDanHariInggris(t *testing.T) {
	hasil := Cocokkan([]BarisMentah{{
		Hari: "Monday", Jam: "07.00-07.40", MataPelajaran: "Matematika", Kelas: "12RPL1", Guru: "Budi",
	}}, katalogUji(), nil)
	if len(hasil) != 1 || hasil[0].HariID != "h-sen" || hasil[0].JamPelajaranID != "j1" || hasil[0].GuruID != "g-budi" {
		t.Fatalf("hasil = %+v", hasil)
	}
}

func TestCocokkanPerluPilihanDanSudahAda(t *testing.T) {
	sudah := map[string]struct{}{kunciSlot("k-1", "h-sen", "j1"): {}}
	hasil := Cocokkan([]BarisMentah{
		{Hari: "Senin", Jam: "1", MataPelajaran: "RPL", Kelas: "XII RPL 1"},
		{Hari: "Senin", Jam: "1", MataPelajaran: "Matematika", Kelas: "XII RPL 1"},
		{Hari: "Selasa", Jam: "2", MataPelajaran: "Bahasa Indonesia", Kelas: "XII RPL 1"},
		{Hari: "Selasa", Jam: "2", MataPelajaran: "Bahasa Indonesia", Kelas: "XII RPL 1"},
	}, katalogUji(), sudah)
	if len(hasil) != 4 {
		t.Fatalf("jumlah = %d", len(hasil))
	}
	if hasil[0].Status != StatusPerluPilihan || hasil[0].MataPelajaranID != "" {
		t.Fatalf("ambigu = %+v", hasil[0])
	}
	if hasil[1].Status != StatusSudahAda {
		t.Fatalf("bentrok db = %+v", hasil[1])
	}
	if hasil[2].Status != StatusSiap || hasil[3].Status != StatusSudahAda {
		t.Fatalf("duplikat batch = %+v %+v", hasil[2], hasil[3])
	}
}

func TestCocokkanBuangIstirahat(t *testing.T) {
	hasil := Cocokkan([]BarisMentah{
		{Hari: "Senin", Jam: "3", MataPelajaran: "Matematika", Kelas: "XII RPL 1"},
		{Hari: "Senin", Jam: "1", MataPelajaran: "Istirahat", Kelas: "XII RPL 1"},
		{Hari: "Senin", Jam: "istirahat", MataPelajaran: "Matematika", Kelas: "XII RPL 1"},
		{},
	}, katalogUji(), nil)
	if len(hasil) != 0 {
		t.Fatalf("hasil = %+v", hasil)
	}
}
