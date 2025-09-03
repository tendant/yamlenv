package yamlenv

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test configuration struct
type TestConfig struct {
	App struct {
		Name  string `koanf:"name"`
		Port  int    `koanf:"port"`
		Debug bool   `koanf:"debug"`
	} `koanf:"app"`
	DB struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
		Name string `koanf:"name"`
	} `koanf:"db"`
	Timeout time.Duration `koanf:"timeout"`
	Version string        `koanf:"version"`
}

// Helper function to create temporary YAML files
func createTempYAML(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "test-*.yaml")
	require.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	t.Cleanup(func() {
		os.Remove(tmpFile.Name())
	})

	return tmpFile.Name()
}

// Helper function to set environment variables with cleanup
func setEnvVar(t *testing.T, key, value string) {
	originalValue := os.Getenv(key)
	err := os.Setenv(key, value)
	require.NoError(t, err)

	t.Cleanup(func() {
		if originalValue == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, originalValue)
		}
	})
}

// Test basic YAML loading without environment variables or local overrides
func TestLoadConfig_BasicYAMLOnly(t *testing.T) {
	baseYAML := `
app:
  name: testapp
  port: 8080
  debug: false
db:
  host: localhost
  port: 5432
  name: testdb
timeout: 30s
version: "1.0.0"
`

	baseFile := createTempYAML(t, baseYAML)

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile: baseFile,
		Target:   &cfg,
	})

	require.NoError(t, err)
	assert.Equal(t, "testapp", cfg.App.Name)
	assert.Equal(t, 8080, cfg.App.Port)
	assert.False(t, cfg.App.Debug)
	assert.Equal(t, "localhost", cfg.DB.Host)
	assert.Equal(t, 5432, cfg.DB.Port)
	assert.Equal(t, "testdb", cfg.DB.Name)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, "1.0.0", cfg.Version)
}

// Test local file overriding base file values
func TestLoadConfig_WithLocalOverride(t *testing.T) {
	baseYAML := `
app:
  name: testapp
  port: 8080
  debug: false
db:
  host: localhost
  port: 5432
version: "1.0.0"
`

	localYAML := `
app:
  port: 3000
  debug: true
db:
  host: dev-db.local
`

	baseFile := createTempYAML(t, baseYAML)
	localFile := createTempYAML(t, localYAML)

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile:  baseFile,
		LocalFile: localFile,
		Target:    &cfg,
	})

	require.NoError(t, err)
	// Base values preserved
	assert.Equal(t, "testapp", cfg.App.Name)
	assert.Equal(t, "1.0.0", cfg.Version)
	assert.Equal(t, 5432, cfg.DB.Port)

	// Local overrides applied
	assert.Equal(t, 3000, cfg.App.Port)
	assert.True(t, cfg.App.Debug)
	assert.Equal(t, "dev-db.local", cfg.DB.Host)
}

// Test missing base file error
func TestLoadConfig_MissingBaseFile(t *testing.T) {
	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile: "nonexistent.yaml",
		Target:   &cfg,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load base yaml")
}

// Test invalid base YAML error
func TestLoadConfig_InvalidBaseYAML(t *testing.T) {
	invalidYAML := `
app:
  name: testapp
  port: invalid_port
  debug: [unclosed array
`

	baseFile := createTempYAML(t, invalidYAML)

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile: baseFile,
		Target:   &cfg,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load base yaml")
}

// Test invalid local YAML error
func TestLoadConfig_InvalidLocalYAML(t *testing.T) {
	baseYAML := `
app:
  name: testapp
  port: 8080
`

	invalidLocalYAML := `
app:
  port: [invalid yaml
`

	baseFile := createTempYAML(t, baseYAML)
	localFile := createTempYAML(t, invalidLocalYAML)

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile:  baseFile,
		LocalFile: localFile,
		Target:    &cfg,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "load local yaml")
}

