package database

import (
	"context"
	"errors"

	"github.com/danielcbailey/Cookbook/core/models"
)

var ErrNotFound = errors.New("not found")

type Database interface {
	NewTransaction(ctx context.Context) (Transaction, error)
}

type Transaction interface {
	Commit() error
	Rollback() error

	// Users
	GetUserByPasswordHash(email string, passwordHash string) (*models.User, error)
	GetUserByID(userID int64) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(user *models.User) error
	DeleteUser(userID int64) error
	UpdateUserUsage(userID int64, photoStorageDelta int, recipeExtractionDelta int) error

	// Recipes
	ListRecipesByUserID(userID int64) ([]*models.Recipe, error)
	SearchRecipesBySemanticSimilarity(userID int64, embedding []float32, limit int) ([]*models.Recipe, error)
	GetRecipeByID(recipeID int64) (*models.Recipe, error)
	CreateRecipe(recipe *models.Recipe) error
	UpdateRecipe(recipe *models.Recipe) error
	DeleteRecipe(recipeID int64) error
	CreateOrUpdateRecipeStep(step *models.RecipeStep) error
	DeleteRecipeStep(step *models.RecipeStep) error
}
