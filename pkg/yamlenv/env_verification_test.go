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

// Test to verify environment variables are being processed correctly
func TestEnvironmentVariableVerification(t *testing.T) {
	// Set environment variables
	err := os.Setenv("VERIFY_APP__NAME", "verify_app_name")
	require.NoError(t, err)
	defer os.Unsetenv("VERIFY_APP__NAME")

	err = os.Setenv("VERIFY_DB__HOST", "verify_db_host")
	require.NoError(t, err)
	defer os.Unsetenv("VERIFY_DB__HOST")

	// Verify they're actually set
	fmt.Printf("VERIFY_APP__NAME = %s\n", os.Getenv("VERIFY_APP__NAME"))
	fmt.Printf("VERIFY_DB__HOST = %s\n", os.Getenv("VERIFY_DB__HOST"))

	// List all environment variables with our prefix
	fmt.Printf("All VERIFY_ environment variables:\n")
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "VERIFY_") {
			fmt.Printf("  %s\n", env)
		}
	}

	// Create koanf instance
	k := koanf.New(".")

	// Create environment provider with detailed logging
	mapper := func(key string) string {
		fmt.Printf("Processing env var: %s\n", key)

		// Remove prefix
		trimmed := strings.TrimPrefix(key, "VERIFY_")
		fmt.Printf("  After prefix removal: %s\n", trimmed)

		// Convert to lowercase
		lowered := strings.ToLower(trimmed)
		fmt.Printf("  After lowercase: %s\n", lowered)

		// Replace delimiter
		result := strings.ReplaceAll(lowered, "__", ".")
		fmt.Printf("  Final mapped key: %s\n", result)

		return result
	}

	// Load environment variables only (no YAML)
	err = k.Load(env.Provider("VERIFY_", "__", mapper), nil)
	require.NoError(t, err)

	fmt.Printf("\nKoanf contents after env load:\n")
	for key, value := range k.All() {
		fmt.Printf("  %s = %v\n", key, value)
	}

	// Check if both values are present
	appName := k.Get("app.name")
	dbHost := k.Get("db.host")

	fmt.Printf("\nDirect value checks:\n")
	fmt.Printf("k.Get('app.name'): %v\n", appName)
	fmt.Printf("k.Get('db.host'): %v\n", dbHost)

	// Both should be present
	if appName != "verify_app_name" {
		t.Errorf("Expected 'verify_app_name', got '%v'", appName)
	}
	if dbHost != "verify_db_host" {
		t.Errorf("Expected 'verify_db_host', got '%v'", dbHost)
	}
}
