package auth

import (
	"testing"
	"time"

	"github.com/grafika-scheduling/backend/src/models"
)

func TestCacheSesiHitDanHapus(t *testing.T) {
	c := baruCacheSesi()
	akun := models.Pengguna{Email: "admin@example.com", Aktif: true}
	c.simpan("hash", akun, time.Now().Add(time.Minute))
	got, ok := c.ambil("hash")
	if !ok || got.Email != akun.Email {
		t.Fatal("cache sesi harus mengembalikan akun")
	}
	c.hapus("hash")
	if _, ok := c.ambil("hash"); ok {
		t.Fatal("logout harus menghapus cache sesi")
	}
}

func TestCacheSesiKedaluwarsa(t *testing.T) {
	c := baruCacheSesi()
	c.simpan("hash", models.Pengguna{Email: "a@b.c", Aktif: true}, time.Now().Add(-time.Second))
	if _, ok := c.ambil("hash"); ok {
		t.Fatal("sesi kedaluwarsa tidak boleh dipakai")
	}
}
