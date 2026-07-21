package local

import (
	"fmt"

	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/pkg/database"
)

func (tx *localTransaction) GetUserByPasswordHash(passwordHash string) (*models.User, error) {
	for _, u := range tx.data.Users {
		if u.PasswordHash == passwordHash {
			copy := u
			return &copy, nil
		}
	}
	return nil, fmt.Errorf("user with password hash: %w", database.ErrNotFound)
}

func (tx *localTransaction) GetUserByID(userID int64) (*models.User, error) {
	u, ok := tx.data.Users[userID]
	if !ok {
		return nil, fmt.Errorf("user %d: %w", userID, database.ErrNotFound)
	}
	return &u, nil
}

func (tx *localTransaction) CreateUser(user *models.User) error {
	user.ID = tx.data.nextID()
	tx.data.Users[user.ID] = *user
	return nil
}

func (tx *localTransaction) UpdateUser(user *models.User) error {
	if _, ok := tx.data.Users[user.ID]; !ok {
		return fmt.Errorf("user %d: %w", user.ID, database.ErrNotFound)
	}
	tx.data.Users[user.ID] = *user
	return nil
}

func (tx *localTransaction) DeleteUser(userID int64) error {
	delete(tx.data.Users, userID)
	for id, r := range tx.data.Recipes {
		if r.UserID == userID {
			delete(tx.data.Recipes, id)
		}
	}
	return nil
}

func (tx *localTransaction) UpdateUserUsage(userID int64, photoStorageDelta int, recipeExtractionDelta int) error {
	u, ok := tx.data.Users[userID]
	if !ok {
		return fmt.Errorf("user %d: %w", userID, database.ErrNotFound)
	}
	u.CurrentPhotoStorageMB += photoStorageDelta
	u.CurrentMonthlyRecipeExtraction += recipeExtractionDelta
	tx.data.Users[userID] = u
	return nil
}
