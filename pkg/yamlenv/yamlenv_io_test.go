package yamlenv

import (
	"embed"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

//go:embed testdata/embed_config.yaml testdata/embed_config.local.yaml
var testEmbedFS embed.FS

type IOTestConfig struct {
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

func TestLoadConfig_FileSource(t *testing.T) {
	var cfg IOTestConfig
	
	err := LoadConfig(LoaderOptions{
		BaseSource: FileSource("testdata/embed_config.yaml"),
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
	if cfg.DB.Host != "embed-db" {
		t.Errorf("expected db host 'embed-db', got '%s'", cfg.DB.Host)
	}
}

func TestLoadConfig_EmbedSource(t *testing.T) {
	var cfg IOTestConfig
	
	err := LoadConfig(LoaderOptions{
		BaseSource: EmbedSource(testEmbedFS, "testdata/embed_config.yaml"),
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
	if cfg.DB.Host != "embed-db" {
		t.Errorf("expected db host 'embed-db', got '%s'", cfg.DB.Host)
	}
}

func TestLoadConfig_ReaderSource(t *testing.T) {
	yamlContent := `
app:
  name: reader-app
  port: 7070
  enabled: false
db:
  host: reader-db
  port: 3333
  username: reader-user
timeout: 45s
`
	
	var cfg IOTestConfig
	
	err := LoadConfig(LoaderOptions{
		BaseSource: ReaderSource(strings.NewReader(yamlContent)),
		Target:     &cfg,
	})
	
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	
	if cfg.App.Name != "reader-app" {
		t.Errorf("expected app name 'reader-app', got '%s'", cfg.App.Name)
	}
	if cfg.App.Port != 7070 {
		t.Errorf("expected app port 7070, got %d", cfg.App.Port)
	}
	if cfg.DB.Host != "reader-db" {
		t.Errorf("expected db host 'reader-db', got '%s'", cfg.DB.Host)
	}
	if cfg.Timeout != 45*time.Second {
		t.Errorf("expected timeout 45s, got %v", cfg.Timeout)
	}
}

func TestLoadConfig_WithLocalSource(t *testing.T) {
	var cfg IOTestConfig
	
	err := LoadConfig(LoaderOptions{
		BaseSource:  EmbedSource(testEmbedFS, "testdata/embed_config.yaml"),
		LocalSource: EmbedSource(testEmbedFS, "testdata/embed_config.local.yaml"),
		Target:      &cfg,
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

func TestLoadConfig_WithEnvOverrides(t *testing.T) {
	// Set test environment variables
	t.Setenv("IOTEST_APP__PORT", "6666")
	t.Setenv("IOTEST_DB__USERNAME", "env-user")
	t.Setenv("IOTEST_APP__ENABLED", "false")
	
	var cfg IOTestConfig
	
	err := LoadConfig(LoaderOptions{
		BaseSource: EmbedSource(testEmbedFS, "testdata/embed_config.yaml"),
		EnvPrefix:  "IOTEST_",
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
	if cfg.App.Port != 6666 {
		t.Errorf("expected app port 6666 (from env), got %d", cfg.App.Port)
	}
	if cfg.DB.Username != "env-user" {
		t.Errorf("expected db username 'env-user' (from env), got '%s'", cfg.DB.Username)
	}
	if cfg.App.Enabled {
		t.Error("expected app enabled to be false (from env)")
	}
}

func TestLoadConfig_MixedSources(t *testing.T) {
	baseYAML := `
app:
  name: mixed-base
  port: 1111
db:
  host: base-db
  port: 2222
timeout: 10s
`
	
	localYAML := `
app:
  port: 3333
db:
  username: local-mixed
`
	
	// Set test environment variables
	t.Setenv("MIXED_DB__HOST", "env-db")
	t.Setenv("MIXED_TIMEOUT", "60s")
	
	var cfg IOTestConfig
	
	err := LoadConfig(LoaderOptions{
		BaseSource:  ReaderSource(strings.NewReader(baseYAML)),
		LocalSource: ReaderSource(strings.NewReader(localYAML)),
		EnvPrefix:   "MIXED_",
		Delimiter:   "__",
		Target:      &cfg,
	})
	
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	
	// From base (unchanged by local or env)
	if cfg.App.Name != "mixed-base" {
		t.Errorf("expected app name 'mixed-base', got '%s'", cfg.App.Name)
	}
	
	// From local override
	if cfg.App.Port != 3333 {
		t.Errorf("expected app port 3333 (from local), got %d", cfg.App.Port)
	}
	if cfg.DB.Username != "local-mixed" {
		t.Errorf("expected db username 'local-mixed' (from local), got '%s'", cfg.DB.Username)
	}
	
	// From environment override
	if cfg.DB.Host != "env-db" {
		t.Errorf("expected db host 'env-db' (from env), got '%s'", cfg.DB.Host)
	}
	if cfg.Timeout != 60*time.Second {
		t.Errorf("expected timeout 60s (from env), got %v", cfg.Timeout)
	}
	
	// Base value with no overrides
	if cfg.DB.Port != 2222 {
		t.Errorf("expected db port 2222 (from base), got %d", cfg.DB.Port)
	}
}

func TestLoadConfig_NilBaseSource(t *testing.T) {
	var cfg IOTestConfig
	
	err := LoadConfig(LoaderOptions{
		BaseSource: nil,
		Target:     &cfg,
	})
	
	if err == nil {
		t.Error("expected error for nil BaseSource, got nil")
	}
	
	if !strings.Contains(err.Error(), "BaseSource cannot be nil") {
		t.Errorf("expected error to contain 'BaseSource cannot be nil', got: %v", err)
	}
}

func TestLoadConfig_SourceError(t *testing.T) {
	var cfg IOTestConfig
	
	// Create a source that will fail
	failingSource := func() (io.ReadCloser, error) {
		return nil, fmt.Errorf("source failure")
	}
	
	err := LoadConfig(LoaderOptions{
		BaseSource: failingSource,
		Target:     &cfg,
	})
	
	if err == nil {
		t.Error("expected error for failing source, got nil")
	}
	
	if !strings.Contains(err.Error(), "load base config") {
		t.Errorf("expected error to contain 'load base config', got: %v", err)
	}
}