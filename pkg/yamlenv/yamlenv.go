package yamlenv

import (
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type LoaderOptions struct {
	BaseFile  string // required (e.g., "config.yaml")
	LocalFile string // optional override (e.g., "config.local.yaml")
	EnvPrefix string // e.g., "FI_STOCK_"
	Delimiter string // e.g., "__"
	Target    interface{}
}

// LoadConfig loads YAML + optional override + ENV into Target struct.
func LoadConfig(opts LoaderOptions) error {
	// Create a new koanf instance for each call to avoid state pollution
	k := koanf.New(".")

	// 1) Base YAML
	if err := k.Load(file.Provider(opts.BaseFile), yaml.Parser()); err != nil {
		return fmt.Errorf("load base yaml: %w", err)
	}

	// 2) Optional local overrides
	if opts.LocalFile != "" {
		if _, err := os.Stat(opts.LocalFile); err == nil {
			if err := k.Load(file.Provider(opts.LocalFile), yaml.Parser()); err != nil {
				return fmt.Errorf("load local yaml: %w", err)
			}
		}
	}

	// 3) ENV overrides
	if opts.EnvPrefix != "" {
		mapper := func(key string) string {
			key = strings.TrimPrefix(key, opts.EnvPrefix)
			key = strings.ToLower(key)
			return strings.ReplaceAll(key, opts.Delimiter, ".")
		}
		if err := k.Load(env.Provider(opts.EnvPrefix, opts.Delimiter, mapper), nil); err != nil {
			return fmt.Errorf("load env: %w", err)
		}
	}

	// 4) Unmarshal into struct
	if err := k.Unmarshal("", opts.Target); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	return nil
}
