package ai

import (
	"strings"
	"testing"

	"github.com/danielcbailey/Cookbook/core/models"
)

// The rendered prompts are package-level vars, so this also covers the markers
// in the embedded files being known: an unknown one panics during init.
func TestRenderedPromptsHaveNoMarkersLeft(t *testing.T) {
	for name, prompt := range map[string]string{
		"ingredientAssociation": ingredientAssociationPrompt,
		"ingredientCreation":    ingredientCreationPrompt,
		"recipeExtraction":      recipeExtractionPrompt,
	} {
		if m := promptMarker.FindString(prompt); m != "" {
			t.Errorf("%s prompt still contains %s after rendering", name, m)
		}
	}
}

func TestRenderedPromptsListEveryEnumValue(t *testing.T) {
	// Each prompt must offer the model every value the models package accepts,
	// otherwise extraction can never produce some of them.
	cases := []struct {
		prompt string
		name   string
		values []string
	}{
		{ingredientCreationPrompt, "IngredientCategory", asStrings(models.IngredientCategoryValues)},
		{recipeExtractionPrompt, "IngredientUnit", asStrings(models.IngredientUnitValues)},
		{recipeExtractionPrompt, "RecipeTimeUnit", asStrings(models.RecipeTimeUnitValues)},
		{recipeExtractionPrompt, "RecipeMealtime", asStrings(models.RecipeMealtimeValues)},
		{recipeExtractionPrompt, "RecipeCategory", asStrings(models.RecipeCategoryValues)},
		{recipeExtractionPrompt, "RecipeCuisine", asStrings(models.RecipeCuisineValues)},
		{recipeExtractionPrompt, "RecipeProtein", asStrings(models.RecipeProteinValues)},
	}

	for _, c := range cases {
		if len(c.values) == 0 {
			t.Errorf("%s has no values", c.name)
			continue
		}
		for _, v := range c.values {
			if !strings.Contains(c.prompt, v) {
				t.Errorf("%s value %q missing from rendered prompt", c.name, v)
			}
		}
	}
}

func TestRenderPromptPanicsOnUnknownMarker(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected a panic for an unknown marker")
		}
	}()
	renderPrompt("acceptable values are:\n{{notAMarker}}")
}

func asStrings[T ~string](values []T) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = string(v)
	}
	return out
}
