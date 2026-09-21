package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Versi migrasi terakhir; setelah -migrate init dipakai untuk force agar up berikutnya no-op.
const LatestMigrationVersion = 20260920123000

// RunMigrateCommand menjalankan subcommand golang-migrate (up, down, version, force).
func RunMigrateCommand(databaseURL, migrationsDir, command string, args []string) error {
	root, err := moduleRoot()
	if err != nil {
		return err
	}
	if migrationsDir == "" {
		migrationsDir = filepath.Join(root, "database", "migrations")
	} else if !filepath.IsAbs(migrationsDir) {
		migrationsDir = filepath.Join(root, migrationsDir)
	}

	switch command {
	case "up":
		return runMigrate(databaseURL, migrationsDir, func(m *migrate.Migrate) error {
			return m.Up()
		})
	case "down":
		return runMigrate(databaseURL, migrationsDir, func(m *migrate.Migrate) error {
			return m.Steps(-1)
		})
	case "version":
		return runMigrate(databaseURL, migrationsDir, func(m *migrate.Migrate) error {
			v, dirty, err := m.Version()
			if err != nil {
				if errors.Is(err, migrate.ErrNilVersion) {
					log.Println("migrate version: (belum ada migrasi)")
					return nil
				}
				return err
			}
			log.Printf("migrate version: %d (dirty=%v)\n", v, dirty)
			return nil
		})
	case "force":
		if len(args) < 1 {
			return fmt.Errorf("force membutuhkan nomor versi, contoh: -migrate force 20260920123000")
		}
		var version uint
		if _, err := fmt.Sscan(args[0], &version); err != nil {
			return fmt.Errorf("versi tidak valid: %w", err)
		}
		return runMigrate(databaseURL, migrationsDir, func(m *migrate.Migrate) error {
			return m.Force(int(version))
		})
	default:
		return fmt.Errorf("perintah migrate tidak dikenal: %q (init, up, down, version, force)", command)
	}
}

func runMigrate(databaseURL, migrationsDir string, fn func(*migrate.Migrate) error) error {
	m, err := newMigrator(databaseURL, migrationsDir)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := fn(m); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}

func newMigrator(databaseURL, migrationsDir string) (*migrate.Migrate, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("buka database: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("driver migrate: %w", err)
	}

	source := "file://" + filepath.ToSlash(migrationsDir)
	m, err := migrate.NewWithDatabaseInstance(source, "postgres", driver)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate instance: %w", err)
	}
	return m, nil
}

// RunInitSQL menjalankan database/migrations/init.sql pada database kosong, lalu force versi migrasi terakhir.
func RunInitSQL(databaseURL string) error {
	root, err := moduleRoot()
	if err != nil {
		return err
	}
	initPath := filepath.Join(root, "database", "migrations", "init.sql")
	sqlBytes, err := os.ReadFile(initPath)
	if err != nil {
		return fmt.Errorf("baca init.sql: %w", err)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("buka database: %w", err)
	}
	defer db.Close()

	if _, err := db.Exec(string(sqlBytes)); err != nil {
		return fmt.Errorf("eksekusi init.sql: %w", err)
	}
	log.Println("init.sql selesai")

	migrationsDir := filepath.Join(root, "database", "migrations")
	if env := os.Getenv("MIGRATIONS_PATH"); env != "" {
		if filepath.IsAbs(env) {
			migrationsDir = env
		} else {
			migrationsDir = filepath.Join(root, env)
		}
	}

	if err := RunMigrateCommand(databaseURL, migrationsDir, "force", []string{fmt.Sprint(LatestMigrationVersion)}); err != nil {
		return fmt.Errorf("force versi migrasi setelah init: %w", err)
	}
	log.Printf("schema_migrations diset ke %d\n", LatestMigrationVersion)
	return nil
}

func moduleRoot() (string, error) {
	if dir := os.Getenv("GIS_MODULE_ROOT"); dir != "" {
		return dir, nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("go.mod tidak ditemukan (jalankan dari root backend)")
	}
	dir = filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod tidak ditemukan")
		}
		dir = parent
	}
}
