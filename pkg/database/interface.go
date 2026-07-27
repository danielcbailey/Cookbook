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
	// CreateUser stores the user and returns its newly assigned ID.
	CreateUser(user *models.User) (int64, error)
	UpdateUser(user *models.User) error
	DeleteUser(userID int64) error
	UpdateUserUsage(userID int64, photoStorageDelta int, recipeExtractionDelta int) error

	// Recipes
	ListRecipesByUserID(userID int64) ([]*models.Recipe, error)
	SearchRecipesBySemanticSimilarity(userID int64, embedding []float32, limit int) ([]*models.Recipe, error)
	GetRecipeByID(recipeID int64) (*models.Recipe, error)
	// CreateRecipe stores the recipe and returns its newly assigned ID.
	CreateRecipe(recipe *models.Recipe) (int64, error)
	UpdateRecipe(recipe *models.Recipe) error
	DeleteRecipe(recipeID int64) error
	// CreateOrUpdateRecipeStep upserts the step on (RecipeID, Index) and returns
	// its ID: newly assigned when inserted, the existing one when updated.
	CreateOrUpdateRecipeStep(step *models.RecipeStep) (int64, error)
	DeleteRecipeStep(step *models.RecipeStep) error

	// Ingredients
	SearchIngredientsBySemanticSimilarity(userID int64, embedding []float32, limit int) ([]*models.Ingredient, error)
	GetIngredientCategories(userID int64) ([]string, error)
	GetIngredientsByCategory(userID int64, category string) ([]*models.Ingredient, error)
	// CreateIngredient stores the ingredient and returns its newly assigned ID.
	CreateIngredient(ingredient *models.Ingredient) (int64, error)
	UpdateIngredient(ingredient *models.Ingredient) error
	DeleteIngredient(ingredient *models.Ingredient) error

	// FoodKeeper
	// SearchFoodKeeperProductsBySemanticSimilarity returns the products whose
	// embeddings are nearest to the given one, closest first. FoodKeeper data is
	// global, so results are not scoped to a user. Products without an embedding
	// are skipped.
	SearchFoodKeeperProductsBySemanticSimilarity(embedding []float32, limit int) ([]*models.FoodKeeperProduct, error)
	DeleteAllFoodKeeperProducts() error
	// CreateFoodKeeperProduct stores the product under its USDA-supplied ID;
	// it does not assign a new one.
	CreateFoodKeeperProduct(product *models.FoodKeeperProduct) error
}
