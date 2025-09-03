package yamlenv

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test LoadConfig with the same structure as the debug test
func TestLoadConfigDebug(t *testing.T) {
	// Create YAML content - same as debug test
	yamlContent := `
app:
  name: yamlvalue
  port: 8080
db:
  host: yamlhost
`

	yamlFile := createTempYAML(t, yamlContent)

	// Set environment variables - same as debug test
	err := os.Setenv("LOADCFG_APP__NAME", "envvalue")
	require.NoError(t, err)
	defer os.Unsetenv("LOADCFG_APP__NAME")

	err = os.Setenv("LOADCFG_DB__HOST", "envhost")
	require.NoError(t, err)
	defer os.Unsetenv("LOADCFG_DB__HOST")

	// Test struct - same as debug test
	type DebugConfig struct {
		App struct {
			Name string `koanf:"name"`
			Port int    `koanf:"port"`
		} `koanf:"app"`
		DB struct {
			Host string `koanf:"host"`
		} `koanf:"db"`
	}

	var cfg DebugConfig
	err = LoadConfig(LoaderOptions{
		BaseFile:  yamlFile,
		EnvPrefix: "LOADCFG_",
		Delimiter: "__",
		Target:    &cfg,
	})

	require.NoError(t, err)

	fmt.Printf("LoadConfig result: %+v\n", cfg)
	fmt.Printf("cfg.App.Name: %s (should be 'envvalue')\n", cfg.App.Name)
	fmt.Printf("cfg.DB.Host: %s (should be 'envhost')\n", cfg.DB.Host)

	// The test - environment should override YAML
	if cfg.App.Name != "envvalue" {
		t.Errorf("Expected 'envvalue', got '%s'", cfg.App.Name)
	}
	if cfg.DB.Host != "envhost" {
		t.Errorf("Expected 'envhost', got '%s'", cfg.DB.Host)
	}
}
