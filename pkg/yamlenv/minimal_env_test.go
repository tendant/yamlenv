package yamlenv

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
	"github.com/stretchr/testify/require"
)

// Minimal test to isolate the exact issue
func TestMinimalEnvironmentProvider(t *testing.T) {
	// Set a simple environment variable
	err := os.Setenv("MINIMAL_TEST", "minimalvalue")
	require.NoError(t, err)
	defer os.Unsetenv("MINIMAL_TEST")

	// Verify it's set
	fmt.Printf("Environment variable MINIMAL_TEST = %s\n", os.Getenv("MINIMAL_TEST"))

	// Create koanf instance
	k := koanf.New(".")

	// Simple mapper that just removes prefix and converts to lowercase
	mapper := func(key string) string {
		fmt.Printf("Mapper input: %s\n", key)
		result := strings.TrimPrefix(key, "MINIMAL_")
		result = strings.ToLower(result)
		fmt.Printf("Mapper output: %s\n", result)
		return result
	}

	// Create environment provider
	envProvider := env.Provider("MINIMAL_", "", mapper)

	// Read from provider
	data, err := envProvider.Read()
	require.NoError(t, err)
	fmt.Printf("Provider data: %v\n", data)

	// Load into koanf
	err = k.Load(envProvider, nil)
	require.NoError(t, err)

	fmt.Printf("Koanf all: %v\n", k.All())
	fmt.Printf("test value: %v\n", k.Get("test"))

	// Test struct
	type MinimalConfig struct {
		Test string `koanf:"test"`
	}

	var cfg MinimalConfig
	err = k.Unmarshal("", &cfg)
	require.NoError(t, err)

	fmt.Printf("Final config: %+v\n", cfg)
	fmt.Printf("Test field: %s\n", cfg.Test)

	// This should work
	if cfg.Test != "minimalvalue" {
		t.Errorf("Expected 'minimalvalue', got '%s'", cfg.Test)
	}
}

// Test with YAML + ENV combination using non-empty delimiter (Fix A)
func TestMinimalYAMLPlusEnv(t *testing.T) {
	// Create simple YAML
	yamlContent := `test: yamlvalue`
	yamlFile := createTempYAML(t, yamlContent)

	// Set environment variable with non-empty delimiter
	err := os.Setenv("COMBO_TEST", "envvalue")
	require.NoError(t, err)
	defer os.Unsetenv("COMBO_TEST")

	fmt.Printf("Environment variable COMBO_TEST = %s\n", os.Getenv("COMBO_TEST"))

	// Test struct
	type ComboConfig struct {
		Test string `koanf:"test"`
	}

	var cfg ComboConfig
	err = LoadConfig(LoaderOptions{
		BaseFile:  yamlFile,
		EnvPrefix: "COMBO_",
		Delimiter: "__", // Fix A: Use non-empty delimiter
		Target:    &cfg,
	})

	require.NoError(t, err)
	fmt.Printf("Final config: %+v\n", cfg)
	fmt.Printf("Test field: %s (should be 'envvalue')\n", cfg.Test)

	// Environment should override YAML
	if cfg.Test != "envvalue" {
		t.Errorf("Expected 'envvalue', got '%s'", cfg.Test)
	}
}
