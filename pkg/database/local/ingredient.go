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

func (tx *localTransaction) ListIngredients(userID int64, offset, limit int) ([]*models.Ingredient, error) {
	var out []*models.Ingredient
	for _, ing := range tx.data.Ingredients {
		if ing.UserID == userID || ing.UserID == 0 {
			cp := copyIngredientModel(ing)
			out = append(out, &cp)
		}
	}
	// Map iteration is unordered, so sorting is what makes paging stable. Names
	// are unique per owner but not across the global and user-owned sets, so
	// break ties on ID to keep page boundaries identical to the postgres backend.
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].ID < out[j].ID
	})
	return applyPaging(out, offset, limit), nil
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

func (tx *localTransaction) GetIngredientByName(userID int64, name string) (*models.Ingredient, error) {
	var ing models.Ingredient
	ok := false

	for _, i := range tx.data.Ingredients {
		if (i.UserID == userID || i.UserID == 0) && i.Name == name {
			ing = i
			ok = true
			break
		}
	}

	if !ok || (ing.UserID != userID && ing.UserID != 0) {
		return nil, fmt.Errorf("ingredient %s: %w", name, database.ErrNotFound)
	}
	cp := copyIngredientModel(ing)
	return &cp, nil
}

func (tx *localTransaction) CreateIngredient(ingredient *models.Ingredient) (int64, error) {
	ingredient.ID = tx.data.nextID()
	tx.data.Ingredients[ingredient.ID] = *ingredient
	return ingredient.ID, nil
}

func (tx *localTransaction) UpdateIngredient(ingredient *models.Ingredient) error {
	existing, ok := tx.data.Ingredients[ingredient.ID]
	if !ok {
		return fmt.Errorf("ingredient %d: %w", ingredient.ID, database.ErrNotFound)
	}

	existing.Name = ingredient.Name
	existing.Category = ingredient.Category
	existing.Density = ingredient.Density
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
