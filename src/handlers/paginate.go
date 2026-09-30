package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const ukuranHalaman = 10

func ukuranDariQuery(c *fiber.Ctx) int {
	switch c.QueryInt("per_page", ukuranHalaman) {
	case 50:
		return 50
	case 100:
		return 100
	default:
		return ukuranHalaman
	}
}

type opsiDaftar struct {
	defaultOrder string
	searchCols   []string
	filterCols   []string
	sortCols     map[string]string
	preloads     []string
}

func (h *PengelolaMaster) halaman(c *fiber.Ctx, base *gorm.DB, dest any, opt opsiDaftar) error {
	page := c.QueryInt("page", 0)
	if page < 1 {
		q := base
		if opt.defaultOrder != "" {
			q = q.Order(opt.defaultOrder)
		}
		for _, nama := range opt.preloads {
			q = q.Preload(nama)
		}
		err := q.Find(dest).Error
		return listJSON(c, err, dest)
	}

	filtered := saringTeks(base, strings.TrimSpace(c.Query("q")), opt.searchCols)
	filtered = saringSama(filtered, c, opt.filterCols)

	var total int64
	if err := filtered.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	order := opt.defaultOrder
	if diurut, ok := urutanAman(c.Query("sort"), opt.sortCols); ok {
		order = diurut
	}

	q := filtered.Session(&gorm.Session{})
	if order != "" {
		q = q.Order(order)
	}
	for _, nama := range opt.preloads {
		q = q.Preload(nama)
	}
	ukuran := ukuranDariQuery(c)
	if err := q.Limit(ukuran).Offset((page - 1) * ukuran).Find(dest).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"data":     dest,
		"total":    total,
		"page":     page,
		"per_page": ukuran,
	})
}

func saringTeks(db *gorm.DB, kata string, kolom []string) *gorm.DB {
	if kata == "" || len(kolom) == 0 {
		return db
	}
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(kata)
	like := "%" + escaped + "%"
	bagian := make([]string, len(kolom))
	argumen := make([]any, len(kolom))
	for i, nama := range kolom {
		bagian[i] = nama + " ILIKE ? ESCAPE '\\'"
		argumen[i] = like
	}
	return db.Where(strings.Join(bagian, " OR "), argumen...)
}

func saringSama(db *gorm.DB, c *fiber.Ctx, kolom []string) *gorm.DB {
	for _, nama := range kolom {
		nilai := strings.TrimSpace(c.Query(nama))
		if nilai == "" {
			continue
		}
		db = db.Where(nama+" = ?", nilai)
	}
	return db
}

func urutanAman(raw string, diizinkan map[string]string) (string, bool) {
	kolom, arah, ok := strings.Cut(raw, ":")
	if !ok {
		return "", false
	}
	sqlKolom, ada := diizinkan[kolom]
	if !ada || sqlKolom == "" {
		return "", false
	}
	if arah != "asc" && arah != "desc" {
		return "", false
	}
	return sqlKolom + " " + strings.ToUpper(arah), true
}
