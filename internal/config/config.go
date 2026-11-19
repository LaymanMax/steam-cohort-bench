package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Config holds runtime configuration for the application.
type Config struct {
	Port                 int
	Host                 string
	DataDir              string
	RawParquetPath       string
	ProcessedParquetPath string
	AnalyticsDBPath      string
	ManifestPath         string
	CacheTTL             time.Duration
	MaxCacheItems        int
	StaticDir            string
	LogLevel             string
}

// Load loads configuration from environment variables with sensible defaults.
func Load() *Config {
	port := getEnvInt("PORT", 8080)
	host := getEnv("HOST", "0.0.0.0")
	dataDir := getEnv("DATA_DIR", "./data")
	logLevel := getEnv("LOG_LEVEL", "info")
	staticDir := getEnv("STATIC_DIR", "./web/dist")

	cacheTTLSeconds := getEnvInt("CACHE_TTL_SECONDS", 600)
	maxCacheItems := getEnvInt("MAX_CACHE_ITEMS", 500)

	return &Config{
		Port:                 port,
		Host:                 host,
		DataDir:              dataDir,
		RawParquetPath:       getEnv("RAW_PARQUET_PATH", filepath.Join(dataDir, "raw", "steam-games.parquet")),
		ProcessedParquetPath: getEnv("PROCESSED_PARQUET_PATH", filepath.Join(dataDir, "processed", "canonical_games.parquet")),
		AnalyticsDBPath:      getEnv("ANALYTICS_DB_PATH", filepath.Join(dataDir, "processed", "analytics.duckdb")),
		ManifestPath:         getEnv("MANIFEST_PATH", filepath.Join(dataDir, "manifests", "manifest.json")),
		CacheTTL:             time.Duration(cacheTTLSeconds) * time.Second,
		MaxCacheItems:        maxCacheItems,
		StaticDir:            staticDir,
		LogLevel:             logLevel,
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
