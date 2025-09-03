package main

import (
	"fmt"
	"time"

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
	Timeout time.Duration `yaml:"timeout"`
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
