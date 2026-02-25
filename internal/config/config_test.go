package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func setupTestConfig(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".config", "sgx")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", dir)
}

func TestLoadLegacyConfig(t *testing.T) {
	setupTestConfig(t, `{"api_key": "console-legacy123", "format": "table"}`)

	cfg, err := loadConfigFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.APIKey != "console-legacy123" {
		t.Errorf("expected legacy api_key, got %q", cfg.APIKey)
	}
	if cfg.Format != "table" {
		t.Errorf("expected format 'table', got %q", cfg.Format)
	}
	if cfg.Projects != nil {
		t.Error("expected nil projects for legacy config")
	}
}

func TestLoadMultiProjectConfig(t *testing.T) {
	setupTestConfig(t, `{
		"default_project": "prod",
		"projects": {
			"prod": {"api_key": "console-prod"},
			"staging": {"api_key": "console-stg", "format": "compact"}
		}
	}`)

	cfg, err := loadConfigFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(cfg.Projects))
	}
	if cfg.DefaultProject != "prod" {
		t.Errorf("expected default_project 'prod', got %q", cfg.DefaultProject)
	}
	if cfg.Projects["staging"].Format != "compact" {
		t.Errorf("expected staging format 'compact', got %q", cfg.Projects["staging"].Format)
	}
}

func TestResolveProject_ExplicitFlag(t *testing.T) {
	cfg := &Config{
		DefaultProject: "prod",
		Projects: map[string]*Project{
			"prod":    {APIKey: "console-prod"},
			"staging": {APIKey: "console-stg"},
		},
	}
	p := resolveProject(cfg, "staging")
	if p == nil || p.APIKey != "console-stg" {
		t.Errorf("expected staging project, got %+v", p)
	}
}

func TestResolveProject_Default(t *testing.T) {
	cfg := &Config{
		DefaultProject: "prod",
		Projects: map[string]*Project{
			"prod": {APIKey: "console-prod"},
		},
	}
	p := resolveProject(cfg, "")
	if p == nil || p.APIKey != "console-prod" {
		t.Errorf("expected default (prod) project, got %+v", p)
	}
}

func TestResolveProject_LegacyFallback(t *testing.T) {
	cfg := &Config{APIKey: "console-legacy"}
	p := resolveProject(cfg, "")
	if p == nil || p.APIKey != "console-legacy" {
		t.Errorf("expected legacy project, got %+v", p)
	}
}

func TestResolveProject_NotFound(t *testing.T) {
	cfg := &Config{
		Projects: map[string]*Project{
			"prod": {APIKey: "console-prod"},
		},
	}
	p := resolveProject(cfg, "nonexistent")
	if p != nil {
		t.Errorf("expected nil for nonexistent project, got %+v", p)
	}
}

func TestAddRemoveProject(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	// Add first project — should become default
	if err := AddProject("prod", "console-prod", "", ""); err != nil {
		t.Fatalf("AddProject failed: %v", err)
	}
	cfg, _ := loadConfigFile()
	if cfg.DefaultProject != "prod" {
		t.Errorf("expected default 'prod', got %q", cfg.DefaultProject)
	}
	if cfg.Projects["prod"].APIKey != "console-prod" {
		t.Errorf("expected api_key 'console-prod', got %q", cfg.Projects["prod"].APIKey)
	}

	// Add second project — default stays
	if err := AddProject("staging", "console-stg", "", "compact"); err != nil {
		t.Fatalf("AddProject failed: %v", err)
	}
	cfg, _ = loadConfigFile()
	if cfg.DefaultProject != "prod" {
		t.Errorf("expected default still 'prod', got %q", cfg.DefaultProject)
	}
	if len(cfg.Projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(cfg.Projects))
	}

	// Remove default — should reassign
	if err := RemoveProject("prod"); err != nil {
		t.Fatalf("RemoveProject failed: %v", err)
	}
	cfg, _ = loadConfigFile()
	if len(cfg.Projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(cfg.Projects))
	}
	if cfg.DefaultProject != "staging" {
		t.Errorf("expected default reassigned to 'staging', got %q", cfg.DefaultProject)
	}
}

func TestSetDefaultProject(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	AddProject("a", "key-a", "", "")
	AddProject("b", "key-b", "", "")

	if err := SetDefaultProject("b"); err != nil {
		t.Fatalf("SetDefaultProject failed: %v", err)
	}
	cfg, _ := loadConfigFile()
	if cfg.DefaultProject != "b" {
		t.Errorf("expected default 'b', got %q", cfg.DefaultProject)
	}
}

func TestSetDefaultProject_NotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	AddProject("a", "key-a", "", "")

	if err := SetDefaultProject("nonexistent"); err == nil {
		t.Error("expected error for nonexistent project")
	}
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"console-abc123xyz", "console-***3xyz"},
		{"short", "***"},
		{"exactly10!", "***"},
		{"console-W6q08wuABCDEF", "console-***CDEF"},
	}
	for _, tc := range tests {
		got := MaskKey(tc.input)
		if got != tc.want {
			t.Errorf("MaskKey(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestSaveConfigOmitsEmptyFields(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	AddProject("prod", "console-prod", "", "")

	path := filepath.Join(dir, ".config", "sgx", "config.json")
	data, _ := os.ReadFile(path)
	var raw map[string]json.RawMessage
	json.Unmarshal(data, &raw)

	// Legacy api_key should not appear when using projects
	if _, ok := raw["api_key"]; ok {
		t.Error("expected api_key to be omitted from config with projects")
	}
}
