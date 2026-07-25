package recipes

import (
	"context"
	"fmt"

	"github.com/danielcbailey/Cookbook/core/config"
	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/internal/ai"
)

func SaveRecipe(ctx context.Context, providers config.Providers, recipe *models.Recipe) error {
	// First, match ingredients with existing ones or create new ingredients scoped to the user
	tx, err := providers.DB().NewTransaction(ctx)
	if err != nil {
		return err
	}
	for i, ingredient := range recipe.Ingredients {
		embedding, err := ai.ConvertToEmbedding(ctx, providers, ingredient.Ingredient.Name)
		if err != nil {
			return fmt.Errorf("failed to get embedding of ingredient name: %w", err)
		}

		candidates, err := tx.SearchIngredientsBySemanticSimilarity(providers.User().ID, embedding, 10)
		if err != nil {
			return fmt.Errorf("failed to search existing ingredients: %w", err)
		}

		match, err := ai.AssociateIngredient(ctx, providers, candidates, recipe.Title, ingredient.Ingredient.Name)
		if err != nil {
			return fmt.Errorf("failed to associate ingredient: %w", err)
		}

		if match.ID == 0 {
			// Must create the ingredient

		}
	}
}
