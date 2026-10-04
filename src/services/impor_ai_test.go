package services

import (
	"testing"

	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/src/dto"
)

func katalogValidasiUji() KatalogValidasi {
	return KatalogValidasi{
		Hari:         map[string]bool{"h1": true},
		Jam:          map[string]bool{"j1": true, "j2": true},
		Kelas:        map[string]bool{"k1": true},
		Mapel:        map[string]bool{"m1": true},
		Guru:         map[string]bool{"g1": true},
		Ruangan:      map[string]bool{"r1": true},
		Jurusan:      map[string]bool{"ju1": true},
		SlotTerpakai: map[string]bool{"k1|h1|j1": true},
	}
}

func validasiRencanaUji(t *testing.T, plan dto.ImportPlan) dto.ImportPlan {
	t.Helper()
	return ValidasiRencana(plan, katalogValidasiUji())
}

func TestValidasiRencanaMenghitungStatus(t *testing.T) {
	plan := dto.ImportPlan{
		MasterUsulan: dto.MasterUsulanPlan{
			Guru:          []dto.MasterGuruPlan{{Ref: "g9", Nama: "Andi Pratama"}},
			MataPelajaran: []dto.MasterMapelPlan{{Ref: "m9", Nama: "Pemrograman Dasar"}},
			Kelas:         []dto.MasterKelasPlan{{Ref: "k9", Nama: "XI RPL 2", JurusanID: "ju1"}},
		},
		BarisJadwal: []dto.BarisJadwalPlan{
			{HariID: "h1", JamPelajaranID: "j2", KelasID: "k1", MataPelajaranID: "m1", Status: "perlu_pilihan"},
			{HariID: "h1", JamPelajaranID: "j1", KelasID: "k1", MataPelajaranID: "m1"},
			{HariID: "h1", JamPelajaranID: "j2", KelasRef: "k9", MapelRef: "m9", GuruRef: "g9"},
			{HariID: "h1", JamPelajaranID: "j2", KelasID: "tidak-ada", MapelRef: "ref-ngawur"},
			{HariID: "h1", KelasID: "k1", MataPelajaranID: "m1"},
		},
	}
	hasil := validasiRencanaUji(t, plan)

	harapan := []string{"siap", "sudah_ada", "akan_dibuat", "perlu_pilihan", "perlu_pilihan"}
	for i, mau := range harapan {
		if hasil.BarisJadwal[i].Status != mau {
			t.Fatalf("baris %d status = %s, ingin %s", i, hasil.BarisJadwal[i].Status, mau)
		}
	}

	if len(hasil.MasterUsulan.Guru) != 1 || len(hasil.MasterUsulan.Kelas) != 1 || len(hasil.MasterUsulan.MataPelajaran) != 1 {
		t.Fatalf("master usulan tervalidasi tidak lengkap: %+v", hasil.MasterUsulan)
	}
}

func TestValidasiRencanaMenolakIdDanRefPalsu(t *testing.T) {
	plan := dto.ImportPlan{
		MasterUsulan: dto.MasterUsulanPlan{
			Guru: []dto.MasterGuruPlan{{Ref: "g9", Nama: ""}}, // nama kosong -> dibuang
		},
		BarisJadwal: []dto.BarisJadwalPlan{
			{
				HariID: "hari-palsu", JamPelajaranID: "jam-palsu",
				KelasID: "kelas-palsu", MataPelajaranID: "mapel-palsu",
				GuruID: "guru-palsu", RuanganID: "ruangan-palsu",
				KelasRef: "ref-palsu", MapelRef: "ref-palsu", GuruRef: "ref-palsu", RuanganRef: "ref-palsu",
			},
		},
	}
	hasil := validasiRencanaUji(t, plan)

	if len(hasil.MasterUsulan.Guru) != 0 {
		t.Fatalf("guru tanpa nama harus dibuang, dapat %d", len(hasil.MasterUsulan.Guru))
	}
	b := hasil.BarisJadwal[0]
	if b.HariID != "" || b.JamPelajaranID != "" || b.KelasID != "" || b.MataPelajaranID != "" {
		t.Fatalf("id palsu harus dikosongkan: %+v", b)
	}
	if b.KelasRef != "" || b.MapelRef != "" || b.GuruRef != "" || b.RuanganRef != "" {
		t.Fatalf("ref palsu harus dikosongkan: %+v", b)
	}
	if b.Status != "perlu_pilihan" {
		t.Fatalf("status = %s, ingin perlu_pilihan", b.Status)
	}
}

