package yamlenv

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Fixed tests that properly handle prefix/delimiter/casing issues
// Based on the identified root causes of override failures

func TestLoadConfig_EnvironmentOverrides_Fixed(t *testing.T) {
	baseYAML := `
app:
  name: baseapp
  port: 8080
  debug: false
db:
  host: localhost
  port: 5432
version: "1.0.0"
`

	baseFile := createTempYAML(t, baseYAML)

	// CRITICAL: Ensure prefix includes trailing underscore and matches exactly
	// Environment variables with proper prefix and delimiter
	setEnvVar(t, "MYAPP_APP__NAME", "envapp")    // MYAPP_ prefix, __ delimiter -> app.name
	setEnvVar(t, "MYAPP_APP__PORT", "9000")      // MYAPP_ prefix, __ delimiter -> app.port
	setEnvVar(t, "MYAPP_APP__DEBUG", "true")     // MYAPP_ prefix, __ delimiter -> app.debug
	setEnvVar(t, "MYAPP_DB__HOST", "env-db.com") // MYAPP_ prefix, __ delimiter -> db.host
	setEnvVar(t, "MYAPP_VERSION", "2.0.0")       // MYAPP_ prefix, no delimiter -> version

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile:  baseFile,
		EnvPrefix: "MYAPP_", // Must match the prefix in env vars exactly
		Delimiter: "__",     // Must match the delimiter in env vars exactly
		Target:    &cfg,
	})

	require.NoError(t, err)

	// These should now work correctly
	assert.Equal(t, "envapp", cfg.App.Name)    // Environment override
	assert.Equal(t, 9000, cfg.App.Port)        // Environment override
	assert.True(t, cfg.App.Debug)              // Environment override
	assert.Equal(t, "env-db.com", cfg.DB.Host) // Environment override
	assert.Equal(t, "2.0.0", cfg.Version)      // Environment override

	// Base value preserved where no env override
	assert.Equal(t, 5432, cfg.DB.Port) // No env var set
}

func TestLoadConfig_PrefixMismatch_Demonstration(t *testing.T) {
	baseYAML := `
app:
  name: baseapp
`

	baseFile := createTempYAML(t, baseYAML)

	// Set environment variable with WRONG prefix (missing trailing underscore)
	setEnvVar(t, "WRONGAPP__NAME", "shouldnotwork")

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile:  baseFile,
		EnvPrefix: "WRONG_", // Different from env var prefix
		Delimiter: "__",
		Target:    &cfg,
	})

	require.NoError(t, err)

	// Should NOT be overridden due to prefix mismatch
	assert.Equal(t, "baseapp", cfg.App.Name)
}

func TestLoadConfig_DelimiterMismatch_Demonstration(t *testing.T) {
	baseYAML := `
app:
  name: baseapp
`

	baseFile := createTempYAML(t, baseYAML)

	// Set environment variable with WRONG delimiter
	setEnvVar(t, "DELIM_APP_NAME", "shouldnotwork") // Using single underscore

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile:  baseFile,
		EnvPrefix: "DELIM_",
		Delimiter: "__", // Expecting double underscore, but env var uses single
		Target:    &cfg,
	})

	require.NoError(t, err)

	// Should NOT be overridden due to delimiter mismatch
	assert.Equal(t, "baseapp", cfg.App.Name)
}

func TestLoadConfig_CorrectDelimiter_Fixed(t *testing.T) {
	baseYAML := `
app:
  name: baseapp
`

	baseFile := createTempYAML(t, baseYAML)

	// Set environment variable with CORRECT delimiter
	setEnvVar(t, "DELIM_APP_NAME", "shouldwork") // Using single underscore

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile:  baseFile,
		EnvPrefix: "DELIM_",
		Delimiter: "_", // Matching the single underscore in env var
		Target:    &cfg,
	})

	require.NoError(t, err)

	// Should be overridden with correct delimiter
	assert.Equal(t, "shouldwork", cfg.App.Name)
}

