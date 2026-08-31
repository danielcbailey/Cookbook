package recipes

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/danielcbailey/Cookbook/core/apicommon"
	"github.com/danielcbailey/Cookbook/core/config"
	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/internal/ai"
	"github.com/danielcbailey/Cookbook/pkg/database"
	"github.com/google/uuid"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const maxRecipeImagePixels = 2_000_000
const maxRecipeImageSize = 20 * apicommon.MiB
const maxRecipeDownloadImageArea = 50_000_000
const recipeImageJPEGQuality = 85

func SaveRecipe(ctx context.Context, providers config.Providers, recipe *models.Recipe) (int64, error) {
	// First, match ingredients with existing ones or create new ingredients scoped to the user
	tx, err := providers.DB().NewTransaction(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var existing *models.Recipe
	if recipe.ID != 0 {
		existing, err = tx.GetRecipeByID(recipe.ID)
		if err != nil && errors.Is(err, database.ErrNotFound) {
			tx.Rollback()
			return 0, apicommon.NewUserFacingError("invalid recipe ID")
		} else if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to query database for existing recipe: %w", err)
		}

		// Existing recipe must belong to the same user
		if existing.UserID != providers.User().ID {
			tx.Rollback()
			return 0, apicommon.NewUserFacingError("invalid recipe ID")
		}
	} else {
		recipe.UserID = providers.User().ID
	}

	err = checkRecipeIngredients(recipe, existing)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Updating ingredients with IDs
	err = getOrCreateIngredients(ctx, providers, tx, recipe)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Generating new embedding for recipe
	err = updateRecipeEmbedding(ctx, providers, recipe, existing)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	// Saving new images
	savedImages, err := saveImages(ctx, providers, recipe)
	if err != nil {
		tx.Rollback()
		rollbackSavedImages(ctx, providers, savedImages)
		return 0, err
	}

	// Updating database
	if existing == nil {
		id, err := tx.CreateRecipe(recipe)
		if err != nil {
			tx.Rollback()
			rollbackSavedImages(ctx, providers, savedImages)
			return 0, err
		}
		recipe.ID = id
	} else {
		err = tx.UpdateRecipe(recipe)
		if err != nil {
			tx.Rollback()
			rollbackSavedImages(ctx, providers, savedImages)
			return 0, err
		}
	}

	err = updateRecipeSteps(tx, recipe, existing)
	if err != nil {
		tx.Rollback()
		rollbackSavedImages(ctx, providers, savedImages)
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		rollbackSavedImages(ctx, providers, savedImages)
		return 0, fmt.Errorf("failed to commit recipe changes: %w", err)
	}

	// Deleting former images now that the new changes have been finalized
	removeOldImages(ctx, providers, existing, recipe)

	return recipe.ID, nil
}

func updateRecipeEmbedding(ctx context.Context, providers config.Providers, recipe, existing *models.Recipe) error {
	if providers.Config().SemanticSearchEnabled && shouldGenerateNewEmbedding(recipe, existing) {
		contentBuilder := strings.Builder{}
		contentBuilder.WriteString(recipe.Title)
		contentBuilder.WriteString(" Description: ")
		contentBuilder.WriteString(recipe.Description)
		contentBuilder.WriteString(" Tags: ")
		for i, tag := range recipe.Tags {
			if i > 0 {
				contentBuilder.WriteString(", ")
			}
			contentBuilder.WriteString(tag.Name)
		}

		embedding, err := ai.ConvertToEmbedding(ctx, providers, contentBuilder.String())
		if err != nil {
			return err
		}
		recipe.Embedding = embedding
	}

	return nil
}

