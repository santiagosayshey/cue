package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Library struct {
	Path string `yaml:"path"`
	Type string `yaml:"type"`
}

type Config struct {
	Library  []Library `yaml:"libraries"`
	Database []string  `yaml:"databases"`
}

func Load(path string) (Config, error) {
	cfgData, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	err = yaml.Unmarshal(cfgData, &cfg)
	if err != nil {
		return Config{}, err
	}
	return cfg, nil
}
