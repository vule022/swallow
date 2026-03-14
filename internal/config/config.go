package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds all swallow configuration.
type Config struct {
	Provider       string  `json:"provider"`
	Model          string  `json:"model"`
	BaseURL        string  `json:"base_url"`
	DefaultProject string  `json:"default_project"`
	CopyMode       string  `json:"copy_mode"`
	StorageBackend string  `json:"storage_backend"`
	MaxTokens      int     `json:"max_tokens"`
	Temperature    float64 `json:"temperature"`
}

// Dir returns the path to the swallow config directory (~/.swallow).
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: cannot find home directory: %w", err)
	}
	return filepath.Join(home, ConfigDirName), nil
}

// Path returns the full path to config.json.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

// DBPath returns the full path to swallow.db.
func DBPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, DBFileName), nil
}

// Defaults returns a Config populated with default values.
func Defaults() *Config {
	return &Config{
		Provider:       DefaultProvider,
		Model:          DefaultModel,
		BaseURL:        DefaultBaseURL,
		CopyMode:       DefaultCopyMode,
		StorageBackend: "sqlite",
		MaxTokens:      DefaultMaxTokens,
		Temperature:    DefaultTemperature,
	}
}

// Load reads config from disk, falling back to defaults for missing fields.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	cfg := Defaults()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}

	// Apply defaults for zero-value fields.
	if cfg.Provider == "" {
		cfg.Provider = DefaultProvider
	}
	if cfg.Model == "" {
		cfg.Model = DefaultModel
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = DefaultMaxTokens
	}
	if cfg.Temperature == 0 {
		cfg.Temperature = DefaultTemperature
	}
	if cfg.CopyMode == "" {
		cfg.CopyMode = DefaultCopyMode
	}

	return cfg, nil
}

// Save writes the config to disk, creating the directory if needed.
func Save(cfg *Config) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("config: create dir %s: %w", dir, err)
	}

	path := filepath.Join(dir, ConfigFileName)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("config: write %s: %w", path, err)
	}
	return nil
}

// APIKey returns the API key from the environment.
func APIKey() string {
	return os.Getenv(EnvAPIKey)
}

// Validate checks that the config is usable for planning operations.
func (c *Config) Validate() error {
	if c.Provider == "" {
		return fmt.Errorf("config: provider is required")
	}
	if c.Model == "" {
		return fmt.Errorf("config: model is required")
	}
	if c.BaseURL == "" {
		return fmt.Errorf("config: base_url is required")
	}
	return nil
}
