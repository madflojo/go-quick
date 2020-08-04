package config

import (
	"os"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	cfg := New()

	if cfg.ListenAddr != "0.0.0.0:8443" {
		t.Errorf("Unexpected value for Listen Address - %s", cfg.ListenAddr)
	}

	if cfg.EnableTLS != true {
		t.Errorf("Unexpected value for Enabling TLS - %t", cfg.EnableTLS)
	}

	if cfg.Debug != false {
		t.Errorf("Unexpected value for Debug - %t", cfg.Debug)
	}
}

func TestConfigFromEnv(t *testing.T) {
	// Save defaults
	def := os.Getenv("LISTEN_ADDR")

	// Setup some not so true values
	os.Setenv("LISTEN_ADDR", "localhost:9000")

	cfg, err := NewFromEnv()
	if err != nil {
		t.Errorf("Error pulling config from environment")
	}

	if cfg.ListenAddr != "localhost:9000" {
		t.Errorf("Invalid Listen Address - %s", cfg.ListenAddr)
	}

	// Reset Default
	os.Setenv("LISTEN_ADDR", def)
}
