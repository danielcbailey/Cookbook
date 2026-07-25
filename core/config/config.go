package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	OpenAIKey             string `yaml:"openAIKey" env:"OPENAI_KEY"`
	TokenExpirySeconds    int    `yaml:"tokenExpirySeconds"`
	RecipeImportEnabled   bool   `yaml:"recipeImportEnabled"`
	SemanticSearchEnabled bool   `yaml:"semanticSearchEnabled"`
	HTTPPort              int    `yaml:"httpPort"`
}

func LoadConfig() (*Config, error) {
	// Load .env if present; ignore if missing
	_ = godotenv.Load()

	config := &Config{}
	err := cleanenv.ReadConfig("config.yaml", config)
	if err != nil {
		return nil, err
	}

	return config, checkConfig(config)
}

func checkConfig(config *Config) error {
	if (config.RecipeImportEnabled || config.SemanticSearchEnabled) && config.OpenAIKey == "" {
		return fmt.Errorf("openai_key required for recipe import and semantic search, but was not set")
	}

	return nil
}
