package recipes

import (
	"context"
	"errors"
	"path/filepath"
	"time"

	"github.com/danielcbailey/Cookbook/core/config"
	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/pkg/objectstore"
)

const presignedURLValidDuration = 5 * time.Minute

func RecipeListingPrepareImageURLs(ctx context.Context, providers config.Providers, recipes []*models.RecipeListing) error {
	// Must create presigned URLs to the images or create a external-facing URL that the server proxies the content for
	replaceFn := func(key string) (string, error) {
		return imageURLReplace(ctx, providers, key)
	}

	for _, recipe := range recipes {
		var err error
		recipe.ImageURL, err = replaceFn(recipe.ImageURL)
		if err != nil {
			return err
		}
	}

	return nil
}

func RecipePrepareImageURLs(ctx context.Context, providers config.Providers, recipe *models.Recipe) error {
	// Must create presigned URLs to the images or create a external-facing URL that the server proxies the content for
	replaceFn := func(key string) (string, error) {
		return imageURLReplace(ctx, providers, key)
	}

	var err error
	recipe.ImageURL, err = replaceFn(recipe.ImageURL)
	if err != nil {
		return err
	}

	for i, step := range recipe.Steps {
		recipe.Steps[i].ImageURL, err = replaceFn(step.ImageURL)
		if err != nil {
			return err
		}
	}

	return nil
}

func imageURLReplace(ctx context.Context, providers config.Providers, key string) (string, error) {
	if key == "" {
		return "", nil
	}

	url, err := providers.ObjectStore().GetPresignedURL(ctx, key, presignedURLValidDuration)
	if err != nil {
		if errors.Is(err, objectstore.ErrPresignUnsupported) {
			return filepath.Join("/api/v1/images/", key), nil
		}

		return "", err
	}

	return url, nil
}