func checkRecipeIngredients(new, old *models.Recipe) error {
	getIDSet := func(recipe *models.Recipe) map[int64]struct{} {
		ret := make(map[int64]struct{})
		for _, ingr := range recipe.Ingredients {
			ret[ingr.ID] = struct{}{}
		}
		for _, step := range recipe.Steps {
			for _, ingr := range step.Ingredients {
				ret[ingr.ID] = struct{}{}
			}
		}

		return ret
	}

	newSet := getIDSet(new)
	if old == nil {
		if _, ok := newSet[0]; !ok || len(newSet) != 1 {
			return apicommon.NewUserFacingError("expected all zero IDs for new recipe's ingredients")
		}
		return nil
	}

	oldSet := getIDSet(old)
	for id := range newSet {
		if _, ok := oldSet[id]; !ok {
			return apicommon.NewUserFacingError("invalid recipe ingredient ID: %d", id)
		}
	}

	return nil
}

func shouldGenerateNewEmbedding(new, old *models.Recipe) bool {
	if old == nil {
		return true
	}

	if new.Title != old.Title {
		return true
	} else if new.Description != old.Description {
		return true
	} else if len(new.Tags) != len(old.Tags) {
		return true
	}

	oldTags := make(map[string]struct{})
	for _, tag := range old.Tags {
		oldTags[tag.Name] = struct{}{}
	}

	for _, tag := range new.Tags {
		if _, ok := oldTags[tag.Name]; !ok {
			return true
		}
	}

	return false
}

func updateRecipeSteps(tx database.Transaction, new, old *models.Recipe) error {
	getIDSet := func(recipe *models.Recipe) map[int64]*models.RecipeStep {
		ret := make(map[int64]*models.RecipeStep)
		for i, step := range recipe.Steps {
			if step.ID == 0 {
				continue
			}
			ret[step.ID] = &recipe.Steps[i]
		}

		return ret
	}

	// Creating/updating new steps
	for i, step := range new.Steps {
		new.Steps[i].Index = i

		if step.RecipeID == 0 {
			new.Steps[i].RecipeID = new.ID
		} else if step.RecipeID != new.ID {
			return apicommon.NewUserFacingError("step's recipe ID does not match recipe ID")
		}

		id, err := tx.CreateOrUpdateRecipeStep(&new.Steps[i])
		if err != nil {
			return fmt.Errorf("failed to create/update recipe step: %w", err)
		}

		new.Steps[i].ID = id
	}

	if old == nil {
		return nil
	}

	newSet := getIDSet(new)
	oldSet := getIDSet(old)

	for i, step := range newSet {
		if step.ID == 0 {
			continue
		}
		if _, ok := oldSet[step.ID]; !ok {
			return apicommon.NewUserFacingError("invalid step ID %d at index %d", step.ID, i)
		}
	}

	for oldStepID, oldStep := range oldSet {
		if _, found := newSet[oldStepID]; found {
			continue
		}

		err := tx.DeleteRecipeStep(oldStep)
		if err != nil {
			return fmt.Errorf("failed to remove old recipe step: %w", err)
		}
	}

	return nil
}

func saveImages(ctx context.Context, providers config.Providers, recipe *models.Recipe) ([]string, error) {
	var savedImages []string

	if needsSaving, err := imageNeedsSaved(recipe.ImageURL); needsSaving {
		recipe.ImageURL, err = saveImage(ctx, providers, recipe.ImageURL, "title")
		if err != nil {
			return nil, fmt.Errorf("failed to save image: %w", err)
		}
		savedImages = append(savedImages, recipe.ImageURL)
	} else if err != nil {
		return nil, err
	}

	for i, step := range recipe.Steps {
		if needsSaving, err := imageNeedsSaved(step.ImageURL); needsSaving {
			recipe.Steps[i].ImageURL, err = saveImage(ctx, providers, step.ImageURL, fmt.Sprintf("step %d", i+1))
			if err != nil {
				return savedImages, fmt.Errorf("failed to save image: %w", err)
			}
			savedImages = append(savedImages, recipe.Steps[i].ImageURL)
		} else if err != nil {
			return savedImages, err
		}
	}

	return savedImages, nil
}

func rollbackSavedImages(ctx context.Context, providers config.Providers, savedImages []string) {
	for _, path := range savedImages {
		err := deleteFile(ctx, providers, path)
		if err != nil {
			providers.Log().Error("failed to rollback saved image", slog.String("object_path", path), slog.Any("error", err))
		}
	}
}

