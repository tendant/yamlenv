package yamlenv

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Complete test suite with working environment variable tests
// These tests work correctly with the fixed library code

func TestLoadConfig_EnvironmentVariables_Working(t *testing.T) {
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

	baseFile := createTempYAML(t, baseYAML)

	// Set environment variables with unique prefix to avoid conflicts
	setEnvVar(t, "WORKING_APP__NAME", "envapp")
	setEnvVar(t, "WORKING_APP__PORT", "9000")
	setEnvVar(t, "WORKING_DB__HOST", "env-db.example.com")
	setEnvVar(t, "WORKING_VERSION", "2.0.0")

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseSource: FileSource(baseFile),
		EnvPrefix: "WORKING_",
		Delimiter: "__",
		Target:    &cfg,
	})

	require.NoError(t, err)

	// Environment overrides should work
	assert.Equal(t, "envapp", cfg.App.Name)
	assert.Equal(t, 9000, cfg.App.Port)
	assert.Equal(t, "env-db.example.com", cfg.DB.Host)
	assert.Equal(t, "2.0.0", cfg.Version)

	// Base values preserved where no env override
	assert.False(t, cfg.App.Debug)
	assert.Equal(t, 5432, cfg.DB.Port)
}

func TestLoadConfig_PriorityOrder_Working(t *testing.T) {
	baseYAML := `
app:
  name: baseapp
  port: 8080
  debug: false
version: "1.0.0"
`

	localYAML := `
app:
  name: localapp
  port: 3000
version: "1.1.0"
`

	baseFile := createTempYAML(t, baseYAML)
	localFile := createTempYAML(t, localYAML)

	// Environment should override both
	setEnvVar(t, "PRIORITY2_APP__NAME", "envapp")
	setEnvVar(t, "PRIORITY2_VERSION", "2.0.0")

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseSource:  FileSource(baseFile),
		LocalSource: FileSource(localFile),
		EnvPrefix:   "PRIORITY2_",
		Delimiter:   "__",
		Target:      &cfg,
	})

	require.NoError(t, err)

	// Environment has highest priority
	assert.Equal(t, "envapp", cfg.App.Name)
	assert.Equal(t, "2.0.0", cfg.Version)

	// Local overrides base where no env override
	assert.Equal(t, 3000, cfg.App.Port)

	// Base value where no local or env override
	assert.False(t, cfg.App.Debug)
}

func TestLoadConfig_DifferentDelimiters_Working(t *testing.T) {
	baseYAML := `
app:
  name: testapp
  port: 8080
db:
  host: localhost
`

	baseFile := createTempYAML(t, baseYAML)

	// Test with different delimiter
	setEnvVar(t, "DELIM2_APP_NAME", "delimapp")
	setEnvVar(t, "DELIM2_DB_HOST", "delim-db.local")

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseSource: FileSource(baseFile),
		EnvPrefix: "DELIM2_",
		Delimiter: "_",
		Target:    &cfg,
	})

	require.NoError(t, err)
	assert.Equal(t, "delimapp", cfg.App.Name)
	assert.Equal(t, "delim-db.local", cfg.DB.Host)
	assert.Equal(t, 8080, cfg.App.Port) // Not overridden
}

func TestLoadConfig_NoEnvPrefix_Working(t *testing.T) {
	baseYAML := `
app:
  name: testapp
  port: 8080
`

	baseFile := createTempYAML(t, baseYAML)

	// Set env vars without prefix - use unique names
	setEnvVar(t, "APP__NAME", "noprefix")
	setEnvVar(t, "APP__PORT", "9000")

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseSource: FileSource(baseFile),
		EnvPrefix: "", // No prefix
		Delimiter: "__",
		Target:    &cfg,
	})

	require.NoError(t, err)
	assert.Equal(t, "noprefix", cfg.App.Name)
	assert.Equal(t, 9000, cfg.App.Port)
}

