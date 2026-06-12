package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	DatabaseURL   string
	DefaultUID    uint64
	AutoMigrate   bool
	UploadDir     string
	PublicBaseURL string
	MaxUploadMB   int64
}

func Load() Config {
	uid, _ := strconv.ParseUint(getEnv("DEFAULT_USER_ID", "1"), 10, 64)
	maxMB, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_MB", "10"), 10, 64)

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = os.Getenv("POSTGRES_URL")
	}
	if dbURL == "" {
		// 本地 Docker Postgres 默认
		dbURL = "postgres://delicious:delicious@127.0.0.1:5432/delicious?sslmode=disable"
	}

	return Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   dbURL,
		DefaultUID:    uid,
		AutoMigrate:   getEnv("AUTO_MIGRATE", "true") == "true",
		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
		PublicBaseURL: getEnv("PUBLIC_BASE_URL", ""),
		MaxUploadMB:   maxMB,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
