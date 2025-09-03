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

// Test environment provider in isolation
func TestEnvProviderIsolated(t *testing.T) {
	// Set environment variables
	err := os.Setenv("ENVTEST_APP__NAME", "envvalue")
	require.NoError(t, err)
	defer os.Unsetenv("ENVTEST_APP__NAME")

	err = os.Setenv("ENVTEST_APP__PORT", "9999")
	require.NoError(t, err)
	defer os.Unsetenv("ENVTEST_APP__PORT")

	// Create koanf instance
	k := koanf.New(".")

	// Test environment provider directly
	mapper := func(key string) string {
		fmt.Printf("Mapper input: %s\n", key)
		key = strings.TrimPrefix(key, "ENVTEST_")
		key = strings.ToLower(key)
		result := strings.ReplaceAll(key, "__", ".")
		fmt.Printf("Mapper output: %s\n", result)
		return result
	}

	// Create the environment provider
	envProvider := env.Provider("ENVTEST_", "__", mapper)

	// Read from the provider
	data, err := envProvider.Read()
	require.NoError(t, err)

	fmt.Printf("Environment provider data: %v\n", data)

	// Load into koanf
	err = k.Load(envProvider, nil)
	require.NoError(t, err)

	fmt.Printf("Koanf all data: %v\n", k.All())
	fmt.Printf("app.name: %v\n", k.Get("app.name"))
	fmt.Printf("app.port: %v\n", k.Get("app.port"))
}
