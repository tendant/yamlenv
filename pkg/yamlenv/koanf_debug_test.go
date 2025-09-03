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

// Test koanf environment provider directly
func TestKoanfEnvProvider(t *testing.T) {
	// Set environment variables
	err := os.Setenv("KOANF_APP__NAME", "koanftest")
	require.NoError(t, err)
	defer os.Unsetenv("KOANF_APP__NAME")

	err = os.Setenv("KOANF_APP__PORT", "8888")
	require.NoError(t, err)
	defer os.Unsetenv("KOANF_APP__PORT")

	// Create koanf instance
	k := koanf.New(".")

	// Load base YAML first
	baseYAML := `
app:
  name: baseapp
  port: 3000
`
	baseFile := createTempYAML(t, baseYAML)
	err = k.Load(file.Provider(baseFile), yaml.Parser())
	require.NoError(t, err)

	fmt.Printf("After YAML load: %v\n", k.All())

	// Test environment provider
	mapper := func(key string) string {
		fmt.Printf("Mapper called with key: %s\n", key)
		key = strings.TrimPrefix(key, "KOANF_")
		key = strings.ToLower(key)
		result := strings.ReplaceAll(key, "__", ".")
		fmt.Printf("Mapped to: %s\n", result)
		return result
	}

	err = k.Load(env.Provider("KOANF_", "__", mapper), nil)
	require.NoError(t, err)

	fmt.Printf("After env load: %v\n", k.All())

	// Check specific values
	fmt.Printf("app.name: %v\n", k.Get("app.name"))
	fmt.Printf("app.port: %v\n", k.Get("app.port"))

	var cfg TestConfig
	err = k.Unmarshal("", &cfg)
	require.NoError(t, err)

	fmt.Printf("Final config: %+v\n", cfg)
}
