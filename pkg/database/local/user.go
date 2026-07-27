package local

import (
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/pkg/database"
)

func (tx *localTransaction) GetUserByPasswordHash(email string, passwordHash string) (*models.User, error) {
	var found *models.User
	for _, u := range tx.data.Users {
		if u.Email == email {
			copy := u
			found = &copy
			break
		}
	}
	if found == nil {
		return nil, fmt.Errorf("user with password hash: %w", database.ErrNotFound)
	}
	if subtle.ConstantTimeCompare([]byte(found.PasswordHash), []byte(passwordHash)) != 1 {
		return nil, fmt.Errorf("user with password hash: %w", database.ErrNotFound)
	}
	return found, nil
}

func (tx *localTransaction) GetUserByID(userID int64) (*models.User, error) {
	u, ok := tx.data.Users[userID]
	if !ok {
		return nil, fmt.Errorf("user %d: %w", userID, database.ErrNotFound)
	}
	return &u, nil
}

func (tx *localTransaction) CreateUser(user *models.User) (int64, error) {
	now := time.Now()
	user.ID = tx.data.nextID()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.LastUsageReset = now
	tx.data.Users[user.ID] = *user
	return user.ID, nil
}

func (tx *localTransaction) UpdateUser(user *models.User) error {
	if _, ok := tx.data.Users[user.ID]; !ok {
		return fmt.Errorf("user %d: %w", user.ID, database.ErrNotFound)
	}
	user.UpdatedAt = time.Now()
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

	now := time.Now()
	if now.Year() != u.LastUsageReset.Year() || now.Month() != u.LastUsageReset.Month() {
		u.CurrentMonthlyRecipeExtraction = 0
		u.LastUsageReset = now
	}

	u.CurrentPhotoStorageMB += photoStorageDelta
	u.CurrentMonthlyRecipeExtraction += recipeExtractionDelta
	u.UpdatedAt = now
	tx.data.Users[userID] = u
	return nil
}
