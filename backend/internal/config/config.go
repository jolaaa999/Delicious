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
	BlobToken           string
	UseBlob             bool
	SpoonacularAPIKey   string
	OnlineSearchEnabled bool
}

func Load() Config {
	uid, _ := strconv.ParseUint(getEnv("DEFAULT_USER_ID", "1"), 10, 64)
	maxMB, _ := strconv.ParseInt(getEnv("MAX_UPLOAD_MB", "10"), 10, 64)
	onVercel := os.Getenv("VERCEL") != ""

	dbURL := firstEnv(
		"DATABASE_URL",
		"POSTGRES_URL",
		"PRISMA_DATABASE_URL",
		"POSTGRES_PRISMA_URL",
		"POSTGRES_URL_NON_POOLING",
	)
	if dbURL == "" && !onVercel {
		dbURL = "postgres://delicious:delicious@127.0.0.1:5432/delicious?sslmode=disable"
	}

	blobToken := os.Getenv("BLOB_READ_WRITE_TOKEN")
	publicBase := getEnv("PUBLIC_BASE_URL", "")
	if publicBase == "" {
		if host := os.Getenv("VERCEL_URL"); host != "" {
			publicBase = "https://" + host
		}
	}

	uploadDir := getEnv("UPLOAD_DIR", "./uploads")
	if blobToken != "" || onVercel {
		uploadDir = ""
	}

	return Config{
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         dbURL,
		DefaultUID:          uid,
		AutoMigrate:         getEnv("AUTO_MIGRATE", "true") == "true",
		UploadDir:           uploadDir,
		PublicBaseURL:       publicBase,
		MaxUploadMB:         maxMB,
		BlobToken:           blobToken,
		UseBlob:             blobToken != "",
		SpoonacularAPIKey:   os.Getenv("SPOONACULAR_API_KEY"),
		OnlineSearchEnabled: getEnv("ENCYCLOPEDIA_ONLINE_SEARCH", "true") == "true",
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if v := os.Getenv(key); v != "" {
			return v
		}
	}
	return ""
}
