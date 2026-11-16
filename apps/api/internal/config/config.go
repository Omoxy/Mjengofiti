package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	App        AppConfig
	Database   DatabaseConfig
	Auth       AuthConfig
	Map        MapConfig
	Ingestion  IngestionConfig
	RateLimit  RateLimitConfig
}

// AppConfig holds general application settings.
type AppConfig struct {
	Env         string
	Port        int
	LogLevel    string
	CORSOrigins []string
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
	MaxConns int
	MinConns int
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	Enabled   bool
	IssuerURL string
	ClientID  string
	Audience  string
	JWTSecret string
}

// MapConfig holds map display defaults.
type MapConfig struct {
	TileURL       string
	DefaultCenterLng float64
	DefaultCenterLat float64
	DefaultZoom      int
}

// IngestionConfig holds data ingestion settings.
type IngestionConfig struct {
	MaxFileSizeMB int
	TempDir       string
	Workers       int
}

// RateLimitConfig holds rate limiting settings.
type RateLimitConfig struct {
	RequestsPerMinute int
	Burst             int
}

// DSN returns a PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

// Load reads configuration from environment variables with sensible defaults.
// It also checks for a local .env file in the current or parent directories.
func Load() (*Config, error) {
	loadDotEnv()

	cfg := &Config{
		App: AppConfig{
			Env:         strings.ToLower(envOrDefault("APP_ENV", "development")),
			Port:        envIntOrDefault("APP_PORT", 8080),
			LogLevel:    envOrDefault("APP_LOG_LEVEL", "info"),
			CORSOrigins: strings.Split(envOrDefault("APP_CORS_ORIGINS", "http://localhost:5173"), ","),
		},
		Database: DatabaseConfig{
			Host:     envOrDefault("DB_HOST", "localhost"),
			Port:     envIntOrDefault("DB_PORT", 5432),
			Name:     envOrDefault("DB_NAME", "mjengofiti"),
			User:     envOrDefault("DB_USER", "mjengofiti"),
			Password: envOrDefault("DB_PASSWORD", ""),
			SSLMode:  envOrDefault("DB_SSLMODE", "disable"),
			MaxConns: envIntOrDefault("DB_MAX_CONNS", 25),
			MinConns: envIntOrDefault("DB_MIN_CONNS", 5),
		},
		Auth: AuthConfig{
			Enabled:   envBoolOrDefault("AUTH_ENABLED", false),
			IssuerURL: envOrDefault("AUTH_ISSUER_URL", ""),
			ClientID:  envOrDefault("AUTH_CLIENT_ID", ""),
			Audience:  envOrDefault("AUTH_AUDIENCE", ""),
			JWTSecret: envOrDefault("JWT_SECRET", ""),
		},
		Map: MapConfig{
			TileURL:          envOrDefault("MAP_TILE_URL", "https://api.maptiler.com/maps/streets-v2/style.json"),
			DefaultCenterLng: envFloatOrDefault("MAP_DEFAULT_CENTER_LNG", 36.8219),
			DefaultCenterLat: envFloatOrDefault("MAP_DEFAULT_CENTER_LAT", -1.2921),
			DefaultZoom:      envIntOrDefault("MAP_DEFAULT_ZOOM", 11),
		},
		Ingestion: IngestionConfig{
			MaxFileSizeMB: envIntOrDefault("INGESTION_MAX_FILE_SIZE_MB", 500),
			TempDir:       envOrDefault("INGESTION_TEMP_DIR", "./tmp/ingestion"),
			Workers:       envIntOrDefault("INGESTION_WORKERS", 2),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: envIntOrDefault("RATE_LIMIT_REQUESTS_PER_MINUTE", 60),
			Burst:             envIntOrDefault("RATE_LIMIT_BURST", 10),
		},
	}

	if cfg.Auth.Enabled {
		secretValue := strings.ToLower(strings.TrimSpace(cfg.Auth.JWTSecret))
		if len([]byte(cfg.Auth.JWTSecret)) < 32 || strings.Contains(secretValue, "change_me") || strings.Contains(secretValue, "placeholder") || strings.Contains(secretValue, "your_") || strings.Contains(secretValue, "replace_me") {
			return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters when AUTH_ENABLED is true")
		}
		if cfg.Auth.IssuerURL == "" || !strings.HasPrefix(strings.ToLower(cfg.Auth.IssuerURL), "https://") {
			return nil, fmt.Errorf("AUTH_ISSUER_URL must be an HTTPS issuer when AUTH_ENABLED is true")
		}
		if cfg.Auth.Audience == "" {
			return nil, fmt.Errorf("AUTH_AUDIENCE is required when AUTH_ENABLED is true")
		}
	}
	if (cfg.App.Env == "production" || cfg.App.Env == "staging") && !cfg.Auth.Enabled {
		return nil, fmt.Errorf("AUTH_ENABLED must be true in %s", cfg.App.Env)
	}
	if cfg.App.Env == "production" || cfg.App.Env == "staging" {
		if cfg.Database.Password == "" {
			return nil, fmt.Errorf("DB_PASSWORD is required in %s", cfg.App.Env)
		}
		if strings.EqualFold(cfg.Database.SSLMode, "disable") {
			return nil, fmt.Errorf("DB_SSLMODE must enable TLS in %s", cfg.App.Env)
		}
		for _, rawOrigin := range cfg.App.CORSOrigins {
			origin := strings.TrimSpace(rawOrigin)
			parsed, err := url.Parse(origin)
			if err != nil || parsed.Scheme != "https" || parsed.Host == "" || (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.Fragment != "" {
				return nil, fmt.Errorf("APP_CORS_ORIGINS must contain only HTTPS origins in %s", cfg.App.Env)
			}
		}
	}

	return cfg, nil
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envIntOrDefault(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}

func envFloatOrDefault(key string, defaultVal float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return defaultVal
}

func envBoolOrDefault(key string, defaultVal bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultVal
}

func loadDotEnv() {
	candidates := []string{".env", "../.env", "../../.env"}
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			v = strings.Trim(v, `"'`)
			if os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
		break
	}
}
