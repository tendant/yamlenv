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
	BaseFile       string // required
	LocalFile      string // optional
	EnvPrefix      string // e.g. "WORKING_"
	Delimiter      string // nesting delimiter in env, e.g. "__"; "" = no nesting
	Target         any    // &cfg
	NormalizeDash  bool   // if true, convert "_" in ENV path to "-" in YAML keys (for kebab-case YAML like "app-name")
	ForceLowerYAML bool   // if true, normalize YAML keys to lowercase to match ENV mapping
	DebugKeys      bool   // if true, print final keys for debugging
}

// normalizeAllKeys rebuilds the koanf tree with normalized keys.
func normalizeAllKeys(src *koanf.Koanf, normalizer func(string) string) *koanf.Koanf {
	dst := koanf.New(".")
	for _, k := range src.Keys() {
		v := src.Get(k)
		nk := normalizer(k)
		dst.Set(nk, v)
	}
	return dst
}

// LoadConfig loads YAML + optional override + ENV into Target struct.
func LoadConfig(opts LoaderOptions) error {
	// Validate that delimiter is not empty when EnvPrefix is provided
	if opts.EnvPrefix != "" && opts.Delimiter == "" {
		return fmt.Errorf("delimiter cannot be empty when EnvPrefix is provided - use a non-empty delimiter like '__' for proper environment variable mapping")
	}

	// Always use a fresh instance.
	k := koanf.New(".")

	// 1) Base YAML
	if err := k.Load(file.Provider(opts.BaseFile), yaml.Parser()); err != nil {
		return fmt.Errorf("load base yaml: %w", err)
	}

	// 2) Optional local YAML
	if opts.LocalFile != "" {
		if _, err := os.Stat(opts.LocalFile); err == nil {
			if err := k.Load(file.Provider(opts.LocalFile), yaml.Parser()); err != nil {
				return fmt.Errorf("load local yaml: %w", err)
			}
		}
	}

	// 2.5) Normalize YAML keys (case-insensitive stability).
	// Many teams keep YAML keys lowercase. If your YAML isn't, enabling this avoids shadow trees.
	if opts.ForceLowerYAML {
		k = normalizeAllKeys(k, func(key string) string {
			return strings.ToLower(key)
		})
	}

	// 3) ENV overrides (LAST). Map: WORKING_APP__NAME -> app.name
	// koanf/env splits on delimiter; empty delimiter splits every rune (bad).
	delimiterForProvider := opts.Delimiter
	if delimiterForProvider == "" {
		delimiterForProvider = "§§" // any token that won't appear in your env names
	}

	// Path transformer: strip prefix, lower-case, replace delimiter with dot,
	// and optionally translate _ -> - for kebab-case YAML keys.
	transform := func(key string) string {
		key = strings.TrimPrefix(key, opts.EnvPrefix)
		key = strings.ToLower(key)
		if opts.Delimiter != "" {
			key = strings.ReplaceAll(key, opts.Delimiter, ".")
		}
		if opts.NormalizeDash {
			// Convert env underscores in path segments to dashes, e.g. app_name -> app-name
			parts := strings.Split(key, ".")
			for i := range parts {
				parts[i] = strings.ReplaceAll(parts[i], "_", "-")
			}
			key = strings.Join(parts, ".")
		}
		return key
	}

	if err := k.Load(env.Provider(opts.EnvPrefix, delimiterForProvider, transform), nil); err != nil {
		return fmt.Errorf("load env: %w", err)
	}

	// Optional debug: list final visible keys
	if opts.DebugKeys {
		for _, key := range k.Keys() {
			fmt.Println("[yamlenv] key:", key)
		}
	}

	// 4) Unmarshal into typed struct
	// WORKAROUND: Create a fresh koanf instance and manually copy all values to avoid unmarshaling bug
	freshK := koanf.New(".")
	for key, value := range k.All() {
		freshK.Set(key, value)
	}
	
	if err := freshK.Unmarshal("", opts.Target); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	return nil
}
