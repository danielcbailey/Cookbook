package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	OpenAIKey string `yaml:"openai_key"`
}

var config *Config

func LoadConfig() (*Config, error) {
	config = &Config{}
	b, e := os.ReadFile("config.yaml")
	if e != nil {
		return nil, fmt.Errorf("failed to read config.yaml: %w", e)
	}

	if e := yaml.Unmarshal(b, config); e != nil {
		return nil, fmt.Errorf("failed to unmarshal config.yaml: %w", e)
	}

	return config, nil
}
