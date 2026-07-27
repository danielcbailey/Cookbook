// Command foodkeeper-import destructively reimports the USDA FoodKeeper data
// into the local database backend: the existing products are cleared and
// rewritten from the current file content, with embeddings computed for
// semantic ingredient association.
//
// Run it from the directory containing config.yaml; the config must provide
// openAIKey for computing embeddings.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/danielcbailey/Cookbook/core/config"
	"github.com/danielcbailey/Cookbook/pkg/database/local"
	"github.com/danielcbailey/Cookbook/pkg/foodkeeper"
	"github.com/openai/openai-go/v3"
	openaioption "github.com/openai/openai-go/v3/option"
)

func main() {
	dataPath := flag.String("data", "assets/databases/foodkeeper.json", "path to the FoodKeeper JSON file")
	storeDir := flag.String("store", ".", "directory containing the local database store")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}
	if cfg.OpenAIKey == "" {
		logger.Error("openAIKey must be configured to compute embeddings")
		os.Exit(1)
	}

	if err := os.MkdirAll(*storeDir, 0o755); err != nil {
		logger.Error("failed to create store directory", slog.Any("error", err))
		os.Exit(1)
	}
	db, err := local.NewLocal(*storeDir)
	if err != nil {
		logger.Error("failed to open local database", slog.Any("error", err))
		os.Exit(1)
	}

	openAIClient := openai.NewClient(openaioption.WithAPIKey(cfg.OpenAIKey))
	// No cache provider: ConvertToEmbedding fetches from OpenAI every run.
	providers := config.NewProviders().
		WithConfig(cfg).
		WithOpenAIClient(&openAIClient).
		WithDB(db).
		WithLog(logger)

	count, err := foodkeeper.Import(context.Background(), providers, *dataPath)
	if err != nil {
		logger.Error("foodkeeper import failed", slog.Any("error", err))
		os.Exit(1)
	}
	logger.Info("foodkeeper import complete", slog.Int("products", count))
}
