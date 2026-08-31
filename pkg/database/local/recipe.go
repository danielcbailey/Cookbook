package local

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/pkg/database"
)

func (tx *localTransaction) ListRecipesByUserID(userID int64, offset, limit int) ([]*models.RecipeListing, error) {
	out := tx.listingsWhere(func(r models.Recipe) bool {
		return r.UserID == userID
	})
	return applyPaging(out, offset, limit), nil
}

func (tx *localTransaction) ListRecipesBySemanticSimilarity(userID int64, embedding []float32, limit int) ([]*models.RecipeListing, error) {
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
	out := make([]*models.RecipeListing, len(candidates))
	for i, c := range candidates {
		l := recipeListing(c.recipe)
		out[i] = &l
	}
	return out, nil
}

func (tx *localTransaction) ListRecipesByCategory(userID int64, category string, limit int) ([]*models.RecipeListing, error) {
	out := tx.listingsWhere(func(r models.Recipe) bool {
		return r.UserID == userID && r.Category == category
	})
	return applyPaging(out, 0, limit), nil
}

func (tx *localTransaction) ListRecipesByProtein(userID int64, protein string, limit int) ([]*models.RecipeListing, error) {
	out := tx.listingsWhere(func(r models.Recipe) bool {
		return r.UserID == userID && r.Protein == protein
	})
	return applyPaging(out, 0, limit), nil
}

func (tx *localTransaction) ListRecipesByMeal(userID int64, meal string, limit int) ([]*models.RecipeListing, error) {
	out := tx.listingsWhere(func(r models.Recipe) bool {
		return r.UserID == userID && r.SuggestedMeal == meal
	})
	return applyPaging(out, 0, limit), nil
}

func (tx *localTransaction) ListRecipesByTag(userID int64, tag string, limit int) ([]*models.RecipeListing, error) {
	out := tx.listingsWhere(func(r models.Recipe) bool {
		if r.UserID != userID {
			return false
		}
		for _, t := range r.Tags {
			if t.Name == tag {
				return true
			}
		}
		return false
	})
	return applyPaging(out, 0, limit), nil
}

func (tx *localTransaction) ListRecipesByTitleSearch(userID int64, query string, limit int) ([]*models.RecipeListing, error) {
	needle := strings.ToLower(query)
	out := tx.listingsWhere(func(r models.Recipe) bool {
		return r.UserID == userID && strings.Contains(strings.ToLower(r.Title), needle)
	})
	return applyPaging(out, 0, limit), nil
}

func (tx *localTransaction) ListRecipeCategories(userID int64) ([]string, error) {
	return tx.distinctRecipeValues(userID, func(r models.Recipe) string { return r.Category }), nil
}

func (tx *localTransaction) ListRecipeProteins(userID int64) ([]string, error) {
	return tx.distinctRecipeValues(userID, func(r models.Recipe) string { return r.Protein }), nil
}

func (tx *localTransaction) ListRecipeMealtimes(userID int64) ([]string, error) {
	return tx.distinctRecipeValues(userID, func(r models.Recipe) string { return r.SuggestedMeal }), nil
}

// ListRecipeTags reads the tag store rather than scanning recipes: reconcileTags
// never prunes it, so a tag the user once applied stays listed after its last
// recipe drops it. This matches the postgres backend, which reads recipe_tags.
func (tx *localTransaction) ListRecipeTags(userID int64) ([]string, error) {
	out := make([]string, 0, len(tx.data.Tags))
	for _, t := range tx.data.Tags {
		if t.UserID == userID {
			out = append(out, t.Name)
		}
	}
	sort.Strings(out)
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
	existing.TotalTime = recipe.TotalTime
	existing.Nutrition = recipe.Nutrition
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

// recipeListing projects a stored recipe down to the fields a listing carries.
func recipeListing(r models.Recipe) models.RecipeListing {
	return models.RecipeListing{
		ID:       r.ID,
		UserID:   r.UserID,
		Title:    r.Title,
		ImageURL: r.ImageURL,
		Servings: r.Servings,
		Calories: r.Nutrition.Calories,
		Time:     r.TotalTime,
		Tags:     copyTags(r.Tags),
	}
}

// listingsWhere collects the recipes matching keep, newest first. The store is a
// map, so its iteration order is random and this sort is what makes results
// stable; the tie-break on ID is what lets offset paging return disjoint pages
// when several recipes share a CreatedAt.
func (tx *localTransaction) listingsWhere(keep func(models.Recipe) bool) []*models.RecipeListing {
	type entry struct {
		listing   models.RecipeListing
		createdAt int64
	}
	var entries []entry
	for _, r := range tx.data.Recipes {
		if keep(r) {
			entries = append(entries, entry{listing: recipeListing(r), createdAt: r.CreatedAt})
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].createdAt != entries[j].createdAt {
			return entries[i].createdAt > entries[j].createdAt
		}
		return entries[i].listing.ID > entries[j].listing.ID
	})

	out := make([]*models.RecipeListing, len(entries))
	for i := range entries {
		out[i] = &entries[i].listing
	}
	return out
}

// applyPaging windows an already-sorted slice. A negative offset is clamped to
// zero and a limit of zero or less means unlimited, matching the postgres
// backend.
func applyPaging[T any](items []T, offset, limit int) []T {
	offset = max(offset, 0)
	if offset >= len(items) {
		return nil
	}
	items = items[offset:]
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

// distinctRecipeValues returns the sorted, de-duplicated non-empty values of one
// recipe field across a user's recipes.
func (tx *localTransaction) distinctRecipeValues(userID int64, value func(models.Recipe) string) []string {
	seen := make(map[string]bool)
	for _, r := range tx.data.Recipes {
		if r.UserID != userID {
			continue
		}
		if v := value(r); v != "" {
			seen[v] = true
		}
	}
	out := make([]string, 0, len(seen))
	for v := range seen {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}
