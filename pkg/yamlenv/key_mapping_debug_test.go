package yamlenv

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
)

// Debug test to examine exact key paths used by koanf
func TestKeyMappingDebug(t *testing.T) {
	// Create YAML content
	yamlContent := `
app:
  name: yamlvalue
  port: 8080
db:
  host: yamlhost
`

	yamlFile := createTempYAML(t, yamlContent)

	// Set environment variables
	err := os.Setenv("DEBUG_APP__NAME", "envvalue")
	require.NoError(t, err)
	defer os.Unsetenv("DEBUG_APP__NAME")

	err = os.Setenv("DEBUG_DB__HOST", "envhost")
	require.NoError(t, err)
	defer os.Unsetenv("DEBUG_DB__HOST")

	// Create koanf instance
	k := koanf.New(".")

	// Load YAML first
	err = k.Load(file.Provider(yamlFile), yaml.Parser())
	require.NoError(t, err)

	fmt.Printf("After YAML load - All keys: %v\n", k.All())
	fmt.Printf("YAML keys individually:\n")
	for key, value := range k.All() {
		fmt.Printf("  %s = %v\n", key, value)
	}

	// Create environment provider with detailed logging
	mapper := func(key string) string {
		fmt.Printf("Mapper input: '%s'\n", key)

		// Remove prefix
		trimmed := strings.TrimPrefix(key, "DEBUG_")
		fmt.Printf("After prefix removal: '%s'\n", trimmed)

		// Convert to lowercase
		lowered := strings.ToLower(trimmed)
		fmt.Printf("After lowercase: '%s'\n", lowered)

		// Replace delimiter
		result := strings.ReplaceAll(lowered, "__", ".")
		fmt.Printf("Final mapped key: '%s'\n", result)

		return result
	}

	// Load environment variables
	err = k.Load(env.Provider("DEBUG_", "__", mapper), nil)
	require.NoError(t, err)

	fmt.Printf("\nAfter ENV load - All keys: %v\n", k.All())
	fmt.Printf("ENV keys individually:\n")
	for key, value := range k.All() {
		fmt.Printf("  %s = %v\n", key, value)
	}

	// Check specific values
	fmt.Printf("\nSpecific value checks:\n")
	fmt.Printf("k.Get('app.name'): %v\n", k.Get("app.name"))
	fmt.Printf("k.Get('app.port'): %v\n", k.Get("app.port"))
	fmt.Printf("k.Get('db.host'): %v\n", k.Get("db.host"))

	// Test struct unmarshaling
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
	err = k.Unmarshal("", &cfg)
	require.NoError(t, err)

	fmt.Printf("\nFinal unmarshaled config: %+v\n", cfg)
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
