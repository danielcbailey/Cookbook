package models

type RecipeTag struct {
	ID     int64  `json:"id"`
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
}

type RecipeTime struct {
	StepID    int64          `json:"step_id,omitempty"`
	Index     int            `json:"index"`
	StartTime float32        `json:"start_time"`
	EndTime   float32        `json:"end_time,omitempty"`
	Unit      RecipeTimeUnit `json:"unit"`
}

type RecipeIngredient struct {
	ID         int64          `json:"id"`
	StepID     int64          `json:"step_id,omitempty"`
	Index      int            `json:"index"`
	Ingredient Ingredient     `json:"ingredient"`
	Quantity   float32        `json:"quantity"`
	Unit       IngredientUnit `json:"unit"`
}

type RecipeStep struct {
	ID          int64              `json:"id"`
	RecipeID    int64              `json:"omit"`
	Index       int                `json:"index"`
	Title       string             `json:"title"`
	ImageURL    string             `json:"image_url"`
	BodyText    string             `json:"body_text"` // contains references to ingredients/times in the format of {{ingr-idx}} and {{time-idx}}
	Ingredients []RecipeIngredient `json:"ingredients"`
	Times       []RecipeTime       `json:"times"`
}

type Recipe struct {
	ID          int64              `json:"id"`
	UserID      int64              `json:"user_id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	ImageURL    string             `json:"image_url"`
	Author      string             `json:"author"`
	Publisher   string             `json:"publisher"`
	Servings    int                `json:"servings"`
	Ingredients []RecipeIngredient `json:"ingredients"`
	Steps       []RecipeStep       `json:"steps"`

	Category      string      `json:"category"`
	Protein       string      `json:"protein"`
	Cuisine       string      `json:"cuisine"`
	SuggestedMeal string      `json:"suggested_meal"`
	Tags          []RecipeTag `json:"tags"`
	Embedding     []float32   `json:"embedding,omitempty"`
	CreatedAt     int64       `json:"created_at"`
	UpdatedAt     int64       `json:"updated_at"`
}

type RecipeTimeUnit string

const (
	RecipeTimeUnitSeconds RecipeTimeUnit = "second"
	RecipeTimeUnitMinutes RecipeTimeUnit = "minute"
	RecipeTimeUnitHours   RecipeTimeUnit = "hour"
	RecipeTimeUnitDays    RecipeTimeUnit = "day"
	RecipeTimeUnitWeeks   RecipeTimeUnit = "week"
)
