package config

import (
	"os"
	"testing"
)

func TestNewConfigValidFile(t *testing.T) {
	content := `{
		"server": {"address": "127.0.0.1", "port": 8053},
		"cache": {"default_ttl": 300},
		"logging": {"level": "info"}
	}`
	tmpFile := createTempFile(t, content)
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	cfg, err := NewConfig(tmpFile)
	if err != nil {
		t.Fatalf("NewConfig returned error: %v", err)
	}

	if cfg.Server.Address != "127.0.0.1" {
		t.Errorf("Expected server address '127.0.0.1', got '%s'", cfg.Server.Address)
	}
	if cfg.Server.Port != 8053 {
		t.Errorf("Expected server port 8053, got %d", cfg.Server.Port)
	}
	if cfg.Cache.DefaultTTL != 300 {
		t.Errorf("Expected default TTL 300, got %d", cfg.Cache.DefaultTTL)
	}
}

func TestNewConfigInvalidFile(t *testing.T) {
	_, err := NewConfig("nonexistent_file.json")
	if err == nil {
		t.Fatalf("Expected error for nonexistent file, got nil")
	}
}

func TestNewConfigInvalidJSON(t *testing.T) {
	content := `{
		"server": {"address": "127.0.0.1", "port": "not_an_int"},
		"cache": {"default_ttl": 300},
		"logging": {"level": "info"}
	}`
	tmpFile := createTempFile(t, content)
	defer func() {
		_ = os.Remove(tmpFile)
	}()

	_, err := NewConfig(tmpFile)
	if err == nil {
		t.Fatalf("Expected error for invalid JSON, got nil")
	}
}

func createTempFile(t testing.TB, content string) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "config*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	func() {
		_ = tmpFile.Close()
	}()
	return tmpFile.Name()
}
