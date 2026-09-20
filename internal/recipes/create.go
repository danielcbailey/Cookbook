package recipes

import (
	"bytes"
	"context"
	"encoding/base64"
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
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/danielcbailey/Cookbook/core/apicommon"
	"github.com/danielcbailey/Cookbook/core/config"
	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/internal/ai"
	"github.com/danielcbailey/Cookbook/internal/pantry"
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
	savedImages, err := saveImages(ctx, providers, recipe, existing)
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

func saveImages(ctx context.Context, providers config.Providers, recipe *models.Recipe, existing *models.Recipe) ([]string, error) {
	var savedImages []string

	if needsSaving, err := imageNeedsSaved(providers, recipe.ImageURL); needsSaving {
		recipe.ImageURL, err = saveImage(ctx, providers, recipe.ImageURL, "title")
		if err != nil {
			return nil, fmt.Errorf("failed to save image: %w", err)
		}
		savedImages = append(savedImages, recipe.ImageURL)
	} else if err != nil {
		return nil, err
	} else {
		recipe.ImageURL = getOriginalCoverImageURI(existing)
	}

	for i, step := range recipe.Steps {
		if needsSaving, err := imageNeedsSaved(providers, step.ImageURL); needsSaving {
			recipe.Steps[i].ImageURL, err = saveImage(ctx, providers, step.ImageURL, fmt.Sprintf("step %d", i+1))
			if err != nil {
				return savedImages, fmt.Errorf("failed to save image: %w", err)
			}
			savedImages = append(savedImages, recipe.Steps[i].ImageURL)
		} else if err != nil {
			return savedImages, err
		} else {
			recipe.Steps[i].ImageURL = getOriginalStepImageURI(existing, step.ID)
		}
	}

	return savedImages, nil
}

func getOriginalCoverImageURI(existing *models.Recipe) string {
	if existing != nil && existing.ImageURL != "" {
		return existing.ImageURL
	}
	return ""
}

func getOriginalStepImageURI(existing *models.Recipe, stepID int64) string {
	if existing == nil {
		return ""
	}

	for _, s := range existing.Steps {
		if s.ID == stepID {
			return s.ImageURL
		}
	}

	return ""
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

func imageNeedsSaved(providers config.Providers, path string) (bool, error) {
	return strings.TrimSpace(path) != "" && !providers.ObjectStore().HostsURL(path), nil
}

// saveImage downloads a remote image, scales it down to at most
// maxRecipeImagePixels, and stores it in the object store as a JPEG. It returns
// the object path, which replaces the remote URL on the recipe.
func saveImage(ctx context.Context, providers config.Providers, imageURL, imgCtx string) (string, error) {
	contents, isEmbedded, err := decodeEmbeddedImage(imageURL)
	if err != nil {
		return "", fmt.Errorf("failed to decode embedded image: %w", err)
	} else if !isEmbedded {
		contents, err = downloadImage(ctx, imageURL, imgCtx)
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

func decodeEmbeddedImage(urlStr string) (data []byte, isEmbedded bool, err error) {
	if !strings.HasPrefix(urlStr, "data:image/") {
		// The URL may have been percent-encoded in transit. PathUnescape is used
		// rather than QueryUnescape because the latter also turns '+' into a
		// space, which corrupts the base64 payload of an already-decoded data URL.
		unescaped, unescapeErr := url.PathUnescape(urlStr)
		if unescapeErr != nil || !strings.HasPrefix(unescaped, "data:image/") {
			return nil, false, nil
		}

		urlStr = unescaped
	}

	semiIndex := strings.Index(urlStr, ";")
	commaIndex := strings.Index(urlStr, ",")
	if semiIndex == -1 || commaIndex == -1 || semiIndex > commaIndex {
		return nil, false, fmt.Errorf("invalid embedded image URL")
	}

	encoding := urlStr[semiIndex+1 : commaIndex]
	if encoding != "base64" {
		return nil, false, fmt.Errorf("unsupported embedded image encoding: %s", encoding)
	}

	data, err = base64.StdEncoding.DecodeString(urlStr[commaIndex+1:])
	isEmbedded = true
	return
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
		return nil, apicommon.NewUserFacingError("failed to get image contents: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, apicommon.NewUserFacingError("image host responded with invalid status, expected OK. Got %s (%d)", resp.Status, resp.StatusCode)
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

func consolidateRecipeIngredients(providers config.Providers, recipe *models.Recipe) {
	distinctIngredients := make(map[string]models.RecipeIngredient)
	extras := make([]models.RecipeIngredient, 0)

	for _, ingr := range recipe.Ingredients {
		existing, ok := distinctIngredients[ingr.Ingredient.Name]
		if !ok {
			existing = ingr
		} else {
			var err error
			existing, err = pantry.AddRecipeIngredients(existing, ingr)
			if err != nil {
				providers.Log().Info("could not add ingredients",
					slog.String("ingredient_name", ingr.Ingredient.Name),
					slog.String("acc_unit", string(existing.Unit)),
					slog.String("add_unit", string(ingr.Unit)))
				extras = append(extras, ingr)
				continue
			}
		}

		distinctIngredients[ingr.Ingredient.Name] = existing
	}

	// Converting to slice
	recipe.Ingredients = make([]models.RecipeIngredient, 0, len(distinctIngredients)+len(extras))

	for _, ingr := range distinctIngredients {
		recipe.Ingredients = append(recipe.Ingredients, ingr)
	}

	for _, ingr := range extras {
		recipe.Ingredients = append(recipe.Ingredients, ingr)
	}

	// Sorting by quantity, name as a fallback
	slices.SortFunc(recipe.Ingredients, func(a, b models.RecipeIngredient) int {
		aG, aErr := pantry.ConvertIngredientUnit(a.Ingredient.Density, a.Quantity, a.Unit, models.IngredientUnitGrams)
		bG, bErr := pantry.ConvertIngredientUnit(b.Ingredient.Density, b.Quantity, b.Unit, models.IngredientUnitGrams)

		diff := func(x float32) int {
			if x < 0 {
				return int(math.Floor(float64(x)))
			}
			return int(math.Ceil(float64(x)))
		}

		switch {
		case aErr == nil && bErr == nil:
			return diff(bG - aG)
		case aErr == nil:
			return -1
		case bErr == nil:
			return 1
		default:
			return strings.Compare(a.Ingredient.Name, b.Ingredient.Name)
		}
	})
}

func getOrCreateIngredients(ctx context.Context, providers config.Providers, tx database.Transaction, recipe *models.Recipe) error {
	recipe.Ingredients = []models.RecipeIngredient{}
	for _, step := range recipe.Steps {
		ingredients := models.GetRecipeStepIngredients(step)
		recipe.Ingredients = append(recipe.Ingredients, ingredients...)
	}

	mapping := make(map[string]int64)

	for _, ingredient := range recipe.Ingredients {
		match, err := tx.GetIngredientByName(providers.User().ID, ingredient.Ingredient.Name)
		if err != nil && !errors.Is(err, database.ErrNotFound) {
			return err
		}

		var id int64
		if match == nil || match.ID == 0 {
			// Must create the ingredient
			providers.Log().Debug("creating ingredient", slog.String("recipe_ingredient", ingredient.Ingredient.Name))

			id, err = createIngredient(ctx, providers, tx, ingredient.Ingredient.Name)
			if err != nil {
				return fmt.Errorf("failed to create new ingredient: %w", err)
			}
		} else {
			id = match.ID
		}

		mapping[strings.ToLower(ingredient.Ingredient.Name)] = id
	}

	for i, ingredient := range recipe.Ingredients {
		recipe.Ingredients[i].Ingredient.ID = mapping[strings.ToLower(ingredient.Ingredient.Name)]
	}

	consolidateRecipeIngredients(providers, recipe)

	return nil
}

func createIngredient(ctx context.Context, providers config.Providers, tx database.Transaction, name string) (int64, error) {
	if name == "" {
		return 0, fmt.Errorf("ingredient name may not be emtpy")
	}

	// Name may be different than recipe original name, so a new embedding is required.
	embedding, err := ai.ConvertToEmbedding(ctx, providers, name)
	if err != nil {
		return 0, err
	}

	foodkeeperID, density, category, err := getFoodkeeperID(ctx, providers, tx, name, embedding)
	if err != nil {
		providers.Log().Warn("failed to associate with foodkeeper", slog.Any("error", err))
	}

	ingredient := &models.Ingredient{
		UserID:       providers.User().ID,
		Name:         name,
		Category:     category,
		Density:      density,
		FoodKeeperID: foodkeeperID,
		Embedding:    embedding,
	}

	return tx.CreateIngredient(ingredient)
}

func getFoodkeeperID(ctx context.Context, providers config.Providers, tx database.Transaction, name string, embedding []float32) (int64, float64, string, error) {
	query := name
	var density float64
	var category string
	for range 2 {
		candidates, err := tx.SearchFoodKeeperProductsBySemanticSimilarity(embedding, 10)
		if err != nil {
			return 0, 0, "", err
		}

		var selected *models.FoodKeeperProduct
		selected, density, category, err = ai.AssociateNewIngredient(ctx, providers, candidates, query)
		if err != nil {
			return 0, density, category, err
		}

		if selected.ID == 0 {
			// Suggested an alternative query, max 1 retry
			providers.Log().Debug("foodkeeper association alternative query", slog.String("ingredient_name", name), slog.String("alternative_query", selected.Name))

			query = selected.Name
			embedding, err = ai.ConvertToEmbedding(ctx, providers, query)
			if err != nil {
				return 0, density, category, err
			}
		} else {
			providers.Log().Debug("associated ingredient with foodkeeper", slog.String("ingredient_name", name), slog.String("foodkeeper_name", selected.Name))
			return selected.ID, density, category, nil
		}
	}

	return 0, density, category, nil
}