func TestLoadConfig_CaseSensitivity_Fixed(t *testing.T) {
	// YAML with lowercase keys (standard)
	baseYAML := `
app:
  name: baseapp
database:
  hostname: localhost
`

	baseFile := createTempYAML(t, baseYAML)

	// Environment variables in UPPERCASE (standard)
	setEnvVar(t, "CASE_APP__NAME", "UPPERCASE_VALUE")
	setEnvVar(t, "CASE_DATABASE__HOSTNAME", "UPPERCASE_HOST")

	type CaseConfig struct {
		App struct {
			Name string `yaml:"name"`
		} `yaml:"app"`
		Database struct {
			Hostname string `yaml:"hostname"`
		} `yaml:"database"`
	}

	var cfg CaseConfig
	err := LoadConfig(LoaderOptions{
		BaseFile:  baseFile,
		EnvPrefix: "CASE_",
		Delimiter: "__",
		Target:    &cfg,
	})

	require.NoError(t, err)

	// Should work because our mapper converts to lowercase
	assert.Equal(t, "UPPERCASE_VALUE", cfg.App.Name)
	assert.Equal(t, "UPPERCASE_HOST", cfg.Database.Hostname)
}

func TestLoadConfig_ComplexTypes_Fixed(t *testing.T) {
	baseYAML := `
server:
  timeout: 30s
  enabled: false
  port: 8080
`

	baseFile := createTempYAML(t, baseYAML)

	// Test different data types
	setEnvVar(t, "TYPES_SERVER__TIMEOUT", "60s")  // duration
	setEnvVar(t, "TYPES_SERVER__ENABLED", "true") // boolean
	setEnvVar(t, "TYPES_SERVER__PORT", "9000")    // integer

	type TypesConfig struct {
		Server struct {
			Timeout time.Duration `yaml:"timeout"`
			Enabled bool          `yaml:"enabled"`
			Port    int           `yaml:"port"`
		} `yaml:"server"`
	}

	var cfg TypesConfig
	err := LoadConfig(LoaderOptions{
		BaseFile:  baseFile,
		EnvPrefix: "TYPES_",
		Delimiter: "__",
		Target:    &cfg,
	})

	require.NoError(t, err)

	// Verify type conversions work correctly
	assert.Equal(t, 60*time.Second, cfg.Server.Timeout)
	assert.True(t, cfg.Server.Enabled)
	assert.Equal(t, 9000, cfg.Server.Port)
}

func TestLoadConfig_DebugMapping(t *testing.T) {
	baseYAML := `
app:
  name: baseapp
`

	baseFile := createTempYAML(t, baseYAML)

	// Set environment variable
	setEnvVar(t, "DEBUG_APP__NAME", "debugvalue")

	// Test with debug output to see what's happening
	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile:  baseFile,
		EnvPrefix: "DEBUG_",
		Delimiter: "__",
		Target:    &cfg,
	})

	require.NoError(t, err)

	// This should work now
	assert.Equal(t, "debugvalue", cfg.App.Name)

	// Print for manual verification
	t.Logf("Final config: %+v", cfg)
	t.Logf("App.Name: %s", cfg.App.Name)
}

// Test the exact pattern from the demo
func TestLoadConfig_DemoPattern_Fixed(t *testing.T) {
	// Use the exact same YAML structure as the demo
	demoYAML := `
app:
  name: fi-stock
  port: 8080
db:
  host: localhost
  port: 5432
`

	baseFile := createTempYAML(t, demoYAML)

	// Use the exact same pattern as the demo
	setEnvVar(t, "SAMPLE_APP__NAME", "demo-override")
	setEnvVar(t, "SAMPLE_APP__PORT", "7777")
	setEnvVar(t, "SAMPLE_DB__HOST", "demo-db")

	// Use the same struct as the demo
	type DemoConfig struct {
		App struct {
			Name string `yaml:"name"`
			Port int    `yaml:"port"`
		} `yaml:"app"`
		DB struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		} `yaml:"db"`
		Timeout time.Duration `yaml:"timeout"`
	}

	var cfg DemoConfig
	err := LoadConfig(LoaderOptions{
		BaseFile:  baseFile,
		LocalFile: "", // No local file
		EnvPrefix: "SAMPLE_",
		Delimiter: "__",
		Target:    &cfg,
	})

	require.NoError(t, err)

	// These should now work
	assert.Equal(t, "demo-override", cfg.App.Name)
	assert.Equal(t, 7777, cfg.App.Port)
	assert.Equal(t, "demo-db", cfg.DB.Host)
	assert.Equal(t, 5432, cfg.DB.Port) // Not overridden
}
