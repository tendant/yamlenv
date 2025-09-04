# yamlenv

A Go library for loading configuration from multiple sources with priority-based merging: YAML files, local overrides, and environment variables.

## Features

- **Multi-source configuration**: Load from base YAML, optional local overrides, and environment variables
- **Priority-based merging**: Environment variables override local YAML, which overrides base YAML
- **Flexible environment mapping**: Configurable prefix and delimiter for environment variable mapping
- **Struct-based configuration**: Direct unmarshaling into Go structs with standard yaml tags
- **Embedded filesystem support**: Load configuration files from Go's embed.FS for single binary deployments
- **Generic IO interface**: Use any data source that implements `io.Reader` - files, embedded files, HTTP responses, or in-memory data
- **Direct implementation**: Lightweight implementation using only `gopkg.in/yaml.v3` without heavy dependencies

## Installation

```bash
go get github.com/tendant/yamlenv
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/tendant/yamlenv/pkg/yamlenv"
)

type Config struct {
    App struct {
        Name string `yaml:"name"`
        Port int    `yaml:"port"`
    } `yaml:"app"`
    DB struct {
        Host string `yaml:"host"`
        Port int    `yaml:"port"`
    } `yaml:"db"`
}

func main() {
    var cfg Config
    err := yamlenv.LoadConfig(yamlenv.LoaderOptions{
        BaseSource:  yamlenv.FileSource("config.yaml"),
        LocalSource: yamlenv.FileSource("config.local.yaml"), // optional
        EnvPrefix:   "MYAPP_",
        Delimiter:   "__",
        Target:      &cfg,
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Config: %+v\n", cfg)
}
```

## Configuration Priority

yamlenv loads configuration in the following order (later sources override earlier ones):

1. **Base YAML file** (required) - e.g., `config.yaml`
2. **Local YAML file** (optional) - e.g., `config.local.yaml`
3. **Environment variables** (optional) - with configurable prefix and delimiter

## API Reference

### LoaderOptions

```go
type LoaderOptions struct {
    BaseSource  ConfigSource // Required: function that returns base config reader
    LocalSource ConfigSource // Optional: function that returns local override config reader
    EnvPrefix   string       // Environment variable prefix (e.g., "MYAPP_")
    Delimiter   string       // Environment variable delimiter (e.g., "__")
    Target      interface{}  // Pointer to struct to unmarshal into
}
```

### ConfigSource

`ConfigSource` is a function type that returns an `io.ReadCloser`:

```go
type ConfigSource func() (io.ReadCloser, error)
```

### Built-in ConfigSource Factories

yamlenv provides several built-in factories to create `ConfigSource` instances:

```go
// FileSource creates a ConfigSource from a file path
func FileSource(filename string) ConfigSource

// EmbedSource creates a ConfigSource from an embedded filesystem
func EmbedSource(fsys fs.FS, filename string) ConfigSource

// ReaderSource creates a ConfigSource from an io.Reader (useful for testing)
func ReaderSource(reader io.Reader) ConfigSource
```

### LoadConfig

```go
func LoadConfig(opts LoaderOptions) error
```

Loads configuration from multiple sources and unmarshals into the target struct.

## Complete Example

### 1. Base configuration file (`config.yaml`)

```yaml
app:
  name: myapp
  port: 8080
  debug: false

database:
  host: localhost
  port: 5432
  name: myapp_db

timeout: 30s
```

### 2. Local override file (`config.local.yaml`) - optional

```yaml
app:
  debug: true
  port: 3000

database:
  host: dev-db.local
```

### 3. Go struct definition

```go
type Config struct {
    App struct {
        Name  string `yaml:"name"`
        Port  int    `yaml:"port"`
        Debug bool   `yaml:"debug"`
    } `yaml:"app"`
    
    Database struct {
        Host string `yaml:"host"`
        Port int    `yaml:"port"`
        Name string `yaml:"name"`
    } `yaml:"database"`
    
    Timeout time.Duration `yaml:"timeout"`
}
```

### 4. Loading configuration

```go
func loadConfig() (*Config, error) {
    var cfg Config
    
    err := yamlenv.LoadConfig(yamlenv.LoaderOptions{
        BaseSource:  yamlenv.FileSource("config.yaml"),
        LocalSource: yamlenv.FileSource("config.local.yaml"),
        EnvPrefix:   "MYAPP_",
        Delimiter:   "__",
        Target:      &cfg,
    })
    
    return &cfg, err
}
```

## Environment Variable Mapping

Environment variables are mapped to YAML keys using the following rules:

