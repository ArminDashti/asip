package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
	SSLMode  string
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode,
	)
}

type Config struct {
	Port               int
	DB                 DBConfig
	GinMode            string
	CORSAllowedOrigins []string

	SyncEnabled     bool
	SyncOnStartup   bool
	SyncHourUTC     int
	RepoAsMetadata  string
	RepoAsIPBlocks  string
	RepoGeoIPBlocks string
	URLAsMetadata   string
	URLAsIPBlocks   string
	URLGeoIPBlocks  string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port, err := strconv.Atoi(getEnv("PORT", "3000"))
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	syncHourUTC, err := strconv.Atoi(getEnv("SYNC_HOUR_UTC", "2"))
	if err != nil {
		return nil, fmt.Errorf("invalid SYNC_HOUR_UTC: %w", err)
	}
	if syncHourUTC < 0 || syncHourUTC > 23 {
		return nil, fmt.Errorf("SYNC_HOUR_UTC must be between 0 and 23")
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	return &Config{
		Port: port,
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			Username: getEnv("DB_USERNAME", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Database: getEnv("DB_NAME", "as_ip"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		GinMode:            getEnv("GIN_MODE", "debug"),
		CORSAllowedOrigins: parseCSVOrigins(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),
		SyncEnabled:        getEnvBool("SYNC_ENABLED", true),
		SyncOnStartup:      getEnvBool("SYNC_ON_STARTUP", false),
		SyncHourUTC:        syncHourUTC,
		RepoAsMetadata:     getEnv("REPO_AS_METADATA_PATH", `C:\Users\armin\Documents\GitHub\as-metadata`),
		RepoAsIPBlocks:     getEnv("REPO_AS_IP_BLOCKS_PATH", `C:\Users\armin\Documents\GitHub\as-ip-blocks`),
		RepoGeoIPBlocks:    getEnv("REPO_GEO_IP_BLOCKS_PATH", `C:\Users\armin\Documents\GitHub\geo-ip-blocks`),
		URLAsMetadata:      getEnv("REPO_AS_METADATA_URL", "https://github.com/ipverse/as-metadata"),
		URLAsIPBlocks:      getEnv("REPO_AS_IP_BLOCKS_URL", "https://github.com/ipverse/as-ip-blocks"),
		URLGeoIPBlocks:     getEnv("REPO_GEO_IP_BLOCKS_URL", "https://github.com/ipverse/geo-ip-blocks"),
	}, nil
}

func parseCSVOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin != "" {
			origins = append(origins, origin)
		}
	}
	if len(origins) == 0 {
		return []string{"http://localhost:5173"}
	}
	return origins
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
