package recipes

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/danielcbailey/Cookbook/core/apicommon"
	"github.com/danielcbailey/Cookbook/core/config"
)

func DeleteRecipe(ctx context.Context, providers config.Providers, recipeID int64) error {
	tx, err := providers.DB().NewTransaction(ctx)
	if err != nil {
		return err
	}

	existing, err := tx.GetRecipeByID(recipeID)
	if err != nil {
		return err
	}

	if existing.UserID != providers.User().ID {
		return apicommon.NewUserFacingError("invalid recipe ID")
	}

	// Deleting recipe steps
	for _, step := range existing.Steps {
		err := tx.DeleteRecipeStep(&step)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.DeleteRecipe(recipeID)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit recipe deletion: %w", err)
	}

	// Deleting images - failed deletions won't block user's action
	if existing.ImageURL != "" {
		err := providers.ObjectStore().DeleteFile(ctx, existing.ImageURL)
		if err != nil {
			providers.Log().Error("failed to delete image", slog.String("object_path", existing.ImageURL), slog.Any("error", err))
		}
	}

	for _, step := range existing.Steps {
		if step.ImageURL == "" {
			continue
		}

		err := providers.ObjectStore().DeleteFile(ctx, step.ImageURL)
		if err != nil {
			providers.Log().Error("failed to delete image", slog.String("object_path", existing.ImageURL), slog.Any("error", err))
		}
	}

	return nil
}
