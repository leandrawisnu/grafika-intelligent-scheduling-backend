package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/grafika-scheduling/backend/src/models"
	"gorm.io/gorm"
)

const MasaSesi = 24 * time.Hour

var (
	ErrKredensial    = errors.New("kredensial salah")
	ErrTerlaluSering = errors.New("terlalu sering")
	ErrSesi          = errors.New("sesi tidak berlaku")
)

type Layanan struct {
	db         *gorm.DB
	pembatas   *Pembatas
	hashKosong string
}

func NewLayanan(db *gorm.DB) *Layanan {
	raw := make([]byte, 32)
	_, _ = rand.Read(raw)
	hash, err := HashSandi(base64.RawURLEncoding.EncodeToString(raw))
	if err != nil {
		log.Fatalf("hash dummy: %v", err)
	}
	return &Layanan{db: db, pembatas: BaruPembatas(8, 15*time.Minute), hashKosong: hash}
}

func (l *Layanan) Masuk(ip, email, sandi string) (string, models.Pengguna, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if l.pembatas.TerlaluSering(ip) || l.pembatas.TerlaluSering("email:"+email) {
		return "", models.Pengguna{}, ErrTerlaluSering
	}

	var pengguna models.Pengguna
	err := l.db.Where("email = ?", email).First(&pengguna).Error
	ada := err == nil && pengguna.Aktif
	hash := l.hashKosong
	if ada {
		hash = pengguna.PasswordHash
	}
	ok, verr := CocokkanSandi(hash, sandi)
	if verr != nil || !ok || !ada {
		l.pembatas.Catat(ip)
		l.pembatas.Catat("email:" + email)
		return "", models.Pengguna{}, ErrKredensial
	}

	token, err := l.BuatSesi(pengguna.ID)
	if err != nil {
		return "", models.Pengguna{}, err
	}
	return token, pengguna, nil
}

func (l *Layanan) BuatSesi(penggunaID uuid.UUID) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	sesi := models.Sesi{
		PenggunaID:  penggunaID,
		TokenHash:   HashToken(token),
		Kedaluwarsa: time.Now().Add(MasaSesi),
	}
	if err := l.db.Create(&sesi).Error; err != nil {
		return "", err
	}
	return token, nil
}

func (l *Layanan) PenggunaDariToken(token string) (models.Pengguna, error) {
	if token == "" {
		return models.Pengguna{}, ErrSesi
	}
	var sesi models.Sesi
	err := l.db.Preload("Pengguna").
		Where("token_hash = ? AND kedaluwarsa > ?", HashToken(token), time.Now()).
		First(&sesi).Error
	if err != nil || sesi.Pengguna == nil || !sesi.Pengguna.Aktif {
		return models.Pengguna{}, ErrSesi
	}
	return *sesi.Pengguna, nil
}

func (l *Layanan) HapusToken(token string) {
	if token == "" {
		return
	}
	l.db.Where("token_hash = ?", HashToken(token)).Delete(&models.Sesi{})
}

func (l *Layanan) PastikanAdmin(email, sandi string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || sandi == "" {
		log.Println("akun admin awal tidak dibuat: GIS_ADMIN_EMAIL atau GIS_ADMIN_PASSWORD kosong")
		return nil
	}
	if len(sandi) < 8 {
		return errors.New("GIS_ADMIN_PASSWORD terlalu pendek (minimal 8 karakter)")
	}
	var n int64
	if err := l.db.Model(&models.Pengguna{}).Where("email = ?", email).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	hash, err := HashSandi(sandi)
	if err != nil {
		return err
	}
	akun := models.Pengguna{
		Email:        email,
		PasswordHash: hash,
		Peran:        "admin",
		Aktif:        true,
	}
	if err := l.db.Create(&akun).Error; err != nil {
		return err
	}
	log.Printf("akun admin awal dibuat untuk %s\n", email)
	return nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

type ember struct {
	jumlah int
	reset  time.Time
}

type Pembatas struct {
	mu      sync.Mutex
	batas   int
	jendela time.Duration
	hit     map[string]*ember
}

func BaruPembatas(batas int, jendela time.Duration) *Pembatas {
	return &Pembatas{batas: batas, jendela: jendela, hit: map[string]*ember{}}
}

func (p *Pembatas) TerlaluSering(kunci string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	e, ok := p.hit[kunci]
	if !ok || time.Now().After(e.reset) {
		return false
	}
	return e.jumlah >= p.batas
}

func (p *Pembatas) Catat(kunci string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	e, ok := p.hit[kunci]
	if !ok || now.After(e.reset) {
		p.hit[kunci] = &ember{jumlah: 1, reset: now.Add(p.jendela)}
		return
	}
	e.jumlah++
}
