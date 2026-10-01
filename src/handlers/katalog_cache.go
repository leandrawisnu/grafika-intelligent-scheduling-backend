package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/src/models"
)

const tahanKatalog = 30 * time.Second

// Baris katalog hanya kolom yang UI pakai untuk label. Tanpa timestamp dan tanpa preload relasi.
type barisHari struct {
	ID         uuid.UUID `gorm:"column:id" json:"id"`
	Nama       string    `gorm:"column:nama" json:"nama"`
	UrutanHari int16     `gorm:"column:urutan_hari" json:"urutan_hari"`
	AkhirPekan bool      `gorm:"column:akhir_pekan" json:"akhir_pekan"`
}

type barisJam struct {
	ID           uuid.UUID `gorm:"column:id" json:"id"`
	JamKe        int16     `gorm:"column:jam_ke" json:"jam_ke"`
	WaktuMulai   string    `gorm:"column:waktu_mulai" json:"waktu_mulai"`
	WaktuSelesai string    `gorm:"column:waktu_selesai" json:"waktu_selesai"`
	Istirahat    bool      `gorm:"column:istirahat" json:"istirahat"`
}

type barisKelas struct {
	ID         uuid.UUID `gorm:"column:id" json:"id"`
	Kode       string    `gorm:"column:kode" json:"kode"`
	Nama       string    `gorm:"column:nama" json:"nama"`
	Tingkat    int16     `gorm:"column:tingkat" json:"tingkat"`
	JurusanID  uuid.UUID `gorm:"column:jurusan_id" json:"jurusan_id"`
	SemesterID uuid.UUID `gorm:"column:semester_id" json:"semester_id"`
}

type barisGuru struct {
	ID          uuid.UUID `gorm:"column:id" json:"id"`
	NIP         string    `gorm:"column:nip" json:"nip"`
	NamaLengkap string    `gorm:"column:nama_lengkap" json:"nama_lengkap"`
	Aktif       bool      `gorm:"column:aktif" json:"aktif"`
}

type barisMapel struct {
	ID                uuid.UUID `gorm:"column:id" json:"id"`
	Kode              string    `gorm:"column:kode" json:"kode"`
	Nama              string    `gorm:"column:nama" json:"nama"`
	JamWajibPerMinggu float64   `gorm:"column:jam_wajib_per_minggu" json:"jam_wajib_per_minggu"`
	Tingkat           int16     `gorm:"column:tingkat" json:"tingkat"`
}

type barisRuangan struct {
	ID          uuid.UUID `gorm:"column:id" json:"id"`
	Kode        string    `gorm:"column:kode" json:"kode"`
	Nama        string    `gorm:"column:nama" json:"nama"`
	Kapasitas   int       `gorm:"column:kapasitas" json:"kapasitas"`
	TipeRuangan string    `gorm:"column:tipe_ruangan" json:"tipe_ruangan"`
	Aktif       bool      `gorm:"column:aktif" json:"aktif"`
}

type barisJurusan struct {
	ID   uuid.UUID `gorm:"column:id" json:"id"`
	Kode string    `gorm:"column:kode" json:"kode"`
	Nama string    `gorm:"column:nama" json:"nama"`
}

type badanKatalog struct {
	Hari          []barisHari    `json:"hari"`
	JamPelajaran  []barisJam     `json:"jam_pelajaran"`
	Kelas         []barisKelas   `json:"kelas"`
	Guru          []barisGuru    `json:"guru"`
	MataPelajaran []barisMapel   `json:"mata_pelajaran"`
	Ruangan       []barisRuangan `json:"ruangan"`
	Jurusan       []barisJurusan `json:"jurusan"`
}

type entriKatalog struct {
	json        []byte
	etag        string
	kedaluwarsa time.Time
}

type penerbanganKatalog struct {
	wg   sync.WaitGroup
	json []byte
	etag string
	err  error
}

// cacheKatalog menyatukan permintaan bersamaan per lingkup peran.
// CRUD master memanggil kosongkan setelah tulis selesai di-commit.
type cacheKatalog struct {
	mu       sync.Mutex
	tahan    time.Duration
	generasi uint64
	entri    map[string]entriKatalog
	terbang  map[string]*penerbanganKatalog
}

func baruCacheKatalog(tahan time.Duration) *cacheKatalog {
	return &cacheKatalog{
		tahan:   tahan,
		entri:   map[string]entriKatalog{},
		terbang: map[string]*penerbanganKatalog{},
	}
}

func (c *cacheKatalog) kosongkan() {
	c.mu.Lock()
	c.generasi++
	c.entri = map[string]entriKatalog{}
	c.mu.Unlock()
}

func (c *cacheKatalog) ambil(kunci string, bangun func() ([]byte, string, error)) ([]byte, string, error) {
	c.mu.Lock()
	if e, ok := c.entri[kunci]; ok && time.Now().Before(e.kedaluwarsa) {
		c.mu.Unlock()
		return e.json, e.etag, nil
	}
	if f, ok := c.terbang[kunci]; ok {
		c.mu.Unlock()
		f.wg.Wait()
		return f.json, f.etag, f.err
	}
	gen := c.generasi
	f := &penerbanganKatalog{}
	f.wg.Add(1)
	c.terbang[kunci] = f
	c.mu.Unlock()

	body, etag, err := bangun()
	c.mu.Lock()
	if err == nil && c.generasi == gen {
		c.entri[kunci] = entriKatalog{
			json:        body,
			etag:        etag,
			kedaluwarsa: time.Now().Add(c.tahan),
		}
	}
	delete(c.terbang, kunci)
	c.mu.Unlock()

	f.json, f.etag, f.err = body, etag, err
	f.wg.Done()
	return body, etag, err
}

