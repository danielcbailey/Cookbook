package recipes

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/danielcbailey/Cookbook/core/apicommon"
	"github.com/danielcbailey/Cookbook/core/config"
	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/internal/ai"
	"github.com/danielcbailey/Cookbook/pkg/database"
)

const maxScrapeSize = 1 * apicommon.MiB
const maxPhotoArea = 10_000_000 // pixels
const maxPhotos = 3

func ScrapeRecipeWeb(ctx context.Context, providers config.Providers, url string) (*models.Recipe, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, apicommon.NewUserFacingError("failed to get site contents: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		return nil, apicommon.NewUserFacingError("website responded with invalid status, expected OK. Got %s (%d)", resp.Status, resp.StatusCode)
	} else if cType := resp.Header.Get("Content-Type"); strings.SplitN(cType, ";", 2)[0] != "text/html" {
		return nil, apicommon.NewUserFacingError("invalid website content type, expected text/html. Got %s", cType)
	}

	limited := io.LimitReader(resp.Body, maxScrapeSize+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}

	if len(body) > maxScrapeSize {
		return nil, apicommon.NewUserFacingError("website content is too long. Max size is %d KiB", maxScrapeSize/apicommon.KiB)
	}

	recipe, err := ai.ExtractRecipeFromHTML(ctx, providers, string(body))
	if err != nil {
		return nil, err
	}

	err = matchIngredients(ctx, providers, recipe)
	return recipe, err
}

func ScrapeRecipePhotos(ctx context.Context, providers config.Providers, files []ai.FileAttachment) (*models.Recipe, error) {
	if len(files) > maxPhotos {
		return nil, apicommon.NewUserFacingError("too many photos, maximum of %d allowed", maxPhotos)
	}

	for i, file := range files {
		if file.MimeType != "image/jpeg" && file.MimeType != "image/png" {
			return nil, apicommon.NewUserFacingError("photo %d has unsupported type %s, only JPEG and PNG are allowed", i+1, file.MimeType)
		}

		cfg, _, err := image.DecodeConfig(bytes.NewReader(file.Content))
		if err != nil {
			return nil, apicommon.NewUserFacingError("photo %d is not a valid image: %v", i+1, err)
		}

		if cfg.Width*cfg.Height > maxPhotoArea {
			return nil, apicommon.NewUserFacingError("photo %d exceeds the maximum image size of %d megapixels", i+1, maxPhotoArea/1_000_000)
		}
	}

	recipe, err := ai.ExtractRecipeFromPhotos(ctx, providers, files)
	if err != nil {
		return nil, err
	}

	err = matchIngredients(ctx, providers, recipe)
	return recipe, err
}

// matchIngredients attempts to match all ingredients to existing ingredients in the database
func matchIngredients(ctx context.Context, providers config.Providers, recipe *models.Recipe) error {
	if !providers.Config().SemanticSearchEnabled {
		return nil
	}

	tx, err := providers.DB().NewTransaction(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, step := range recipe.Steps {
		ingredients := models.GetRecipeStepIngredients(step)
		for _, ingr := range ingredients {
			match, err := matchIngredientByName(ctx, providers, tx, ingr.Ingredient.Name, recipe.Title)
			if err != nil {
				return fmt.Errorf("failed to match ingredient %q: %w", ingr.Ingredient.Name, err)
			} else if match != nil && match.ID != 0 {
				providers.Log().Debug("found matching ingredient", slog.String("recipe_ingredient", ingr.Ingredient.Name), slog.String("match_name", match.Name), slog.String("match_category", match.Category))
			}

			matchedIngr := ingr
			matchedIngr.Ingredient = *match
			ingr.Replace(&recipe.Steps[i], matchedIngr)
		}
	}

	return nil
}

func matchIngredientByName(ctx context.Context, providers config.Providers, tx database.Transaction, name, recipeTitle string) (*models.Ingredient, error) {
	embedding, err := ai.ConvertToEmbedding(ctx, providers, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding of ingredient name: %w", err)
	}

	candidates, err := tx.SearchIngredientsBySemanticSimilarity(providers.User().ID, embedding, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to search existing ingredients: %w", err)
	}

	return ai.AssociateIngredient(ctx, providers, candidates, recipeTitle, name)
}
