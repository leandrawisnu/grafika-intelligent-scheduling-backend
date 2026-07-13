package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL  string
	MLServiceURL string
	ServerPort   string
}

func Load() *Config {
	// Cari .env di current dir, parent dir, atau project root
	paths := []string{
		".env",
		"../.env",
		filepath.Join(os.Getenv("HOME"), ".env"),
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			if err := godotenv.Load(p); err != nil {
				log.Printf("Gagal load %s: %v", p, err)
			}
			break
		}
	}

	return &Config{
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://grafika:grafika_dev@localhost:5432/grafika?sslmode=disable"),
		MLServiceURL: getEnv("ML_SERVICE_URL", "http://localhost:8000"),
		ServerPort:   getEnv("SERVER_PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
