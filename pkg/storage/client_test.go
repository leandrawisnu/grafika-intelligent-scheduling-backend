package storage

import "testing"

func TestKey(t *testing.T) {
	kunci, err := Key(PrefixJadwal, "kelas-1.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if kunci != "jadwal/kelas-1.pdf" {
		t.Fatalf("kunci = %q", kunci)
	}

	kunci, err = Key("/"+PrefixEkspor+"/", "/rekap.xlsx/")
	if err != nil {
		t.Fatal(err)
	}
	if kunci != "ekspor/rekap.xlsx" {
		t.Fatalf("kunci = %q", kunci)
	}

	if _, err := Key(PrefixDokumen, ""); err == nil {
		t.Fatal("nama kosong harus gagal")
	}
	if _, err := Key("", "a.pdf"); err == nil {
		t.Fatal("prefix kosong harus gagal")
	}
	if _, err := Key(PrefixDokumen, "../rahasia"); err == nil {
		t.Fatal("nama dengan .. harus gagal")
	}
	if _, err := Key(PrefixDokumen, "folder/a.pdf"); err == nil {
		t.Fatal("nama bersarang harus gagal")
	}
}

func TestParseKey(t *testing.T) {
	kunci, err := ParseKey("/dokumen/surat.pdf")
	if err != nil {
		t.Fatal(err)
	}
	if kunci != "dokumen/surat.pdf" {
		t.Fatalf("kunci = %q", kunci)
	}
	if _, err := ParseKey("tanpa-prefix"); err == nil {
		t.Fatal("kunci tanpa prefix harus gagal")
	}
}