func removeOldImages(ctx context.Context, providers config.Providers, old, new *models.Recipe) {
	if old == nil {
		return
	}

	getImagePaths := func(recipe *models.Recipe) map[string]struct{} {
		ret := make(map[string]struct{}, 0)
		if recipe.ImageURL != "" {
			ret[recipe.ImageURL] = struct{}{}
		}

		for _, step := range recipe.Steps {
			if step.ImageURL != "" {
				ret[step.ImageURL] = struct{}{}
			}
		}

		return ret
	}

	oldSet := getImagePaths(old)
	newSet := getImagePaths(new)

	for oldURL := range oldSet {
		if _, found := newSet[oldURL]; !found {
			// removed image
			err := deleteFile(ctx, providers, oldURL)
			if err != nil {
				// Removing old images is not critical from the user's perspective to updating/creating a recipe.
				// Therefore, it should not block the action.
				providers.Log().Error("failed to delete old image", slog.Any("error", err))
			}
		}
	}
}

func imageNeedsSaved(path string) (bool, error) {
	parsedURL, err := url.Parse(path)
	if err != nil {
		return false, err
	}

	return parsedURL.IsAbs(), nil
}

// saveImage downloads a remote image, scales it down to at most
// maxRecipeImagePixels, and stores it in the object store as a JPEG. It returns
// the object path, which replaces the remote URL on the recipe.
func saveImage(ctx context.Context, providers config.Providers, imageURL, imgCtx string) (string, error) {
	contents, err := downloadImage(ctx, imageURL, imgCtx)
	if err != nil {
		return "", err
	}

	// DecodeConfig first so an image that is small on the wire but enormous
	// once decoded is rejected before it is allocated.
	cfg, _, err := image.DecodeConfig(bytes.NewReader(contents))
	if err != nil {
		return "", fmt.Errorf("failed to decode image header: %w", err)
	}
	if cfg.Width*cfg.Height > maxRecipeDownloadImageArea {
		return "", apicommon.NewUserFacingError("%s image exceeds the maximum size of %d megapixels, got %d", imgCtx, maxRecipeDownloadImageArea/1_000_000, cfg.Width*cfg.Height/1_000_000)
	}

	img, _, err := image.Decode(bytes.NewReader(contents))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	img = scaleImage(img)

	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, img, &jpeg.Options{Quality: recipeImageJPEGQuality}); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	objPath := fmt.Sprintf("recipes/%d/%s.jpg", providers.User().ID, uuid.Must(uuid.NewV7()))
	if err := storeFile(ctx, providers, objPath, encoded.Bytes()); err != nil {
		return "", fmt.Errorf("failed to store image: %w", err)
	}

	return objPath, nil
}

func downloadImage(ctx context.Context, imageURL, imgCtx string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get image contents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("image host responded with invalid status, expected OK. Got %s (%d)", resp.Status, resp.StatusCode)
	}

	limited := io.LimitReader(resp.Body, maxRecipeImageSize+1)
	contents, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}

	if len(contents) > maxRecipeImageSize {
		contentLength, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
		return nil, apicommon.NewUserFacingError("recipe %s image is too large. Max size is %d MiB, got %d", imgCtx, maxRecipeImageSize/apicommon.MiB, contentLength/apicommon.MiB)
	}

	return contents, nil
}

// scaleImage shrinks img so its area is at most maxRecipeImagePixels, preserving
// the aspect ratio. Images already within the limit are returned unchanged.
func scaleImage(img image.Image) image.Image {
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width*height <= maxRecipeImagePixels {
		return img
	}

	// Area scales with the square of the linear ratio.
	ratio := math.Sqrt(float64(maxRecipeImagePixels) / float64(width*height))
	scaledWidth := max(int(float64(width)*ratio), 1)
	scaledHeight := max(int(float64(height)*ratio), 1)

	scaled := image.NewRGBA(image.Rect(0, 0, scaledWidth, scaledHeight))
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), img, bounds, draw.Src, nil)
	return scaled
}

