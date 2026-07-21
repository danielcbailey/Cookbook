package models

type Ingredient struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Name         string    `json:"name"`
	FoodKeeperID int64     `json:"food_keeper_id,omitempty"`
	Embedding    []float32 `json:"embedding,omitempty"`
}

type PantryItem struct {
	ID              int64              `json:"id"`
	UserID          int64              `json:"user_id"`
	Ingredient      Ingredient         `json:"ingredient"`
	Quantity        float32            `json:"quantity"`
	Unit            IngredientUnit     `json:"unit"`
	PurchaseDate    int64              `json:"purchase_date"` // Unix timestamp
	OpeningDate     int64              `json:"opening_date"`
	ExpirationDate  int64              `json:"expiration_date"`
	CurrentLocation PantryItemLocation `json:"current_location"`
}

type PantryItemLocation string

const (
	PantryItemLocationFridge  PantryItemLocation = "fridge"
	PantryItemLocationFreezer PantryItemLocation = "freezer"
	PantryItemLocationPantry  PantryItemLocation = "pantry"
)

type IngredientUnit string

const (
	IngredientUnitGrams       IngredientUnit = "gram"
	IngredientUnitKilograms   IngredientUnit = "kilogram"
	IngredientUnitMilligrams  IngredientUnit = "milligram"
	IngredientUnitOunces      IngredientUnit = "ounce"
	IngredientUnitPounds      IngredientUnit = "pound"
	IngredientUnitMilliliters IngredientUnit = "milliliter"
	IngredientUnitLiters      IngredientUnit = "liter"
	IngredientUnitCups        IngredientUnit = "cup"
	IngredientUnitTablespoons IngredientUnit = "us-tbsp"
	IngredientUnitTeaspoons   IngredientUnit = "us-tsp"
	IngredientUnitFluidOunces IngredientUnit = "us-fl-oz"
	IngredientUnitGallons     IngredientUnit = "us-gal"
	IngredientUnitQuarts      IngredientUnit = "us-qt"
	IngredientUnitPieces      IngredientUnit = "piece"
	IngredientUnitPinches     IngredientUnit = "pinch"
	IngredientUnitContainers  IngredientUnit = "container"
)
