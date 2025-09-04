package yamlenv

import (
	"embed"
	"testing"
	"time"
)

//go:embed testdata/embed_config.yaml testdata/embed_config.local.yaml
var embedFS embed.FS

type EmbedConfig struct {
	App struct {
		Name    string `yaml:"name"`
		Port    int    `yaml:"port"`
		Enabled bool   `yaml:"enabled"`
	} `yaml:"app"`
	DB struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Username string `yaml:"username"`
	} `yaml:"db"`
	Timeout time.Duration `yaml:"timeout"`
}

func TestLoadConfig_EmbedFS_BaseOnly(t *testing.T) {
	var cfg EmbedConfig
	
	err := LoadConfig(LoaderOptions{
		BaseSource: EmbedSource(embedFS, "testdata/embed_config.yaml"),
		Target:     &cfg,
	})
	
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	
	if cfg.App.Name != "embed-app" {
		t.Errorf("expected app name 'embed-app', got '%s'", cfg.App.Name)
	}
	if cfg.App.Port != 8080 {
		t.Errorf("expected app port 8080, got %d", cfg.App.Port)
	}
	if !cfg.App.Enabled {
		t.Error("expected app enabled to be true")
	}
	if cfg.DB.Host != "embed-db" {
		t.Errorf("expected db host 'embed-db', got '%s'", cfg.DB.Host)
	}
	if cfg.DB.Port != 5432 {
		t.Errorf("expected db port 5432, got %d", cfg.DB.Port)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", cfg.Timeout)
	}
}

func TestLoadConfig_EmbedFS_WithLocal(t *testing.T) {
	var cfg EmbedConfig
	
	err := LoadConfig(LoaderOptions{
		BaseSource:  EmbedSource(embedFS, "testdata/embed_config.yaml"),
		LocalSource: EmbedSource(embedFS, "testdata/embed_config.local.yaml"),
		Target:      &cfg,
	})
	
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	
	// Base config values
	if cfg.App.Name != "embed-app" {
		t.Errorf("expected app name 'embed-app', got '%s'", cfg.App.Name)
	}
	if cfg.DB.Host != "embed-db" {
		t.Errorf("expected db host 'embed-db', got '%s'", cfg.DB.Host)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", cfg.Timeout)
	}
	
	// Local override values
	if cfg.App.Port != 9090 {
		t.Errorf("expected app port 9090 (from local), got %d", cfg.App.Port)
	}
	if cfg.DB.Port != 3306 {
		t.Errorf("expected db port 3306 (from local), got %d", cfg.DB.Port)
	}
	if cfg.DB.Username != "local-user" {
		t.Errorf("expected db username 'local-user' (from local), got '%s'", cfg.DB.Username)
	}
}

func TestLoadConfig_EmbedFS_WithEnvOverrides(t *testing.T) {
	// Set test environment variables
	t.Setenv("TEST_APP__PORT", "7777")
	t.Setenv("TEST_DB__USERNAME", "env-user")
	t.Setenv("TEST_APP__ENABLED", "false")
	
	var cfg EmbedConfig
	
	err := LoadConfig(LoaderOptions{
		BaseSource: EmbedSource(embedFS, "testdata/embed_config.yaml"),
		EnvPrefix:  "TEST_",
		Delimiter:  "__",
		Target:     &cfg,
	})
	
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	
	// Base config values (unchanged)
	if cfg.App.Name != "embed-app" {
		t.Errorf("expected app name 'embed-app', got '%s'", cfg.App.Name)
	}
	if cfg.DB.Host != "embed-db" {
		t.Errorf("expected db host 'embed-db', got '%s'", cfg.DB.Host)
	}
	
	// Environment overrides
	if cfg.App.Port != 7777 {
		t.Errorf("expected app port 7777 (from env), got %d", cfg.App.Port)
	}
	if cfg.DB.Username != "env-user" {
		t.Errorf("expected db username 'env-user' (from env), got '%s'", cfg.DB.Username)
	}
	if cfg.App.Enabled {
		t.Error("expected app enabled to be false (from env)")
	}
}

func TestLoadConfig_EmbedFS_NonexistentFile(t *testing.T) {
	var cfg EmbedConfig
	
	err := LoadConfig(LoaderOptions{
		BaseSource: EmbedSource(embedFS, "testdata/nonexistent.yaml"),
		Target:     &cfg,
	})
	
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
	
	if !containsString(err.Error(), "load base config") {
		t.Errorf("expected error to contain 'load base config', got: %v", err)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsString(s[1:], substr)))
}