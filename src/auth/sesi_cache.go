package auth

import (
	"sync"
	"time"

	"github.com/grafika-scheduling/backend/src/models"
)

const tahanSesiCache = 20 * time.Second

type entriSesi struct {
	akun        models.Pengguna
	kedaluwarsa time.Time
}

// cacheSesi menyimpan akun untuk token yang baru dicek.
// Logout menghapus entri. Akun yang dinonaktifkan paling lama tertahan selama tahanSesiCache.
type cacheSesi struct {
	mu    sync.Mutex
	entri map[string]entriSesi
}

func baruCacheSesi() *cacheSesi {
	return &cacheSesi{entri: map[string]entriSesi{}}
}

func (c *cacheSesi) ambil(hash string) (models.Pengguna, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entri[hash]
	if !ok || !time.Now().Before(e.kedaluwarsa) {
		if ok {
			delete(c.entri, hash)
		}
		return models.Pengguna{}, false
	}
	return e.akun, true
}

func (c *cacheSesi) simpan(hash string, akun models.Pengguna, kedaluwarsa time.Time) {
	c.mu.Lock()
	c.entri[hash] = entriSesi{akun: akun, kedaluwarsa: kedaluwarsa}
	c.mu.Unlock()
}

func (c *cacheSesi) hapus(hash string) {
	c.mu.Lock()
	delete(c.entri, hash)
	c.mu.Unlock()
}