func TestValidasiRencanaIdMenangAtasRef(t *testing.T) {
	plan := dto.ImportPlan{
		MasterUsulan: dto.MasterUsulanPlan{
			MataPelajaran: []dto.MasterMapelPlan{{Ref: "m9", Nama: "Pemrograman Dasar"}},
		},
		BarisJadwal: []dto.BarisJadwalPlan{
			{HariID: "h1", JamPelajaranID: "j2", KelasID: "k1", MataPelajaranID: "m1", MapelRef: "m9"},
		},
	}
	hasil := validasiRencanaUji(t, plan)
	b := hasil.BarisJadwal[0]
	if b.MapelRef != "" {
		t.Fatalf("ref harus dibersihkan bila id sudah terisi: %+v", b)
	}
	if b.Status != "siap" {
		t.Fatalf("status = %s, ingin siap", b.Status)
	}
}

func TestSusunBarisImporMelewatiBarisTidakSiap(t *testing.T) {
	kelasID := uuid.New()
	mapelID := uuid.New()
	hariID := uuid.New()
	jamID := uuid.New()
	baris, dilewati := susunBarisImpor(
		[]dto.BarisJadwalPlan{
			{HariID: hariID.String(), JamPelajaranID: jamID.String(), KelasID: kelasID.String(), MataPelajaranID: mapelID.String(), Status: "siap"},
			{HariID: hariID.String(), JamPelajaranID: jamID.String(), KelasID: kelasID.String(), MataPelajaranID: mapelID.String(), Status: "perlu_pilihan"},
			{HariID: hariID.String(), JamPelajaranID: jamID.String(), KelasID: kelasID.String(), MataPelajaranID: mapelID.String(), Status: "sudah_ada"},
		},
		map[string]uuid.UUID{}, map[string]uuid.UUID{}, map[string]uuid.UUID{}, map[string]uuid.UUID{},
	)
	if len(baris) != 1 {
		t.Fatalf("baris siap = %d, ingin 1", len(baris))
	}
	if dilewati != 2 {
		t.Fatalf("dilewati = %d, ingin 2", dilewati)
	}
	if baris[0].KelasID != kelasID || baris[0].MataPelajaranID != mapelID {
		t.Fatalf("baris tersusun salah: %+v", baris[0])
	}
}

func TestSusunBarisImporMemakaiRefMasterBaru(t *testing.T) {
	refKelas := uuid.New()
	refMapel := uuid.New()
	hariID := uuid.New()
	jamID := uuid.New()
	baris, dilewati := susunBarisImpor(
		[]dto.BarisJadwalPlan{
			{
				HariID: hariID.String(), JamPelajaranID: jamID.String(),
				KelasRef: "k1", MapelRef: "m1", Status: "akan_dibuat",
			},
		},
		map[string]uuid.UUID{"k1": refKelas}, map[string]uuid.UUID{"m1": refMapel},
		map[string]uuid.UUID{}, map[string]uuid.UUID{},
	)
	if dilewati != 0 || len(baris) != 1 {
		t.Fatalf("hasil = %d baris, %d dilewati", len(baris), dilewati)
	}
	if baris[0].KelasID != refKelas || baris[0].MataPelajaranID != refMapel {
		t.Fatalf("ref tidak teresolusi: %+v", baris[0])
	}
}

func TestKodeDariNama(t *testing.T) {
	if got := kodeDariNama("Pemrograman Dasar", "X"); got != "PEMROGRA" {
		t.Fatalf("kode = %s", got)
	}
	if got := kodeDariNama("", "MAPEL"); got != "MAPEL" {
		t.Fatalf("kode kosong = %s", got)
	}
}