// Test missing local file (should be ignored)
func TestLoadConfig_MissingLocalFile(t *testing.T) {
	baseYAML := `
app:
  name: testapp
  port: 8080
`

	baseFile := createTempYAML(t, baseYAML)

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile:  baseFile,
		LocalFile: "nonexistent-local.yaml",
		Target:    &cfg,
	})

	// Should succeed - missing local file is optional
	require.NoError(t, err)
	assert.Equal(t, "testapp", cfg.App.Name)
	assert.Equal(t, 8080, cfg.App.Port)
}

// Test nil target error
func TestLoadConfig_NilTarget(t *testing.T) {
	baseYAML := `
app:
  name: testapp
`

	baseFile := createTempYAML(t, baseYAML)

	err := LoadConfig(LoaderOptions{
		BaseFile: baseFile,
		Target:   nil,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal config")
}

// Test non-pointer target error
func TestLoadConfig_NonPointerTarget(t *testing.T) {
	baseYAML := `
app:
  name: testapp
`

	baseFile := createTempYAML(t, baseYAML)

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile: baseFile,
		Target:   cfg, // Not a pointer
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal config")
}

// Test type mismatch error
func TestLoadConfig_TypeMismatch(t *testing.T) {
	// YAML has string where struct expects int
	baseYAML := `
app:
  name: testapp
  port: "not_a_number"
`

	baseFile := createTempYAML(t, baseYAML)

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile: baseFile,
		Target:   &cfg,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unmarshal config")
}

// Test empty YAML file
func TestLoadConfig_EmptyYAML(t *testing.T) {
	baseFile := createTempYAML(t, "")

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile: baseFile,
		Target:   &cfg,
	})

	// Should succeed with zero values
	require.NoError(t, err)
	assert.Equal(t, "", cfg.App.Name)
	assert.Equal(t, 0, cfg.App.Port)
	assert.False(t, cfg.App.Debug)
}

// Test YAML with only comments
func TestLoadConfig_CommentsOnlyYAML(t *testing.T) {
	commentsYAML := `
# This is a comment
# Another comment
`

	baseFile := createTempYAML(t, commentsYAML)

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile: baseFile,
		Target:   &cfg,
	})

	// Should succeed with zero values
	require.NoError(t, err)
	assert.Equal(t, "", cfg.App.Name)
	assert.Equal(t, 0, cfg.App.Port)
}

// Integration test using the same structure as the demo
func TestLoadConfig_DemoCompatibility(t *testing.T) {
	// Use the same struct as in cmd/demo/main.go
	type DemoConfig struct {
		App struct {
			Name string `koanf:"name"`
			Port int    `koanf:"port"`
		} `koanf:"app"`
		DB struct {
			Host string `koanf:"host"`
			Port int    `koanf:"port"`
		} `koanf:"db"`
		Timeout time.Duration `koanf:"timeout"`
	}

	// Use similar YAML structure as demo
	demoYAML := `
app:
  name: fi-stock
  port: 8080
db:
  host: localhost
  port: 5432
timeout: 30s
`

	baseFile := createTempYAML(t, demoYAML)

	var cfg DemoConfig
	err := LoadConfig(LoaderOptions{
		BaseFile:  baseFile,
		LocalFile: "nonexistent.local.yaml", // Should be ignored
		EnvPrefix: "SAMPLE_",
		Delimiter: "__",
		Target:    &cfg,
	})

	require.NoError(t, err)
	assert.Equal(t, "fi-stock", cfg.App.Name)
	assert.Equal(t, 8080, cfg.App.Port)
	assert.Equal(t, "localhost", cfg.DB.Host)
	assert.Equal(t, 5432, cfg.DB.Port)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
}

// Benchmark test
func BenchmarkLoadConfig(b *testing.B) {
	baseYAML := `
app:
  name: benchapp
  port: 8080
  debug: false
db:
  host: localhost
  port: 5432
  name: benchdb
timeout: 30s
version: "1.0.0"
`

	tmpFile, err := os.CreateTemp("", "bench-*.yaml")
	require.NoError(b, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(baseYAML)
	require.NoError(b, err)
	tmpFile.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var cfg TestConfig
		err := LoadConfig(LoaderOptions{
			BaseFile: tmpFile.Name(),
			Target:   &cfg,
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}
