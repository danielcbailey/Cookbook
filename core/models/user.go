package models

type User struct {
	// PII
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	PasswordHash string `json:"password_hash"`

	// Limits
	RecipeLimit                int `json:"recipe_limit"`
	MaxPhotoStorageMB          int `json:"max_photo_storage_mb"`
	MaxMonthlyRecipeExtraction int `json:"max_monthly_recipe_extraction"`

	// Usage
	CurrentPhotoStorageMB          int `json:"current_photo_storage_mb"`
	CurrentMonthlyRecipeExtraction int `json:"current_monthly_recipe_extraction"`
}