- **Prefix**: Removed from the environment variable name
- **Case**: Converted to lowercase
- **Delimiter**: Replaced with dots (`.`) to create nested keys

### Examples

With `EnvPrefix: "MYAPP_"` and `Delimiter: "__"`:

| Environment Variable | YAML Key | Description |
|---------------------|----------|-------------|
| `MYAPP_APP__NAME` | `app.name` | Sets `app.name` |
| `MYAPP_APP__PORT` | `app.port` | Sets `app.port` |
| `MYAPP_DATABASE__HOST` | `database.host` | Sets `database.host` |
| `MYAPP_TIMEOUT` | `timeout` | Sets root-level `timeout` |

### Setting environment variables

```bash
export MYAPP_APP__NAME="production-app"
export MYAPP_APP__PORT="9000"
export MYAPP_DATABASE__HOST="prod-db.example.com"
export MYAPP_DATABASE__PORT="5433"
```

## Embedded Filesystem Support

yamlenv supports loading configuration files from Go's embedded filesystem (`embed.FS`), which is useful for building single-binary applications with embedded configuration files.

### Example with embed.FS

```go
package main

import (
    "embed"
    "fmt"
    "github.com/tendant/yamlenv/pkg/yamlenv"
)

//go:embed configs/*.yaml
var configFS embed.FS

type Config struct {
    App struct {
        Name string `yaml:"name"`
        Port int    `yaml:"port"`
    } `yaml:"app"`
    DB struct {
        Host string `yaml:"host"`
        Port int    `yaml:"port"`
    } `yaml:"db"`
}

func main() {
    var cfg Config
    
    err := yamlenv.LoadConfig(yamlenv.LoaderOptions{
        BaseSource:  yamlenv.EmbedSource(configFS, "configs/config.yaml"),
        LocalSource: yamlenv.EmbedSource(configFS, "configs/config.local.yaml"), // optional
        EnvPrefix:   "MYAPP_",
        Delimiter:   "__",
        Target:      &cfg,
    })
    
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Config: %+v\n", cfg)
}
```

### Directory Structure for Embedded Configs

```
your-project/
├── main.go
├── configs/
│   ├── config.yaml       # base configuration
│   └── config.local.yaml # optional local overrides
└── go.mod
```

With this setup, both configuration files are embedded into your binary at build time, and you can still override values using environment variables at runtime.

## Advanced Usage

### Using Custom ConfigSources

You can create custom `ConfigSource` functions for any data source:

```go
// Load config from HTTP endpoint
func HttpSource(url string) yamlenv.ConfigSource {
    return func() (io.ReadCloser, error) {
        resp, err := http.Get(url)
        if err != nil {
            return nil, err
        }
        return resp.Body, nil
    }
}

// Load config from in-memory string
func StringSource(content string) yamlenv.ConfigSource {
    return yamlenv.ReaderSource(strings.NewReader(content))
}

// Usage
var cfg Config
err := yamlenv.LoadConfig(yamlenv.LoaderOptions{
    BaseSource: HttpSource("https://config.example.com/base.yaml"),
    LocalSource: StringSource("app:\n  debug: true"),
    Target: &cfg,
})
```

### Mixing Different Sources

You can mix and match different source types:

```go
//go:embed base-config.yaml
var baseConfigFS embed.FS

var cfg Config
err := yamlenv.LoadConfig(yamlenv.LoaderOptions{
    BaseSource:  yamlenv.EmbedSource(baseConfigFS, "base-config.yaml"), // From embedded FS
    LocalSource: yamlenv.FileSource("local.yaml"),                      // From file system
    EnvPrefix:   "APP_",
    Delimiter:   "__",
    Target:      &cfg,
})
```

## Error Handling

The `LoadConfig` function returns detailed errors for different failure scenarios:

```go
err := yamlenv.LoadConfig(opts)
if err != nil {
    // Handle specific error types
    fmt.Printf("Configuration error: %v\n", err)
    return err
}
```

Common error scenarios:
- Base YAML file not found or invalid
- Local YAML file invalid (if specified and exists)
- Environment variable parsing errors
- Struct unmarshaling errors

## Best Practices

1. **Always specify a base file**: The base YAML file is required and should contain sensible defaults
2. **Use local files for development**: Keep `config.local.yaml` in `.gitignore` for local development overrides
3. **Environment variables for deployment**: Use environment variables to override settings in different deployment environments
4. **Validate configuration**: Add validation logic after loading configuration
5. **Use standard yaml tags**: Ensure your struct fields have proper `yaml` tags

## License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
