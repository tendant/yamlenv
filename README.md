# yamlenv

A Go library for loading configuration from multiple sources with priority-based merging: YAML files, local overrides, and environment variables.

## Features

- **Multi-source configuration**: Load from base YAML, optional local overrides, and environment variables
- **Priority-based merging**: Environment variables override local YAML, which overrides base YAML
- **Flexible environment mapping**: Configurable prefix and delimiter for environment variable mapping
- **Struct-based configuration**: Direct unmarshaling into Go structs with koanf tags
- **Built on koanf**: Leverages the powerful [koanf](https://github.com/knadh/koanf) configuration library

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
        Name string `koanf:"name"`
        Port int    `koanf:"port"`
    } `koanf:"app"`
    DB struct {
        Host string `koanf:"host"`
        Port int    `koanf:"port"`
    } `koanf:"db"`
}

func main() {
    var cfg Config
    err := yamlenv.LoadConfig(yamlenv.LoaderOptions{
        BaseFile:  "config.yaml",
        LocalFile: "config.local.yaml", // optional
        EnvPrefix: "MYAPP_",
        Delimiter: "__",
        Target:    &cfg,
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
    BaseFile  string      // Required: path to base YAML file
    LocalFile string      // Optional: path to local override YAML file
    EnvPrefix string      // Environment variable prefix (e.g., "MYAPP_")
    Delimiter string      // Environment variable delimiter (e.g., "__")
    Target    interface{} // Pointer to struct to unmarshal into
}
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
        Name  string `koanf:"name"`
        Port  int    `koanf:"port"`
        Debug bool   `koanf:"debug"`
    } `koanf:"app"`
    
    Database struct {
        Host string `koanf:"host"`
        Port int    `koanf:"port"`
        Name string `koanf:"name"`
    } `koanf:"database"`
    
    Timeout time.Duration `koanf:"timeout"`
}
```

### 4. Loading configuration

```go
func loadConfig() (*Config, error) {
    var cfg Config
    
    err := yamlenv.LoadConfig(yamlenv.LoaderOptions{
        BaseFile:  "config.yaml",
        LocalFile: "config.local.yaml",
        EnvPrefix: "MYAPP_",
        Delimiter: "__",
        Target:    &cfg,
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
5. **Use appropriate koanf tags**: Ensure your struct fields have proper `koanf` tags

## License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
