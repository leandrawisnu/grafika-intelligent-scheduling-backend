package auth

import "testing"

func TestHashSandiCocokDanSalah(t *testing.T) {
	hash, err := HashSandi("sandi-yang-kuat")
	if err != nil {
		t.Fatal(err)
	}
	ok, err := CocokkanSandi(hash, "sandi-yang-kuat")
	if err != nil || !ok {
		t.Fatalf("sandi benar ditolak: ok=%v err=%v", ok, err)
	}
	ok, err = CocokkanSandi(hash, "sandi-lain")
	if err != nil || ok {
		t.Fatalf("sandi salah diterima: ok=%v err=%v", ok, err)
	}
}

func TestCocokkanSandiHashRusak(t *testing.T) {
	if _, err := CocokkanSandi("bukan-hash", "apa"); err == nil {
		t.Fatal("hash rusak harus error")
	}
}
