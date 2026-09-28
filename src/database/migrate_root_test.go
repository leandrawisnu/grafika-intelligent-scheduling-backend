package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestModuleRootTerimaMigrasiTanpaGoMod(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "database", "migrations"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIS_MODULE_ROOT", dir)

	got, err := moduleRoot()
	if err != nil {
		t.Fatal(err)
	}
	if got != dir {
		t.Fatalf("root = %s, ingin %s", got, dir)
	}
}

func TestModuleRootAbaikanRootKosong(t *testing.T) {
	t.Setenv("GIS_MODULE_ROOT", t.TempDir())

	got, err := moduleRoot()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(got, "go.mod")); err != nil {
		t.Fatalf("root %s bukan modul backend: %v", got, err)
	}
}
