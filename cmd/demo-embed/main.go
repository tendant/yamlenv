package main

import (
	"embed"
	"fmt"
	"time"

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
	Timeout time.Duration `yaml:"timeout"`
}

func main() {
	fmt.Println("=== Demo: Loading config from embedded filesystem ===")
	
	var cfg Config
	err := yamlenv.LoadConfig(yamlenv.LoaderOptions{
		BaseSource:  yamlenv.EmbedSource(configFS, "configs/config.yaml"),
		LocalSource: yamlenv.EmbedSource(configFS, "configs/config.local.yaml"),
		EnvPrefix:   "DEMO_",
		Delimiter:   "__",
		Target:      &cfg,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loaded config from embedded FS: %+v\n", cfg)
	
	fmt.Println("\n=== Demo: Environment variable overrides ===")
	fmt.Println("Set environment variables to see overrides:")
	fmt.Println("export DEMO_APP__PORT=9999")
	fmt.Println("export DEMO_DB__HOST=override-db")
	fmt.Println("Then run this demo again to see the overrides in action.")
}