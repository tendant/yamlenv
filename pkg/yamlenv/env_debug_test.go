package yamlenv

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// Debug test to understand environment variable behavior
func TestDebug_EnvironmentVariables(t *testing.T) {
	// Set environment variables
	err := os.Setenv("DEBUG_APP__NAME", "debugapp")
	require.NoError(t, err)
	defer os.Unsetenv("DEBUG_APP__NAME")

	err = os.Setenv("DEBUG_APP__PORT", "9999")
	require.NoError(t, err)
	defer os.Unsetenv("DEBUG_APP__PORT")

	// Check if they're actually set
	fmt.Printf("DEBUG_APP__NAME = %s\n", os.Getenv("DEBUG_APP__NAME"))
	fmt.Printf("DEBUG_APP__PORT = %s\n", os.Getenv("DEBUG_APP__PORT"))

	// List all environment variables with DEBUG_ prefix
	for _, env := range os.Environ() {
		if len(env) > 6 && env[:6] == "DEBUG_" {
			fmt.Printf("Found env var: %s\n", env)
		}
	}

	baseYAML := `
app:
  name: testapp
  port: 8080
`

	baseFile := createTempYAML(t, baseYAML)

	var cfg TestConfig
	err = LoadConfig(LoaderOptions{
		BaseSource: FileSource(baseFile),
		EnvPrefix:  "DEBUG_",
		Delimiter:  "__",
		Target:     &cfg,
	})

	require.NoError(t, err)
	fmt.Printf("Loaded config: %+v\n", cfg)
	fmt.Printf("App.Name: %s (expected: debugapp)\n", cfg.App.Name)
	fmt.Printf("App.Port: %d (expected: 9999)\n", cfg.App.Port)
}
