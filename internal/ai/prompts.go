package ai

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/danielcbailey/Cookbook/core/models"
)

// promptMarker matches a {{name}} placeholder in an embedded prompt file.
var promptMarker = regexp.MustCompile(`{{(\w+)}}`)

// promptValues is every marker a prompt file may use, each expanding to the
// comma-separated list of values the model is allowed to answer with. They are
// derived from the generated enum slices, so adding a constant in models is all
// it takes for the prompts and the validation to stay in agreement.
var promptValues = map[string]string{
	"ingredientsTypes": joinEnum(models.IngredientCategoryValues),
	"ingredientUnits":  joinEnum(models.IngredientUnitValues),
	"recipeTimeUnits":  joinEnum(models.RecipeTimeUnitValues),
	"suggestedMeals":   joinEnum(models.RecipeMealtimeValues),
	"recipeCategories": joinEnum(models.RecipeCategoryValues),
	"recipeCuisines":   joinEnum(models.RecipeCuisineValues),
	"recipeProteins":   joinEnum(models.RecipeProteinValues),
}

func joinEnum[T ~string](values []T) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = string(v)
	}
	return strings.Join(parts, ", ")
}

// renderPrompt expands every {{marker}} in an embedded prompt. Prompts are
// rendered once at startup, so an unknown marker panics rather than reaching
// the model verbatim: a typo in a prompt file is a bug that should surface on
// the first run, not as a confused answer in production.
func renderPrompt(prompt string) string {
	var unknown []string

	rendered := promptMarker.ReplaceAllStringFunc(prompt, func(match string) string {
		name := promptMarker.FindStringSubmatch(match)[1]
		value, ok := promptValues[name]
		if !ok {
			unknown = append(unknown, name)
			return match
		}
		return value
	})

	if len(unknown) > 0 {
		panic(fmt.Sprintf("ai: prompt uses unknown template marker(s): %s", strings.Join(unknown, ", ")))
	}

	return rendered
}
