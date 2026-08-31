package models

import "time"

type User struct {
	// PII
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	PasswordHash string `json:"-"`

	// Limits
	RecipeLimit                int `json:"recipe_limit"`
	MaxPhotoStorageMB          int `json:"max_photo_storage_mb"`
	MaxMonthlyRecipeExtraction int `json:"max_monthly_recipe_extraction"`

	// Usage
	CurrentPhotoStorageBytes       int64 `json:"current_photo_storage_bytes"`
	CurrentMonthlyRecipeExtraction int   `json:"current_monthly_recipe_extraction"`

	// Timestamps
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastUsageReset time.Time `json:"-"`
}