func getOrCreateIngredients(ctx context.Context, providers config.Providers, tx database.Transaction, recipe *models.Recipe) error {
	mapping := make(map[string]int64)

	for _, ingredient := range recipe.Ingredients {
		if !providers.Config().SemanticSearchEnabled {
			// Just checks that ingredient has valid ID
			_, err := tx.GetIngredientByID(providers.User().ID, ingredient.Ingredient.ID)
			if errors.Is(err, database.ErrNotFound) {
				return apicommon.NewUserFacingError("invalid ingredient ID: %d", ingredient.Ingredient.ID)
			}
			return fmt.Errorf("no semantic search recipe ingredients check not implemented")
		}

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
			providers.Log().Debug("creating ingredient", slog.String("recipe_ingredient", ingredient.Ingredient.Name), slog.String("final_name", match.Name), slog.String("ingredient_category", match.Category))

			match.ID, err = createIngredient(ctx, providers, tx, match.Name, match.Category)
			if err != nil {
				return fmt.Errorf("failed to create new ingredient: %w", err)
			}
		} else {
			providers.Log().Debug("found matching ingredient", slog.String("recipe_ingredient", ingredient.Ingredient.Name), slog.String("match_name", match.Name), slog.String("match_category", match.Category))
		}

		mapping[strings.ToLower(ingredient.Ingredient.Name)] = match.ID
	}

	for i, ingredient := range recipe.Ingredients {
		recipe.Ingredients[i].Ingredient.ID = mapping[strings.ToLower(ingredient.Ingredient.Name)]
	}

	for i, step := range recipe.Steps {
		for j, ingredient := range step.Ingredients {
			id, ok := mapping[strings.ToLower(ingredient.Ingredient.Name)]
			if !ok {
				return apicommon.NewUserFacingError("ingredient '%s' from step '%d' is not found in overall ingredients list", ingredient.Ingredient.Name, i+1)
			}
			recipe.Steps[i].Ingredients[j].Ingredient.ID = id
		}
	}

	return nil
}

func createIngredient(ctx context.Context, providers config.Providers, tx database.Transaction, name, category string) (int64, error) {
	if name == "" {
		return 0, fmt.Errorf("ingredient name may not be emtpy")
	} else if category == "" {
		return 0, fmt.Errorf("ingredient category may not be empty. Ingredient: %s", name)
	}

	// Name may be different than recipe original name, so a new embedding is required.
	embedding, err := ai.ConvertToEmbedding(ctx, providers, name)
	if err != nil {
		return 0, err
	}

	foodkeeperID, err := getFoodkeeperID(ctx, providers, tx, name, embedding)
	if err != nil {
		return 0, fmt.Errorf("failed to associate with foodkeeper: %w", err)
	}

	ingredient := &models.Ingredient{
		UserID:       providers.User().ID,
		Name:         name,
		Category:     category,
		FoodKeeperID: foodkeeperID,
		Embedding:    embedding,
	}

	return tx.CreateIngredient(ingredient)
}

func getFoodkeeperID(ctx context.Context, providers config.Providers, tx database.Transaction, name string, embedding []float32) (int64, error) {
	query := name
	for range 2 {
		candidates, err := tx.SearchFoodKeeperProductsBySemanticSimilarity(embedding, 10)
		if err != nil {
			return 0, err
		}

		selected, err := ai.AssociateFoodkeeper(ctx, providers, candidates, query)
		if err != nil {
			return 0, err
		}

		if selected.ID == 0 {
			// Suggested an alternative query, max 1 retry
			providers.Log().Debug("foodkeeper association alternative query", slog.String("ingredient_name", name), slog.String("alternative_query", selected.Name))

			query = selected.Name
			embedding, err = ai.ConvertToEmbedding(ctx, providers, query)
			if err != nil {
				return 0, err
			}
		} else {
			providers.Log().Debug("associated ingredient with foodkeeper", slog.String("ingredient_name", name), slog.String("foodkeeper_name", selected.Name))
			return selected.ID, nil
		}
	}

	return 0, nil
}
