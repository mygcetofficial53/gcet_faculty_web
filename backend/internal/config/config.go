package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Port               int
	GMSPortalURL       string
	JWTSecret          string
	JWTExpiry          time.Duration
	JWTRefreshExp      time.Duration
	SupabaseURL        string
	SupabaseAnonKey    string
	SupabaseServiceKey string
	CORSOrigins        []string
	RateLimitRPS       int
	LogLevel           string
	Environment        string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (non-fatal if missing)
	_ = godotenv.Load()

	cfg := &Config{
		Port:               getEnvInt("PORT", 8080),
		GMSPortalURL:       getEnv("GMS_PORTAL_URL", "http://202.129.240.148:8080/GIS"),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTExpiry:          getEnvDuration("JWT_EXPIRY", 24*time.Hour),
		JWTRefreshExp:      getEnvDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
		SupabaseURL:        getEnv("SUPABASE_URL", "https://fczygwztaxunpselkpdl.supabase.co"),
		SupabaseAnonKey:    getEnv("SUPABASE_ANON_KEY", ""),
		SupabaseServiceKey: getEnv("SUPABASE_SERVICE_KEY", ""),
		CORSOrigins:        getEnvSlice("CORS_ORIGINS", []string{"http://localhost:3000"}),
		RateLimitRPS:       getEnvInt("RATE_LIMIT_RPS", 10),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		Environment:        getEnv("ENVIRONMENT", "development"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if val := os.Getenv(key); val != "" {
		result := []string{}
		for _, s := range splitAndTrim(val) {
			if s != "" {
				result = append(result, s)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return fallback
}

func splitAndTrim(s string) []string {
	parts := []string{}
	current := ""
	for _, c := range s {
		if c == ',' {
			parts = append(parts, current)
			current = ""
		} else if c != ' ' {
			current += string(c)
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	trimmed := make([]string, len(parts))
	for i, p := range parts {
		trimmed[i] = trimString(p)
	}
	return trimmed
}

func trimString(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
