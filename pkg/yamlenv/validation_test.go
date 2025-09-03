package yamlenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test validation for empty delimiter when EnvPrefix is provided
func TestLoadConfig_EmptyDelimiterValidation(t *testing.T) {
	baseYAML := `
app:
  name: testapp
`

	baseFile := createTempYAML(t, baseYAML)

	type TestConfig struct {
		App struct {
			Name string `yaml:"name"`
		} `yaml:"app"`
	}

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile:  baseFile,
		EnvPrefix: "TEST_",
		Delimiter: "", // Empty delimiter should cause validation error
		Target:    &cfg,
	})

	// Should return validation error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delimiter cannot be empty when EnvPrefix is provided")
	assert.Contains(t, err.Error(), "use a non-empty delimiter like '__'")
}

// Test that validation allows empty delimiter when no EnvPrefix is provided
func TestLoadConfig_EmptyDelimiterAllowedWithoutEnvPrefix(t *testing.T) {
	baseYAML := `
app:
  name: testapp
`

	baseFile := createTempYAML(t, baseYAML)

	type TestConfig struct {
		App struct {
			Name string `yaml:"name"`
		} `yaml:"app"`
	}

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile:  baseFile,
		EnvPrefix: "", // No env prefix
		Delimiter: "", // Empty delimiter should be allowed
		Target:    &cfg,
	})

	// Should succeed
	require.NoError(t, err)
	assert.Equal(t, "testapp", cfg.App.Name)
}

// Test that validation allows non-empty delimiter with EnvPrefix
func TestLoadConfig_NonEmptyDelimiterWithEnvPrefix(t *testing.T) {
	baseYAML := `
app:
  name: testapp
`

	baseFile := createTempYAML(t, baseYAML)

	type TestConfig struct {
		App struct {
			Name string `yaml:"name"`
		} `yaml:"app"`
	}

	var cfg TestConfig
	err := LoadConfig(LoaderOptions{
		BaseFile:  baseFile,
		EnvPrefix: "TEST_",
		Delimiter: "__", // Non-empty delimiter should be allowed
		Target:    &cfg,
	})

	// Should succeed
	require.NoError(t, err)
	assert.Equal(t, "testapp", cfg.App.Name)
}
