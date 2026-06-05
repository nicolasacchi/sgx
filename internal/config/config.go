package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	cliconfig "github.com/nicolasacchi/clicore/config"
)

type Project struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url,omitempty"`
	Format  string `json:"format,omitempty"`
}

type Config struct {
	// Legacy fields (backward compat with flat config)
	APIKey  string `json:"api_key,omitempty"`
	BaseURL string `json:"base_url,omitempty"`
	Format  string `json:"format,omitempty"`

	// Multi-project fields
	DefaultProject string              `json:"default_project,omitempty"`
	Projects       map[string]*Project `json:"projects,omitempty"`
}

func configFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "sgx", "config.json"), nil
}

func loadConfigFile() (*Config, error) {
	path, err := configFilePath()
	if err != nil {
		return nil, err
	}
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

func saveConfigFile(cfg *Config) error {
	path, err := configFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	// Atomic temp+rename (clicore) — replaces os.WriteFile (which truncates the
	// live file first) so an interrupted write can't corrupt a config that
	// already holds credentials.
	return cliconfig.WriteFileAtomic(path, append(data, '\n'), 0600)
}

func resolveProject(cfg *Config, projectFlag string) *Project {
	if cfg == nil {
		return nil
	}
	if projectFlag != "" && cfg.Projects != nil {
		if p, ok := cfg.Projects[projectFlag]; ok {
			return p
		}
		return nil
	}
	if cfg.DefaultProject != "" && cfg.Projects != nil {
		if p, ok := cfg.Projects[cfg.DefaultProject]; ok {
			return p
		}
	}
	if cfg.APIKey != "" {
		return &Project{
			APIKey:  cfg.APIKey,
			BaseURL: cfg.BaseURL,
			Format:  cfg.Format,
		}
	}
	return nil
}

func LoadAPIKey(flagValue, projectFlag string) (string, error) {
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
	if err == nil {
		if p := resolveProject(cfg, projectFlag); p != nil && p.APIKey != "" {
			return p.APIKey, nil
		}
	}
	return "", fmt.Errorf("API key required: use --api-key flag, STATSIG_API_KEY/STATSIG_CONSOLE_KEY env var, or ~/.config/sgx/config.json")
}

func LoadBaseURL(flagValue, projectFlag string) string {
	if flagValue != "" {
		return flagValue
	}
	if v := os.Getenv("SGX_BASE_URL"); v != "" {
		return v
	}
	cfg, err := loadConfigFile()
	if err == nil {
		if p := resolveProject(cfg, projectFlag); p != nil && p.BaseURL != "" {
			return p.BaseURL
		}
		if cfg.BaseURL != "" {
			return cfg.BaseURL
		}
	}
	return "https://statsigapi.net"
}

func LoadFormat(flagValue, projectFlag string) string {
	if flagValue != "" {
		return flagValue
	}
	if v := os.Getenv("SGX_FORMAT"); v != "" {
		return v
	}
	cfg, err := loadConfigFile()
	if err == nil {
		if p := resolveProject(cfg, projectFlag); p != nil && p.Format != "" {
			return p.Format
		}
		if cfg.Format != "" {
			return cfg.Format
		}
	}
	return "json"
}

func AddProject(name, apiKey, baseURL, format string) error {
	cfg, err := loadConfigFile()
	if err != nil {
		cfg = &Config{}
	}
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]*Project)
	}
	cfg.Projects[name] = &Project{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Format:  format,
	}
	if cfg.DefaultProject == "" {
		cfg.DefaultProject = name
	}
	return saveConfigFile(cfg)
}

func RemoveProject(name string) error {
	cfg, err := loadConfigFile()
	if err != nil {
		return fmt.Errorf("no config file found")
	}
	if cfg.Projects == nil {
		return fmt.Errorf("project %q not found", name)
	}
	if _, ok := cfg.Projects[name]; !ok {
		return fmt.Errorf("project %q not found", name)
	}
	delete(cfg.Projects, name)
	if cfg.DefaultProject == name {
		cfg.DefaultProject = ""
		for k := range cfg.Projects {
			cfg.DefaultProject = k
			break
		}
	}
	if len(cfg.Projects) == 0 {
		cfg.Projects = nil
	}
	return saveConfigFile(cfg)
}

func SetDefaultProject(name string) error {
	cfg, err := loadConfigFile()
	if err != nil {
		return fmt.Errorf("no config file found")
	}
	if cfg.Projects == nil {
		return fmt.Errorf("project %q not found", name)
	}
	if _, ok := cfg.Projects[name]; !ok {
		return fmt.Errorf("project %q not found", name)
	}
	cfg.DefaultProject = name
	return saveConfigFile(cfg)
}

func ListProjects() (*Config, error) {
	return loadConfigFile()
}

func MaskKey(key string) string {
	if len(key) <= 10 {
		return "***"
	}
	return key[:8] + "***" + key[len(key)-4:]
}
