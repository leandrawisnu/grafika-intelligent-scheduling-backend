package handlers

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestKatalogCacheSatuPembangun(t *testing.T) {
	c := baruCacheKatalog(time.Minute)
	var n atomic.Int32
	bangun := func() ([]byte, string, error) {
		n.Add(1)
		time.Sleep(20 * time.Millisecond)
		return []byte(`{"hari":[]}`), `"e"`, nil
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body, etag, err := c.ambil("admin", bangun)
			if err != nil || string(body) != `{"hari":[]}` || etag != `"e"` {
				t.Errorf("hasil cache tidak sesuai")
			}
		}()
	}
	wg.Wait()
	if n.Load() != 1 {
		t.Fatalf("pembangun terpanggil %d kali, ingin 1", n.Load())
	}
}

func TestKatalogCacheLingkupTerpisah(t *testing.T) {
	c := baruCacheKatalog(time.Minute)
	var n atomic.Int32
	bangun := func() ([]byte, string, error) {
		n.Add(1)
		return []byte("x"), `"e"`, nil
	}
	_, _, _ = c.ambil("admin", bangun)
	_, _, _ = c.ambil("koor:abc", bangun)
	if n.Load() != 2 {
		t.Fatalf("lingkup harus terpisah, pembangun=%d", n.Load())
	}
}

func TestKatalogCacheKosongkan(t *testing.T) {
	c := baruCacheKatalog(time.Minute)
	var n atomic.Int32
	bangun := func() ([]byte, string, error) {
		n.Add(1)
		return []byte("x"), `"e"`, nil
	}
	_, _, _ = c.ambil("admin", bangun)
	c.kosongkan()
	_, _, _ = c.ambil("admin", bangun)
	if n.Load() != 2 {
		t.Fatalf("setelah kosongkan ingin bangun ulang, pembangun=%d", n.Load())
	}
}

func TestMengubahKatalog(t *testing.T) {
	kasus := []struct {
		path string
		ya   bool
	}{
		{"/api/v1/guru", true},
		{"/api/v1/guru/11111111-1111-1111-1111-111111111111", true},
		{"/api/v1/guru/11111111-1111-1111-1111-111111111111/hari-libur", false},
		{"/api/v1/kelas", true},
		{"/api/v1/jam-pelajaran/abc", true},
		{"/api/v1/jadwal-semester/abc/jadwal-kelas", false},
		{"/api/v1/katalog", false},
		{"/api/v1/tahun-ajaran", false},
	}
	for _, k := range kasus {
		if mengubahKatalog(k.path) != k.ya {
			t.Errorf("%s: dapat %v ingin %v", k.path, mengubahKatalog(k.path), k.ya)
		}
	}
}
