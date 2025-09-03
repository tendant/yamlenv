package main

import (
	"fmt"
	"time"

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
	Timeout time.Duration `koanf:"timeout"`
}

func main() {
	var cfg Config
	err := yamlenv.LoadConfig(yamlenv.LoaderOptions{
		BaseFile:  "config.yaml",
		LocalFile: "config.local.yaml",
		EnvPrefix: "SAMPLE_",
		Delimiter: "__",
		Target:    &cfg,
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loaded config: %+v\n", cfg)
}
