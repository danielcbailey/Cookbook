package ai

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/danielcbailey/Cookbook/core/models"
)

func parseRecipeString(t *testing.T, doc string) *models.Recipe {
	t.Helper()
	r, err := parseRecipe(xml.NewDecoder(strings.NewReader(doc)))
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestParseRecipe_TotalTimeAndNutrition(t *testing.T) {
	r := parseRecipeString(t, `<recipe>
		<title>Brownies</title>
		<meta servings="10" author="John Doe" source="Johns Baking Blog"/>
		<time rangeStart="45" rangeEnd="50" unit="minute"/>
		<nutrition serving_size="1" serving_mass_grams="100">
			<nutrient type="calories" value="500"/>
			<nutrient type="total_fat_grams" value="10"/>
			<nutrient type="saturated_fat_grams" value="4"/>
			<nutrient type="trans_fat_grams" value="0.5"/>
			<nutrient type="cholestrol_mg" value="30"/>
			<nutrient type="sodium_mg" value="640"/>
			<nutrient type="total_carbs_grams" value="78.5"/>
			<nutrient type="dietary_fiber_grams" value="6"/>
			<nutrient type="total_sugar_grams" value="4.5"/>
			<nutrient type="protein_grams" value="18.25"/>
		</nutrition>
	</recipe>`)

	wantTime := models.RecipeTime{StartTime: 45, EndTime: 50, Unit: models.RecipeTimeUnitMinutes}
	if r.TotalTime != wantTime {
		t.Fatalf("total time: got %+v, want %+v", r.TotalTime, wantTime)
	}

	want := models.RecipeNutrition{
		NutritionServings:    1,
		NutritionServingMass: 100,
		NutritionServingUnit: models.IngredientUnitGrams,
		Calories:             500,
		TotalFatGrams:        10,
		SaturatedFatGrams:    4,
		TransFatGrams:        0.5,
		CholestrolMilligrams: 30,
		SodiumMilligrams:     640,
		TotalCarbsGrams:      78.5,
		DietaryFiberGrams:    6,
		TotalSugarGrams:      4.5,
		ProteinGrams:         18.25,
	}
	if r.Nutrition != want {
		t.Fatalf("nutrition: got %+v, want %+v", r.Nutrition, want)
	}
}

func TestParseRecipe_TotalTimeExact(t *testing.T) {
	r := parseRecipeString(t, `<recipe><time exact="30" unit="hour"/></recipe>`)

	want := models.RecipeTime{StartTime: 30, Unit: models.RecipeTimeUnitHours}
	if r.TotalTime != want {
		t.Fatalf("total time: got %+v, want %+v", r.TotalTime, want)
	}
}

// Step times are consumed by parseStep, so they must not overwrite the
// recipe-level total time.
func TestParseRecipe_StepTimesDoNotOverwriteTotalTime(t *testing.T) {
	r := parseRecipeString(t, `<recipe>
		<time rangeStart="45" rangeEnd="50" unit="minute"/>
		<steps>
			<step title="Melt">
				Melt the chocolate for <time exact="3" unit="minute"/>.
			</step>
			<step title="Bake">
				Bake for <time rangeStart="30" rangeEnd="35" unit="minute"/>.
			</step>
		</steps>
	</recipe>`)

	wantTotal := models.RecipeTime{StartTime: 45, EndTime: 50, Unit: models.RecipeTimeUnitMinutes}
	if r.TotalTime != wantTotal {
		t.Fatalf("total time: got %+v, want %+v", r.TotalTime, wantTotal)
	}
	if len(r.Steps) != 2 {
		t.Fatalf("steps: got %d, want 2", len(r.Steps))
	}
	if len(r.Steps[0].Times) != 1 || r.Steps[0].Times[0].StartTime != 3 {
		t.Fatalf("step 0 times: got %+v", r.Steps[0].Times)
	}
	if len(r.Steps[1].Times) != 1 || r.Steps[1].Times[0].EndTime != 35 {
		t.Fatalf("step 1 times: got %+v", r.Steps[1].Times)
	}
}

func TestParseRecipe_OmittedNutritionAndTime(t *testing.T) {
	r := parseRecipeString(t, `<recipe><title>Toast</title></recipe>`)

	if r.TotalTime != (models.RecipeTime{}) {
		t.Fatalf("total time should be zero, got %+v", r.TotalTime)
	}
	if r.Nutrition != (models.RecipeNutrition{}) {
		t.Fatalf("nutrition should be zero, got %+v", r.Nutrition)
	}
}

func TestParseRecipe_PartialNutrition(t *testing.T) {
	r := parseRecipeString(t, `<recipe>
		<nutrition serving_size="2">
			<nutrient type="calories" value="220"/>
			<nutrient type="protein_grams" value="9"/>
		</nutrition>
	</recipe>`)

	want := models.RecipeNutrition{
		NutritionServings: 2,
		Calories:          220,
		ProteinGrams:      9,
	}
	if r.Nutrition != want {
		t.Fatalf("nutrition: got %+v, want %+v", r.Nutrition, want)
	}
}

func TestParseRecipe_UnknownNutrientIgnored(t *testing.T) {
	r := parseRecipeString(t, `<recipe>
		<nutrition>
			<nutrient type="potassium_mg" value="400"/>
			<nutrient type="calories" value="150"/>
		</nutrition>
	</recipe>`)

	want := models.RecipeNutrition{Calories: 150}
	if r.Nutrition != want {
		t.Fatalf("nutrition: got %+v, want %+v", r.Nutrition, want)
	}
}

func TestParseRecipe_MalformedNutrientValue(t *testing.T) {
	doc := `<recipe><nutrition><nutrient type="calories" value="lots"/></nutrition></recipe>`

	_, err := parseRecipe(xml.NewDecoder(strings.NewReader(doc)))
	if err == nil {
		t.Fatal("expected an error for an unparseable nutrient value")
	}
	if !strings.Contains(err.Error(), "calories") {
		t.Fatalf("error should name the nutrient, got %v", err)
	}
}
