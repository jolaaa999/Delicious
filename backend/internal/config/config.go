package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	MySQLDSN      string
	JWTSecret     string
	DefaultUID    uint64
	AutoMigrate   bool
	UploadDir     string
	PublicBaseURL string
	MaxUploadMB   int64
}

func Load() Config {
	uid, _ := strconv.ParseUint(getEnv("DEFAULT_USER_ID", "1"), 10, 64)
	maxMB, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_MB", "10"), 10, 64)
	return Config{
		Port:          getEnv("PORT", "8080"),
		MySQLDSN:      getEnv("MYSQL_DSN", "delicious:delicious@tcp(127.0.0.1:3306)/delicious?charset=utf8mb4&parseTime=True&loc=Local"),
		JWTSecret:     getEnv("JWT_SECRET", "delicious-dev-secret"),
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
