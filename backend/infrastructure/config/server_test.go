package config_test

import (
	"testing"

	"github.com/Tattsum/translate-prompt/backend/infrastructure/config"
)

func TestLoadServerFromEnv_Defaults(t *testing.T) {
	t.Setenv("LISTEN_HOST", "")
	t.Setenv("PORT", "")
	t.Setenv("ALLOWED_ORIGINS", "")
	t.Setenv("INVESTIGATE_ENABLED", "")

	cfg := config.LoadServerFromEnv(8080)
	if cfg.ListenHost != "127.0.0.1" {
		t.Fatalf("ListenHost = %q, want 127.0.0.1", cfg.ListenHost)
	}
	if cfg.Port != 8080 {
		t.Fatalf("Port = %d, want 8080", cfg.Port)
	}
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "*" {
		t.Fatalf("AllowedOrigins = %v, want [*]", cfg.AllowedOrigins)
	}
	if !cfg.InvestigateEnabled {
		t.Fatal("InvestigateEnabled = false, want true")
	}
}

func TestLoadServerFromEnv_Production(t *testing.T) {
	t.Setenv("LISTEN_HOST", "0.0.0.0")
	t.Setenv("PORT", "9090")
	t.Setenv("ALLOWED_ORIGINS", "https://translate.tattsum.com")
	t.Setenv("INVESTIGATE_ENABLED", "false")

	cfg := config.LoadServerFromEnv(8080)
	if cfg.ListenHost != "0.0.0.0" {
		t.Fatalf("ListenHost = %q, want 0.0.0.0", cfg.ListenHost)
	}
	if cfg.Port != 9090 {
		t.Fatalf("Port = %d, want 9090", cfg.Port)
	}
	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "https://translate.tattsum.com" {
		t.Fatalf("AllowedOrigins = %v", cfg.AllowedOrigins)
	}
	if cfg.InvestigateEnabled {
		t.Fatal("InvestigateEnabled = true, want false")
	}
}
