package local

import (
	"fmt"
	"sort"

	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/pkg/database"
)

func (tx *localTransaction) SearchIngredientsBySemanticSimilarity(userID int64, embedding []float32, limit int) ([]*models.Ingredient, error) {
	type scored struct {
		ingredient models.Ingredient
		score      float32
	}
	var candidates []scored
	for _, ing := range tx.data.Ingredients {
		if (ing.UserID == userID || ing.UserID == 0) && len(ing.Embedding) > 0 {
			candidates = append(candidates, scored{
				ingredient: ing,
				score:      cosineSimilarity(ing.Embedding, embedding),
			})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}
	out := make([]*models.Ingredient, len(candidates))
	for i, c := range candidates {
		cp := copyIngredientModel(c.ingredient)
		out[i] = &cp
	}
	return out, nil
}

func (tx *localTransaction) GetIngredientCategories(userID int64) ([]string, error) {
	seen := make(map[string]bool)
	for _, ing := range tx.data.Ingredients {
		if (ing.UserID == userID || ing.UserID == 0) && ing.Category != "" {
			seen[ing.Category] = true
		}
	}
	out := make([]string, 0, len(seen))
	for cat := range seen {
		out = append(out, cat)
	}
	sort.Strings(out)
	return out, nil
}

func (tx *localTransaction) GetIngredientsByCategory(userID int64, category string) ([]*models.Ingredient, error) {
	var out []*models.Ingredient
	for _, ing := range tx.data.Ingredients {
		if (ing.UserID == userID || ing.UserID == 0) && ing.Category == category {
			cp := copyIngredientModel(ing)
			out = append(out, &cp)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func (tx *localTransaction) CreateIngredient(ingredient *models.Ingredient) error {
	ingredient.ID = tx.data.nextID()
	tx.data.Ingredients[ingredient.ID] = *ingredient
	return nil
}

func (tx *localTransaction) UpdateIngredient(ingredient *models.Ingredient) error {
	existing, ok := tx.data.Ingredients[ingredient.ID]
	if !ok {
		return fmt.Errorf("ingredient %d: %w", ingredient.ID, database.ErrNotFound)
	}

	existing.Name = ingredient.Name
	existing.Category = ingredient.Category
	existing.FoodKeeperID = ingredient.FoodKeeperID
	existing.Embedding = ingredient.Embedding

	tx.data.Ingredients[ingredient.ID] = existing
	return nil
}

func (tx *localTransaction) DeleteIngredient(ingredient *models.Ingredient) error {
	delete(tx.data.Ingredients, ingredient.ID)
	return nil
}

func copyIngredientModel(ing models.Ingredient) models.Ingredient {
	ing.Embedding = copyFloat32s(ing.Embedding)
	return ing
}
