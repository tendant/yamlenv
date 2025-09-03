package yamlenv

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type LoaderOptions struct {
	BaseFile       string // required
	LocalFile      string // optional
	EnvPrefix      string // e.g. "WORKING_"
	Delimiter      string // nesting delimiter in env, e.g. "__"; "" = no nesting
	Target         any    // &cfg
	NormalizeDash  bool   // if true, convert "_" in ENV path to "-" in YAML keys (for kebab-case YAML like "app-name")
	ForceLowerYAML bool   // if true, normalize YAML keys to lowercase to match ENV mapping
	DebugKeys      bool   // if true, print final keys for debugging
}

// loadYAMLFile loads a YAML file into the target struct
func loadYAMLFile(filename string, target any) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read file %s: %w", filename, err)
	}

	return yaml.Unmarshal(data, target)
}

// getStructPath builds a dot-separated path for a struct field
func getStructPath(field reflect.StructField, yamlTag string) string {
	if yamlTag != "" && yamlTag != "-" {
		return yamlTag
	}
	return strings.ToLower(field.Name)
}

// findEnvValue finds environment variables matching a struct path
func findEnvValue(envPrefix, delimiter string, path string, normalizeDash bool) (string, bool) {
	// Convert path back to env var format: app.name -> APP__NAME
	envPath := strings.ToUpper(path)
	if delimiter != "" {
		envPath = strings.ReplaceAll(envPath, ".", delimiter)
	}
	if normalizeDash {
		// Convert dashes back to underscores for env lookup
		envPath = strings.ReplaceAll(envPath, "-", "_")
	}

	envKey := envPrefix + envPath
	value, exists := os.LookupEnv(envKey)
	return value, exists
}

// setFieldValue sets a struct field value from a string
func setFieldValue(field reflect.Value, value string) error {
	if !field.CanSet() {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("parse duration %q: %w", value, err)
			}
			field.Set(reflect.ValueOf(duration))
		} else {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("parse int %q: %w", value, err)
			}
			field.SetInt(intVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("parse uint %q: %w", value, err)
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("parse float %q: %w", value, err)
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("parse bool %q: %w", value, err)
		}
		field.SetBool(boolVal)
	default:
		return fmt.Errorf("unsupported field type %v", field.Type())
	}
	return nil
}

// applyEnvOverrides recursively applies environment variable overrides
func applyEnvOverrides(val reflect.Value, envPrefix, delimiter string, normalizeDash bool, path string, debugKeys bool) error {
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		// Skip unexported fields
		if !fieldType.IsExported() {
			continue
		}

		// Get yaml or koanf tag or use field name  
		yamlTag := fieldType.Tag.Get("yaml")
		if yamlTag == "" {
			// Fallback to koanf tag for compatibility
			yamlTag = fieldType.Tag.Get("koanf")
		}
		if yamlTag == "-" {
			continue
		}
		// Remove options like ",omitempty"
		if idx := strings.Index(yamlTag, ","); idx >= 0 {
			yamlTag = yamlTag[:idx]
		}
		fieldPath := getStructPath(fieldType, yamlTag)
		if path != "" {
			fieldPath = path + "." + fieldPath
		}

		if field.Kind() == reflect.Struct {
			// Recursively handle nested structs
			if err := applyEnvOverrides(field, envPrefix, delimiter, normalizeDash, fieldPath, debugKeys); err != nil {
				return err
			}
		} else {
			// Check for environment variable override
			if envValue, exists := findEnvValue(envPrefix, delimiter, fieldPath, normalizeDash); exists {
				if debugKeys {
					fmt.Printf("[yamlenv] applying env override: %s = %s\n", fieldPath, envValue)
				}
				if err := setFieldValue(field, envValue); err != nil {
					return fmt.Errorf("set field %s: %w", fieldPath, err)
				}
			}
		}
	}
	return nil
}

// LoadConfig loads YAML + optional override + ENV into Target struct.
func LoadConfig(opts LoaderOptions) error {
	// Validate that delimiter is not empty when EnvPrefix is provided
	if opts.EnvPrefix != "" && opts.Delimiter == "" {
		return fmt.Errorf("delimiter cannot be empty when EnvPrefix is provided - use a non-empty delimiter like '__' for proper environment variable mapping")
	}

	// Validate target
	if opts.Target == nil {
		return fmt.Errorf("target cannot be nil")
	}
	targetValue := reflect.ValueOf(opts.Target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer to struct")
	}
	if targetValue.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}

	// 1) Load base YAML
	if err := loadYAMLFile(opts.BaseFile, opts.Target); err != nil {
		return fmt.Errorf("load base yaml: %w", err)
	}

	// 2) Load optional local YAML (merges with base)
	if opts.LocalFile != "" {
		if _, err := os.Stat(opts.LocalFile); err == nil {
			if err := loadYAMLFile(opts.LocalFile, opts.Target); err != nil {
				return fmt.Errorf("load local yaml: %w", err)
			}
		}
	}

	// 3) Apply environment variable overrides
	if err := applyEnvOverrides(targetValue, opts.EnvPrefix, opts.Delimiter, opts.NormalizeDash, "", opts.DebugKeys); err != nil {
		return fmt.Errorf("apply env overrides: %w", err)
	}

	return nil
}