func TestLoadConfig_ComplexNestedStructure_Working(t *testing.T) {
	complexYAML := `
app:
  name: complex-app
  port: 8080
  debug: true
db:
  host: complex-db.local
  port: 5432
  name: complex_db
timeout: 45s
version: "2.1.0"
`

	baseFile := createTempYAML(t, complexYAML)

	// Test with environment overrides for nested values
	setEnvVar(t, "COMPLEX2_APP__DEBUG", "false")
	setEnvVar(t, "COMPLEX2_DB__PORT", "3306")
	setEnvVar(t, "COMPLEX2_TIMEOUT", "60s")

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseSource: FileSource(baseFile),
		EnvPrefix: "COMPLEX2_",
		Delimiter: "__",
		Target:    &cfg,
	})

	require.NoError(t, err)
	assert.Equal(t, "complex-app", cfg.App.Name)
	assert.Equal(t, 8080, cfg.App.Port)
	assert.False(t, cfg.App.Debug) // Overridden by env
	assert.Equal(t, "complex-db.local", cfg.DB.Host)
	assert.Equal(t, 3306, cfg.DB.Port) // Overridden by env
	assert.Equal(t, "complex_db", cfg.DB.Name)
	assert.Equal(t, 60*time.Second, cfg.Timeout) // Overridden by env
	assert.Equal(t, "2.1.0", cfg.Version)
}

// Test that demonstrates the library works correctly
func TestLoadConfig_FullIntegration(t *testing.T) {
	baseYAML := `
server:
  host: localhost
  port: 8080
  ssl: false
database:
  driver: postgres
  host: localhost
  port: 5432
  name: myapp
  ssl: false
logging:
  level: info
  format: json
cache:
  ttl: 300s
  size: 1000
`

	localYAML := `
server:
  port: 3000
  ssl: true
database:
  host: dev-db
  name: myapp_dev
logging:
  level: debug
`

	baseFile := createTempYAML(t, baseYAML)
	localFile := createTempYAML(t, localYAML)

	// Set some environment overrides
	setEnvVar(t, "MYAPP_SERVER__HOST", "prod-server")
	setEnvVar(t, "MYAPP_DATABASE__PORT", "5433")
	setEnvVar(t, "MYAPP_LOGGING__FORMAT", "text")
	setEnvVar(t, "MYAPP_CACHE__TTL", "600s")

	type FullConfig struct {
		Server struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
			SSL  bool   `yaml:"ssl"`
		} `yaml:"server"`
		Database struct {
			Driver string `yaml:"driver"`
			Host   string `yaml:"host"`
			Port   int    `yaml:"port"`
			Name   string `yaml:"name"`
			SSL    bool   `yaml:"ssl"`
		} `yaml:"database"`
		Logging struct {
			Level  string `yaml:"level"`
			Format string `yaml:"format"`
		} `yaml:"logging"`
		Cache struct {
			TTL  time.Duration `yaml:"ttl"`
			Size int           `yaml:"size"`
		} `yaml:"cache"`
	}

	var cfg FullConfig
	err := LoadConfig(LoaderOptions{
		BaseSource:  FileSource(baseFile),
		LocalSource: FileSource(localFile),
		EnvPrefix:   "MYAPP_",
		Delimiter:   "__",
		Target:      &cfg,
	})

	require.NoError(t, err)

	// Verify priority: ENV > Local > Base
	assert.Equal(t, "prod-server", cfg.Server.Host) // ENV override
	assert.Equal(t, 3000, cfg.Server.Port)          // Local override
	assert.True(t, cfg.Server.SSL)                  // Local override

	assert.Equal(t, "postgres", cfg.Database.Driver) // Base value
	assert.Equal(t, "dev-db", cfg.Database.Host)     // Local override
	assert.Equal(t, 5433, cfg.Database.Port)         // ENV override
	assert.Equal(t, "myapp_dev", cfg.Database.Name)  // Local override
	assert.False(t, cfg.Database.SSL)                // Base value

	assert.Equal(t, "debug", cfg.Logging.Level) // Local override
	assert.Equal(t, "text", cfg.Logging.Format) // ENV override

	assert.Equal(t, 600*time.Second, cfg.Cache.TTL) // ENV override
	assert.Equal(t, 1000, cfg.Cache.Size)           // Base value
}
