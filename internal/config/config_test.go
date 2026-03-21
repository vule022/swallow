package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()
	if cfg.Provider != DefaultProvider {
		t.Errorf("Provider = %q, want %q", cfg.Provider, DefaultProvider)
	}
	if cfg.Model != DefaultModel {
		t.Errorf("Model = %q, want %q", cfg.Model, DefaultModel)
	}
	if cfg.MaxTokens != DefaultMaxTokens {
		t.Errorf("MaxTokens = %d, want %d", cfg.MaxTokens, DefaultMaxTokens)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	// Point home at a temp dir with no config file.
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Provider != DefaultProvider {
		t.Errorf("Provider = %q, want %q", cfg.Provider, DefaultProvider)
	}
}

func TestLoad_ExistingFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, ConfigDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}

	data, _ := json.Marshal(&Config{
		Provider:    "anthropic",
		Model:       "claude-3-5-sonnet-20241022",
		BaseURL:     "https://api.anthropic.com/v1",
		MaxTokens:   2048,
		Temperature: 0.5,
	})
	if err := os.WriteFile(filepath.Join(dir, ConfigFileName), data, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Provider != "anthropic" {
		t.Errorf("Provider = %q, want %q", cfg.Provider, "anthropic")
	}
	if cfg.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("Model = %q", cfg.Model)
	}
	if cfg.MaxTokens != 2048 {
		t.Errorf("MaxTokens = %d, want 2048", cfg.MaxTokens)
	}
}

func TestLoad_AppliesDefaultsForZeroFields(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	dir := filepath.Join(tmp, ConfigDirName)
	os.MkdirAll(dir, 0o700)

	// Write config with only provider set.
	data := []byte(`{"provider":"custom"}`)
	os.WriteFile(filepath.Join(dir, ConfigFileName), data, 0o600)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Model != DefaultModel {
		t.Errorf("Model = %q, want default %q", cfg.Model, DefaultModel)
	}
	if cfg.MaxTokens != DefaultMaxTokens {
		t.Errorf("MaxTokens = %d, want default %d", cfg.MaxTokens, DefaultMaxTokens)
	}
}

func TestValidate(t *testing.T) {
	cfg := Defaults()
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate() on defaults error: %v", err)
	}

	cfg.Model = ""
	if err := cfg.Validate(); err == nil {
		t.Error("Validate() expected error for empty Model")
	}
}
