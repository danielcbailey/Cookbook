package foodkeeper

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/danielcbailey/Cookbook/core/config"
	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/internal/ai"
)

// EmbeddingText returns the text embedded for a product. The format must stay
// stable: it doubles as the embedding cache key, so changing it invalidates
// cached embeddings on reimport.
func EmbeddingText(p *models.FoodKeeperProduct) string {
	var b strings.Builder
	b.WriteString(p.Name)
	if p.NameSubtitle != "" {
		b.WriteString(", ")
		b.WriteString(p.NameSubtitle)
	}
	if p.CategoryName != "" {
		b.WriteString("; category: ")
		b.WriteString(p.CategoryName)
		if p.SubcategoryName != "" {
			b.WriteString(" / ")
			b.WriteString(p.SubcategoryName)
		}
	}
	if p.Keywords != "" {
		b.WriteString("; keywords: ")
		b.WriteString(p.Keywords)
	}
	return b.String()
}

// Import destructively reloads the FoodKeeper products from the JSON file at
// dataPath: all embeddings are computed first, then the existing products are
// deleted and the new ones inserted in a single short transaction, so the
// destructive window stays small and atomic. Returns the number of products
// imported. Requires OpenAI, DB, and Log providers; Cache is optional.
func Import(ctx context.Context, providers config.Providers, dataPath string) (int, error) {
	log := providers.Log()

	products, warnings, err := ParseFile(dataPath)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s: %w", dataPath, err)
	}
	for _, w := range warnings {
		log.Warn("foodkeeper data warning", slog.String("warning", w))
	}
	log.Info("parsed foodkeeper data", slog.Int("products", len(products)))

	for i, p := range products {
		p.Embedding, err = ai.ConvertToEmbedding(ctx, providers, EmbeddingText(p))
		if err != nil {
			return 0, fmt.Errorf("failed to embed product %d (%s): %w", p.ID, p.Name, err)
		}
		if (i+1)%50 == 0 {
			log.Info("computing embeddings", slog.Int("done", i+1), slog.Int("total", len(products)))
		}
	}

	tx, err := providers.DB().NewTransaction(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	if err := tx.DeleteAllFoodKeeperProducts(); err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	for _, p := range products {
		if err := tx.CreateFoodKeeperProduct(p); err != nil {
			_ = tx.Rollback()
			return 0, fmt.Errorf("failed to insert product %d (%s): %w", p.ID, p.Name, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return len(products), nil
}
