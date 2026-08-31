package recipes

import (
	"context"
	"log/slog"

	"github.com/danielcbailey/Cookbook/core/apicommon"
	"github.com/danielcbailey/Cookbook/core/config"
)

func storeFile(ctx context.Context, providers config.Providers, path string, contents []byte) error {
	tx, err := providers.DB().NewTransaction(ctx)
	if err != nil {
		return err
	}
	defer tx.Commit()

	// Checking limits
	user, err := tx.GetUserByID(providers.User().ID)
	if err != nil {
		return err
	}

	// The limit is configured in whole megabytes while usage is tracked in bytes.
	if user.CurrentPhotoStorageBytes+int64(len(contents)) > int64(user.MaxPhotoStorageMB)*apicommon.MiB {
		return apicommon.NewUserFacingError("file size exceeds the maximum allowed storage limit")
	}

	err = providers.ObjectStore().StoreFile(ctx, path, contents)
	if err != nil {
		return err
	}

	// Updating user limits
	err = tx.UpdateUserUsage(providers.User().ID, int64(len(contents)), 0)
	if err != nil {
		// Give it away "for free" since it is an internal failure and it already stored the file
		providers.Log().Error("failed to update user usage after storing file", slog.String("object_path", path), slog.Any("error", err))
	}

	return nil
}

func deleteFile(ctx context.Context, providers config.Providers, path string) error {
	tx, err := providers.DB().NewTransaction(ctx)
	if err != nil {
		return err
	}
	defer tx.Commit()

	size, err := providers.ObjectStore().DeleteFile(ctx, path)
	if err != nil {
		return err
	}

	err = tx.UpdateUserUsage(providers.User().ID, -size, 0)
	if err != nil {
		// Won't return an error because the file is already deleted, so a retry would fail on the deletion step
		providers.Log().Error("failed to update user usage after deleting file", slog.String("object_path", path), slog.Any("error", err))
	}

	return nil
}
