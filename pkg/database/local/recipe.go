package local

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/pkg/database"
)

func (tx *localTransaction) ListRecipesByUserID(userID int64) ([]*models.Recipe, error) {
	var out []*models.Recipe
	for _, r := range tx.data.Recipes {
		if r.UserID == userID {
			cp := copyRecipeShallow(r)
			out = append(out, &cp)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].CreatedAt != out[j].CreatedAt {
			return out[i].CreatedAt > out[j].CreatedAt
		}
		return out[i].ID > out[j].ID
	})
	return out, nil
}

func (tx *localTransaction) SearchRecipesBySemanticSimilarity(userID int64, embedding []float32, limit int) ([]*models.Recipe, error) {
	type scored struct {
		recipe models.Recipe
		score  float32
	}
	var candidates []scored
	for _, r := range tx.data.Recipes {
		if r.UserID == userID && len(r.Embedding) > 0 {
			candidates = append(candidates, scored{
				recipe: r,
				score:  cosineSimilarity(r.Embedding, embedding),
			})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}
	out := make([]*models.Recipe, len(candidates))
	for i, c := range candidates {
		cp := copyRecipeShallow(c.recipe)
		out[i] = &cp
	}
	return out, nil
}

func (tx *localTransaction) GetRecipeByID(recipeID int64) (*models.Recipe, error) {
	r, ok := tx.data.Recipes[recipeID]
	if !ok {
		return nil, fmt.Errorf("recipe %d: %w", recipeID, database.ErrNotFound)
	}
	cp := copyRecipe(r)
	return &cp, nil
}

func (tx *localTransaction) CreateRecipe(recipe *models.Recipe) (int64, error) {
	recipe.ID = tx.data.nextID()
	now := time.Now().Unix()
	recipe.CreatedAt = now
	recipe.UpdatedAt = now
	tx.reconcileTags(recipe)
	if recipe.Steps == nil {
		recipe.Steps = []models.RecipeStep{}
	}
	if recipe.Ingredients == nil {
		recipe.Ingredients = []models.RecipeIngredient{}
	}
	for i := range recipe.Ingredients {
		recipe.Ingredients[i].ID = tx.data.nextID()
	}
	tx.data.Recipes[recipe.ID] = *recipe
	return recipe.ID, nil
}

func (tx *localTransaction) UpdateRecipe(recipe *models.Recipe) error {
	existing, ok := tx.data.Recipes[recipe.ID]
	if !ok {
		return fmt.Errorf("recipe %d: %w", recipe.ID, database.ErrNotFound)
	}

	tx.reconcileTags(recipe)

	existing.Title = recipe.Title
	existing.Description = recipe.Description
	existing.ImageURL = recipe.ImageURL
	existing.Author = recipe.Author
	existing.Publisher = recipe.Publisher
	existing.Servings = recipe.Servings
	existing.Category = recipe.Category
	existing.Protein = recipe.Protein
	existing.Cuisine = recipe.Cuisine
	existing.SuggestedMeal = recipe.SuggestedMeal
	existing.Embedding = recipe.Embedding
	existing.Tags = recipe.Tags
	existing.UpdatedAt = time.Now().Unix()

	tx.data.Recipes[recipe.ID] = existing
	return nil
}

func (tx *localTransaction) DeleteRecipe(recipeID int64) error {
	delete(tx.data.Recipes, recipeID)
	return nil
}

func (tx *localTransaction) CreateOrUpdateRecipeStep(step *models.RecipeStep) (int64, error) {
	recipe, ok := tx.data.Recipes[step.RecipeID]
	if !ok {
		return 0, fmt.Errorf("recipe %d: %w", step.RecipeID, database.ErrNotFound)
	}

	idx := -1
	for i, s := range recipe.Steps {
		if s.Index == step.Index {
			idx = i
			break
		}
	}

	if idx >= 0 {
		step.ID = recipe.Steps[idx].ID
	} else {
		step.ID = tx.data.nextID()
		recipe.Steps = append(recipe.Steps, models.RecipeStep{})
		idx = len(recipe.Steps) - 1
	}

	for i := range step.Ingredients {
		step.Ingredients[i].ID = tx.data.nextID()
		step.Ingredients[i].StepID = step.ID
	}
	for i := range step.Times {
		step.Times[i].StepID = step.ID
	}
	if step.Ingredients == nil {
		step.Ingredients = []models.RecipeIngredient{}
	}
	if step.Times == nil {
		step.Times = []models.RecipeTime{}
	}

	recipe.Steps[idx] = *step
	tx.data.Recipes[step.RecipeID] = recipe
	return step.ID, nil
}

func (tx *localTransaction) DeleteRecipeStep(step *models.RecipeStep) error {
	var recipe models.Recipe
	var recipeID int64
	found := false

	if step.RecipeID != 0 {
		r, ok := tx.data.Recipes[step.RecipeID]
		if !ok {
			return fmt.Errorf("recipe %d: %w", step.RecipeID, database.ErrNotFound)
		}
		recipe = r
		recipeID = step.RecipeID
		found = true
	} else {
		for id, r := range tx.data.Recipes {
			for _, s := range r.Steps {
				if s.ID == step.ID {
					recipe = r
					recipeID = id
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}

	if !found {
		return fmt.Errorf("recipe step %d: %w", step.ID, database.ErrNotFound)
	}

	steps := make([]models.RecipeStep, 0, len(recipe.Steps))
	for _, s := range recipe.Steps {
		if s.ID != step.ID {
			steps = append(steps, s)
		}
	}
	recipe.Steps = steps
	tx.data.Recipes[recipeID] = recipe
	return nil
}

// helpers

func (tx *localTransaction) reconcileTags(recipe *models.Recipe) {
	for i, tag := range recipe.Tags {
		var existingID int64
		for id, t := range tx.data.Tags {
			if t.UserID == recipe.UserID && t.Name == tag.Name {
				existingID = id
				break
			}
		}
		if existingID != 0 {
			recipe.Tags[i].ID = existingID
			recipe.Tags[i].UserID = recipe.UserID
		} else {
			newID := tx.data.nextID()
			recipe.Tags[i].ID = newID
			recipe.Tags[i].UserID = recipe.UserID
			tx.data.Tags[newID] = recipe.Tags[i]
		}
	}
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dot, magA, magB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		magA += float64(a[i]) * float64(a[i])
		magB += float64(b[i]) * float64(b[i])
	}
	denom := math.Sqrt(magA) * math.Sqrt(magB)
	if denom == 0 {
		return 0
	}
	return float32(dot / denom)
}

func copyRecipe(r models.Recipe) models.Recipe {
	r.Embedding = copyFloat32s(r.Embedding)
	r.Tags = copyTags(r.Tags)
	r.Ingredients = copyIngredients(r.Ingredients)
	r.Steps = copySteps(r.Steps)
	return r
}

func copySteps(steps []models.RecipeStep) []models.RecipeStep {
	if steps == nil {
		return nil
	}
	out := make([]models.RecipeStep, len(steps))
	for i, s := range steps {
		s.Ingredients = copyIngredients(s.Ingredients)
		s.Times = copyTimes(s.Times)
		out[i] = s
	}
	return out
}

func copyIngredients(ings []models.RecipeIngredient) []models.RecipeIngredient {
	if ings == nil {
		return nil
	}
	out := make([]models.RecipeIngredient, len(ings))
	for i, ri := range ings {
		ri.Ingredient.Embedding = copyFloat32s(ri.Ingredient.Embedding)
		out[i] = ri
	}
	return out
}

func copyTags(tags []models.RecipeTag) []models.RecipeTag {
	if tags == nil {
		return nil
	}
	out := make([]models.RecipeTag, len(tags))
	copy(out, tags)
	return out
}

func copyTimes(times []models.RecipeTime) []models.RecipeTime {
	if times == nil {
		return nil
	}
	out := make([]models.RecipeTime, len(times))
	copy(out, times)
	return out
}

func copyFloat32s(s []float32) []float32 {
	if s == nil {
		return nil
	}
	out := make([]float32, len(s))
	copy(out, s)
	return out
}

func copyRecipeShallow(r models.Recipe) models.Recipe {
	tags := make([]models.RecipeTag, len(r.Tags))
	copy(tags, r.Tags)
	r.Tags = tags
	r.Steps = nil
	r.Ingredients = nil
	r.Embedding = nil
	return r
}
