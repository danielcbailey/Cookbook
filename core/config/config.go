package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	OpenAIKey             string `yaml:"openai_key"`
	TokenExpirySeconds    int    `yaml:"token_expiry_seconds"`
	RecipeImportEnabled   bool   `yaml:"recipe_import_enabled"`
	SemanticSearchEnabled bool   `yaml:"semantic_search_enabled"`
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

	return config, checkConfig(config)
}

func checkConfig(config *Config) error {
	if (config.RecipeImportEnabled || config.SemanticSearchEnabled) && config.OpenAIKey == "" {
		return fmt.Errorf("openai_key required for recipe import and semantic search, but was not set")
	}

	return nil
}