func etagDari(body []byte) string {
	sum := sha256.Sum256(body)
	return `"` + hex.EncodeToString(sum[:8]) + `"`
}

func (h *PengelolaMaster) bangunKatalog(terbatas uuid.UUID, terbatasOK bool) ([]byte, string, error) {
	var (
		hari    []barisHari
		jam     []barisJam
		kelas   []barisKelas
		guru    []barisGuru
		mapel   []barisMapel
		ruangan []barisRuangan
		jurusan []barisJurusan
		mu      sync.Mutex
		gagal   error
		wg      sync.WaitGroup
	)
	jalan := func(fn func() error) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(); err != nil {
				mu.Lock()
				if gagal == nil {
					gagal = err
				}
				mu.Unlock()
			}
		}()
	}

	jalan(func() error {
		hari = make([]barisHari, 0)
		return h.db.Model(&models.Hari{}).
			Select("id", "nama", "urutan_hari", "akhir_pekan").
			Order("urutan_hari").Find(&hari).Error
	})
	jalan(func() error {
		jam = make([]barisJam, 0)
		return h.db.Model(&models.JamPelajaran{}).
			Select("id", "jam_ke", "waktu_mulai", "waktu_selesai", "istirahat").
			Order("jam_ke").Find(&jam).Error
	})
	jalan(func() error {
		kelas = make([]barisKelas, 0)
		q := h.db.Model(&models.Kelas{}).
			Select("id", "kode", "nama", "tingkat", "jurusan_id", "semester_id").
			Order("kode")
		if terbatasOK {
			q = q.Where("jurusan_id = ?", terbatas)
		}
		return q.Find(&kelas).Error
	})
	jalan(func() error {
		guru = make([]barisGuru, 0)
		return h.db.Model(&models.Guru{}).
			Select("id", "nip", "nama_lengkap", "aktif").
			Order("nama_lengkap").Find(&guru).Error
	})
	jalan(func() error {
		mapel = make([]barisMapel, 0)
		return h.db.Model(&models.MataPelajaran{}).
			Select("id", "kode", "nama", "jam_wajib_per_minggu", "tingkat").
			Order("kode").Find(&mapel).Error
	})
	jalan(func() error {
		ruangan = make([]barisRuangan, 0)
		return h.db.Model(&models.Ruangan{}).
			Select("id", "kode", "nama", "kapasitas", "tipe_ruangan", "aktif").
			Order("kode").Find(&ruangan).Error
	})
	jalan(func() error {
		jurusan = make([]barisJurusan, 0)
		q := h.db.Model(&models.Jurusan{}).
			Select("id", "kode", "nama").
			Order("kode")
		if terbatasOK {
			q = q.Where("id = ?", terbatas)
		}
		return q.Find(&jurusan).Error
	})
	wg.Wait()
	if gagal != nil {
		return nil, "", gagal
	}
	if hari == nil {
		hari = []barisHari{}
	}
	if jam == nil {
		jam = []barisJam{}
	}
	if kelas == nil {
		kelas = []barisKelas{}
	}
	if guru == nil {
		guru = []barisGuru{}
	}
	if mapel == nil {
		mapel = []barisMapel{}
	}
	if ruangan == nil {
		ruangan = []barisRuangan{}
	}
	if jurusan == nil {
		jurusan = []barisJurusan{}
	}
	body, err := json.Marshal(badanKatalog{
		Hari:          hari,
		JamPelajaran:  jam,
		Kelas:         kelas,
		Guru:          guru,
		MataPelajaran: mapel,
		Ruangan:       ruangan,
		Jurusan:       jurusan,
	})
	if err != nil {
		return nil, "", err
	}
	return body, etagDari(body), nil
}

// mengubahKatalog true untuk tulis data yang ikut di GET /katalog.
// Hari libur guru dan jadwal kelas tidak mengubah payload katalog.
func mengubahKatalog(path string) bool {
	bagian := splitPath(path)
	for i, p := range bagian {
		if p == "api" || p == "v1" {
			continue
		}
		switch p {
		case "jurusan", "mata-pelajaran", "kelas", "ruangan", "jam-pelajaran":
			return true
		case "guru":
			return !adaSegmen(bagian[i:], "hari-libur")
		default:
			return false
		}
	}
	return false
}

func splitPath(path string) []string {
	out := make([]string, 0, 6)
	mulai := 0
	for i := 0; i <= len(path); i++ {
		if i == len(path) || path[i] == '/' {
			if i > mulai {
				out = append(out, path[mulai:i])
			}
			mulai = i + 1
		}
	}
	return out
}

func adaSegmen(bagian []string, nama string) bool {
	for _, p := range bagian {
		if p == nama {
			return true
		}
	}
	return false
}
