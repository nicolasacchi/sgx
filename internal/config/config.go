package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	Format  string `json:"format"`
}

func LoadAPIKey(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}

	if v := os.Getenv("STATSIG_API_KEY"); v != "" {
		return v, nil
	}
	if v := os.Getenv("STATSIG_CONSOLE_KEY"); v != "" {
		return v, nil
	}

	cfg, err := loadConfigFile()
	if err == nil && cfg.APIKey != "" {
		return cfg.APIKey, nil
	}

	return "", fmt.Errorf("API key required: use --api-key flag, STATSIG_API_KEY/STATSIG_CONSOLE_KEY env var, or ~/.config/sgx/config.json")
}

func LoadBaseURL(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if v := os.Getenv("SGX_BASE_URL"); v != "" {
		return v
	}
	cfg, err := loadConfigFile()
	if err == nil && cfg.BaseURL != "" {
		return cfg.BaseURL
	}
	return "https://statsigapi.net"
}

func LoadFormat(flagValue string) string {
	if flagValue != "" {
		return flagValue
	}
	if v := os.Getenv("SGX_FORMAT"); v != "" {
		return v
	}
	cfg, err := loadConfigFile()
	if err == nil && cfg.Format != "" {
		return cfg.Format
	}
	return "json"
}

func loadConfigFile() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".config", "sgx", "config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid config file %s: %w", path, err)
	}
	return &cfg, nil
}
