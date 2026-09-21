package config

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv       string
	AppHost      string
	AppPort      int
	DBHost       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBPort       int
	databaseURL  string // set when DATABASE_URL provided explicitly
	MLServiceURL string
	AutoMigrate  bool
}

func Load() *Config {
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

	appPort := getEnvInt("APP_PORT", 0)
	if appPort == 0 {
		appPort = getEnvInt("SERVER_PORT", 8080)
	}

	cfg := &Config{
		AppEnv:       getEnv("APP_ENV", "local"),
		AppHost:      getEnv("APP_HOST", "0.0.0.0"),
		AppPort:      appPort,
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBUser:       getEnv("DB_USER", "grafika"),
		DBPassword:   getEnv("DB_PASSWORD", "grafika_dev"),
		DBName:       getEnv("DB_NAME", "grafika"),
		DBPort:       getEnvInt("DB_PORT", 5432),
		databaseURL:  os.Getenv("DATABASE_URL"),
		MLServiceURL: getEnv("ML_SERVICE_URL", "http://localhost:8000"),
		AutoMigrate:  getEnvBool("GIS_AUTO_MIGRATE", false),
	}
	return cfg
}

// DatabaseURL returns postgres URL from DATABASE_URL or DB_* (format sama pola KAI BIOP).
func (c *Config) DatabaseURL() string {
	if c.databaseURL != "" {
		return c.databaseURL
	}
	port := c.DBPort
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.DBUser, c.DBPassword),
		Host:   fmt.Sprintf("%s:%d", c.DBHost, port),
		Path:   "/" + c.DBName,
	}
	q := u.Query()
	q.Set("sslmode", "disable")
	u.RawQuery = q.Encode()
	return u.String()
}

func (c *Config) ListenAddr() string {
	return fmt.Sprintf("%s:%d", c.AppHost, c.AppPort)
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "TRUE", "yes", "YES":
		return true
	case "0", "false", "FALSE", "no", "NO":
		return false
	default:
		return fallback
	}
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
