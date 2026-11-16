package config_test

import (
	"testing"

	"github.com/mjengofiti/api/internal/config"
)

func TestConfigLoadDefaults(t *testing.T) {
	t.Setenv("APP_PORT", "")
	t.Setenv("APP_ENV", "development")
	t.Setenv("AUTH_ENABLED", "false")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected config to load successfully, got: %v", err)
	}
	if cfg.App.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.App.Port)
	}
	if cfg.App.Env != "development" {
		t.Errorf("expected default env development, got %s", cfg.App.Env)
	}
	if cfg.Map.DefaultZoom != 11 {
		t.Errorf("expected default zoom 11, got %d", cfg.Map.DefaultZoom)
	}
}

func TestConfigCustomPort(t *testing.T) {
	t.Setenv("APP_PORT", "9090")
	t.Setenv("APP_ENV", "development")
	t.Setenv("AUTH_ENABLED", "false")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("expected config to load with custom port, got: %v", err)
	}
	if cfg.App.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.App.Port)
	}
}

func TestProductionRequiresAuthentication(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_ENABLED", "false")

	if _, err := config.Load(); err == nil {
		t.Fatal("expected production configuration without authentication to fail")
	}
}

func TestEnabledAuthenticationRequiresStrongCompleteConfiguration(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_ISSUER_URL", "https://identity.example.test")
	t.Setenv("AUTH_AUDIENCE", "mjengofiti-api")
	t.Setenv("APP_CORS_ORIGINS", "https://app.example.test")
	t.Setenv("DB_PASSWORD", "local-test-password")
	t.Setenv("DB_SSLMODE", "require")

	for _, secret := range []string{"short", "CHANGE_ME_GENERATE_RANDOM_SECRET", "0123456789abcdef0123456789abcdef"} {
		t.Setenv("JWT_SECRET", secret)
		_, err := config.Load()
		if secret != "0123456789abcdef0123456789abcdef" && err == nil {
			t.Fatalf("expected weak or placeholder JWT secret to be rejected")
		}
		if secret == "0123456789abcdef0123456789abcdef" && err != nil {
			t.Fatalf("expected complete auth config to load, got: %v", err)
		}
	}
}

func TestProductionRequiresEncryptedDatabaseAndHTTPSOrigins(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("AUTH_ISSUER_URL", "https://identity.example.test")
	t.Setenv("AUTH_AUDIENCE", "mjengofiti-api")
	t.Setenv("DB_PASSWORD", "local-test-password")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("APP_CORS_ORIGINS", "https://app.example.test")

	if _, err := config.Load(); err != nil {
		t.Fatalf("expected secure production configuration to load, got: %v", err)
	}

	t.Setenv("DB_SSLMODE", "disable")
	if _, err := config.Load(); err == nil {
		t.Fatal("expected production database TLS to be required")
	}

	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("APP_CORS_ORIGINS", "http://app.example.test")
	if _, err := config.Load(); err == nil {
		t.Fatal("expected production CORS origins to require HTTPS")
	}
}

func TestEnabledAuthenticationRequiresHTTPSIssuerAndAudience(t *testing.T) {
	t.Setenv("APP_ENV", "development")
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("AUTH_ISSUER_URL", "http://identity.example.test")
	t.Setenv("AUTH_AUDIENCE", "mjengofiti-api")

	if _, err := config.Load(); err == nil {
		t.Fatal("expected non-HTTPS issuer to be rejected")
	}

	t.Setenv("AUTH_ISSUER_URL", "https://identity.example.test")
	t.Setenv("AUTH_AUDIENCE", "")
	if _, err := config.Load(); err == nil {
		t.Fatal("expected missing audience to be rejected")
	}
}
