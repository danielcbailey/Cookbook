package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/danielcbailey/Cookbook/core/config"
	"github.com/danielcbailey/Cookbook/internal/ai"
	"github.com/openai/openai-go/v3"
	openaioption "github.com/openai/openai-go/v3/option"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	providers := config.NewProviders().WithConfig(cfg)

	openAIClient := openai.NewClient(openaioption.WithAPIKey(cfg.OpenAIKey))
	providers = providers.WithOpenAIClient(&openAIClient)

	jpegImage, err := os.ReadFile("testRecipePhoto.jpg")
	if err != nil {
		slog.Error("failed to read image file", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	recipe, err := ai.ExtractRecipeFromPhoto(ctx, providers, jpegImage, "image/jpeg")
	if err != nil {
		slog.Error("failed to extract recipe from photo", "error", err)
		os.Exit(1)
	}
	b, err := json.MarshalIndent(recipe, "", "  ")
	if err != nil {
		slog.Error("failed to marshal recipe", "error", err)
		os.Exit(1)
	}
	fmt.Println(string(b))
}
