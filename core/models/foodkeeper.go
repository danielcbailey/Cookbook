package models

// FoodKeeperDuration is one storage-duration recommendation from the USDA
// FoodKeeper data. Min/Max are nil when the source data has no value; Metric
// is stored verbatim from the source (e.g. "Days", "Months", "Indefinitely").
type FoodKeeperDuration struct {
	Min    *float64 `json:"min,omitempty"`
	Max    *float64 `json:"max,omitempty"`
	Metric string   `json:"metric,omitempty"`
	Tips   string   `json:"tips,omitempty"`
}

// FoodKeeperProduct is one product entry from the USDA FoodKeeper data.
// The ID is the USDA-supplied product ID, not a generated one.
// "DOP" prefixed durations are measured from the date of purchase.
type FoodKeeperProduct struct {
	ID              int64  `json:"id"`
	CategoryID      int64  `json:"category_id"`
	CategoryName    string `json:"category_name"`
	SubcategoryName string `json:"subcategory_name,omitempty"`
	Name            string `json:"name"`
	NameSubtitle    string `json:"name_subtitle,omitempty"`
	Keywords        string `json:"keywords,omitempty"`

	Pantry                  FoodKeeperDuration `json:"pantry"`
	DOPPantry               FoodKeeperDuration `json:"dop_pantry"`
	PantryAfterOpening      FoodKeeperDuration `json:"pantry_after_opening"`
	Refrigerate             FoodKeeperDuration `json:"refrigerate"`
	DOPRefrigerate          FoodKeeperDuration `json:"dop_refrigerate"`
	RefrigerateAfterOpening FoodKeeperDuration `json:"refrigerate_after_opening"`
	RefrigerateAfterThawing FoodKeeperDuration `json:"refrigerate_after_thawing"`
	Freeze                  FoodKeeperDuration `json:"freeze"`
	DOPFreeze               FoodKeeperDuration `json:"dop_freeze"`

	Embedding []float32 `json:"embedding,omitempty"`
}
