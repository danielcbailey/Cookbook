// --------------------------------------------------------------
// Recipe API Types
// --------------------------------------------------------------

export const RecipeTimeUnit = {
    Seconds: 'second',
    Minutes: 'minute',
    Hours: 'hour',
    Days: 'day',
    Weeks: 'week',
} as const;

export type RecipeTimeUnit = (typeof RecipeTimeUnit)[keyof typeof RecipeTimeUnit];

export const IngredientUnit = {
    Grams: 'gram',
    Kilograms: 'kilogram',
    Milligrams: 'milligram',
    Ounces: 'ounce',
    Pounds: 'pound',
    Milliliters: 'milliliter',
    Liters: 'liter',
    Cups: 'cup',
    Tablespoons: 'us-tbsp',
    Teaspoons: 'us-tsp',
    FluidOunces: 'us-fl-oz',
    Gallons: 'us-gal',
    Quarts: 'us-qt',
    Pieces: 'piece',
    Pinches: 'pinch',
    Containers: 'container',
    Unspecified: 'unspecified',
    Unknown: '',
} as const;

export type IngredientUnit = (typeof IngredientUnit)[keyof typeof IngredientUnit];

export type Ingredient = {
    id: number;
    user_id: number;
    name: string;
    category: string;
    food_keeper_id?: number;
    embedding?: number[];
}

export type RecipeTag = {
    id: number;
    user_id: number;
    name: string;
}

export type RecipeTime = {
    step_id?: number;
    index: number;
    start_time: number;
    end_time?: number;
    unit: RecipeTimeUnit;
}

export type RecipeIngredient = {
    id: number;
    step_id?: number;
    index: number;
    ingredient: Ingredient;
    quantity: number;
    unit: IngredientUnit;
}

export type RecipeStep = {
    id: number;
    index: number;
    title: string;
    image_url: string;
    /** Contains references to ingredients/times in the format {{ingr-idx}} and {{time-idx}}. */
    body_text: string;
    ingredients: RecipeIngredient[];
    times: RecipeTime[];
}

export type RecipeNutrition = {
    servings: number;
    serving_mass: number;
    serving_mass_unit: IngredientUnit;
    is_auto_generated: boolean;
    calories: number;
    total_fat_grams: number;
    saturated_fat_grams: number;
    trans_fat_grams: number;
    cholestrol_mg: number;
    sodium_mg: number;
    total_carbs_grams: number;
    dietary_fiber_grams: number;
    total_sugar_grams: number;
    protein_grams: number;
}

export type Recipe = {
    id: number;
    user_id: number;
    title: string;
    description: string;
    image_url: string;
    author: string;
    publisher: string;
    servings: number;
    ingredients: RecipeIngredient[];
    steps: RecipeStep[];
    total_time: RecipeTime;
    nutrition: RecipeNutrition;

    category: string;
    protein: string;
    cuisine: string;
    suggested_meal: string;
    tags: RecipeTag[];
    embedding?: number[];
    created_at: number;
    updated_at: number;
}

export type RecipeListItem = {
    id: number;
    user_id: number;
    title: string;
    image_url: string;
    servings: number;
    calories: number;
    time: RecipeTime;
    tags: RecipeTag[];
}

// --------------------------------------------------------------
// User API Types
// --------------------------------------------------------------

export type User = {
    id: number;
    email: string;
    first_name: string;
    last_name: string;
    
    recipe_limit: number;
    max_photo_storage_mb: number;
    max_monthly_recipe_extraction: number;

    current_photo_storage_bytes: number;
    current_monthly_recipe_extraction: number;

    created_at: string;
    updated_at: string;
}
