package ai

import (
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/danielcbailey/Cookbook/core/apicommon"
	"github.com/danielcbailey/Cookbook/core/config"
	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"

	_ "embed"
)

//go:embed recipeExtractionPrompt.txt
var recipeExtractionPrompt string

const (
	recipeMediumHTML  = "the HTML content"
	recipeMediumPhoto = "photos"
)

func getRecipeExtractionPrompt(medium string) string {
	return strings.ReplaceAll(recipeExtractionPrompt, "<MEDIUM>", medium)
}

// ExtractRecipeFromMedium takes the provided HTML and extracts a Recipe object.
func ExtractRecipeFromHTML(ctx context.Context, providers config.Providers, text string) (*models.Recipe, error) {
	prompt := getRecipeExtractionPrompt(recipeMediumHTML)

	response, err := providers.OpenAI().Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:           openai.ChatModelGPT5_6Sol,
		Temperature:     param.NewOpt(0.2),
		ReasoningEffort: openai.ReasoningEffortNone,
		Messages: []openai.ChatCompletionMessageParamUnion{
			{
				OfSystem: &openai.ChatCompletionSystemMessageParam{
					Content: openai.ChatCompletionSystemMessageParamContentUnion{
						OfString: param.NewOpt(prompt),
					},
				},
			},
			{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: param.NewOpt(text),
					},
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	modelOutput := response.Choices[len(response.Choices)-1].Message.Content
	if strings.HasPrefix(modelOutput, "<error>Not a recipe</error>") {
		return nil, apicommon.NewUserFacingError("the provided website is not a recipe")
	}

	decoder := xml.NewDecoder(strings.NewReader(modelOutput))

	return parseRecipe(decoder)
}

type FileAttachment struct {
	MimeType string
	Content  []byte
}

func ExtractRecipeFromPhotos(ctx context.Context, providers config.Providers, files []FileAttachment) (*models.Recipe, error) {
	prompt := getRecipeExtractionPrompt(recipeMediumPhoto)

	contentParts := make([]openai.ChatCompletionContentPartUnionParam, len(files))
	for _, file := range files {
		imageURL := &openai.ChatCompletionContentPartImageParam{
			ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
				URL: "data:" + file.MimeType + ";base64," + base64.StdEncoding.EncodeToString(file.Content),
			},
		}

		contentParts = append(contentParts, openai.ChatCompletionContentPartUnionParam{
			OfImageURL: imageURL,
		})
	}

	response, err := providers.OpenAI().Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:           openai.ChatModelGPT5_6Sol,
		Temperature:     param.NewOpt(0.2),
		ReasoningEffort: openai.ReasoningEffortNone,
		Messages: []openai.ChatCompletionMessageParamUnion{
			{
				OfSystem: &openai.ChatCompletionSystemMessageParam{
					Content: openai.ChatCompletionSystemMessageParamContentUnion{
						OfString: param.NewOpt(prompt),
					},
				},
			},
			{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfArrayOfContentParts: contentParts,
					},
				},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	modelOutput := response.Choices[len(response.Choices)-1].Message.Content
	if strings.HasPrefix(modelOutput, "<error>Not a recipe</error>") {
		return nil, apicommon.NewUserFacingError("the provided photo is not a recipe")
	}

	decoder := xml.NewDecoder(strings.NewReader(modelOutput))

	return parseRecipe(decoder)
}

func parseRecipe(decoder *xml.Decoder) (*models.Recipe, error) {
	ret := &models.Recipe{}

	for {
		tok, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		switch se.Name.Local {
		case "title":
			ret.Title, err = readTextContent(decoder)
		case "description":
			ret.Description, err = readTextContent(decoder)
		case "img":
			ret.ImageURL = getAttr(se, "src")
		case "meta":
			ret.Author = getAttr(se, "author")
			ret.Publisher = getAttr(se, "source")
			ret.Category = getAttr(se, "category")
			ret.Protein = getAttr(se, "protein")
			ret.SuggestedMeal = getAttr(se, "suggested_meal")
			ret.Cuisine = getAttr(se, "cuisine")
			servingsStr := getAttr(se, "servings")
			ret.Servings, err = strconv.Atoi(servingsStr)
			if err != nil {
				err = fmt.Errorf("meta failed to parse servings '%s': %w", servingsStr, err)
			}
		case "time":
			// Times nested in steps are consumed by parseStep, so a <time> seen
			// here is the recipe's total time.
			ret.TotalTime, err = parseTimeElement(se)
		case "nutrition":
			ret.Nutrition, err = parseNutrition(decoder, se)
		case "ingredients":
			ret.Ingredients, err = parseIngredients(decoder)
		case "steps":
			ret.Steps, err = parseSteps(decoder)
		}
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func readTextContent(decoder *xml.Decoder) (string, error) {
	var buf strings.Builder
	for {
		tok, err := decoder.Token()
		if err != nil {
			return "", err
		}
		switch t := tok.(type) {
		case xml.CharData:
			buf.Write(t)
		case xml.EndElement:
			return strings.TrimSpace(buf.String()), nil
		}
	}
}

func getAttr(el xml.StartElement, name string) string {
	for _, a := range el.Attr {
		if a.Name.Local == name {
			return a.Value
		}
	}
	return ""
}

func parseIngredientElement(el xml.StartElement) (models.RecipeIngredient, error) {
	name := getAttr(el, "name")
	qtyStr := getAttr(el, "quantity")
	qty, err := strconv.ParseFloat(qtyStr, 32)
	if err != nil {
		return models.RecipeIngredient{}, fmt.Errorf("ingredient '%s' failed to parse quantity '%s': %w", name, qtyStr, err)
	}
	return models.RecipeIngredient{
		Ingredient: models.Ingredient{Name: name},
		Quantity:   float32(qty),
		Unit:       models.IngredientUnit(getAttr(el, "unit")),
	}, nil
}

func parseTimeElement(el xml.StartElement) (models.RecipeTime, error) {
	t := models.RecipeTime{Unit: models.RecipeTimeUnit(getAttr(el, "unit"))}
	if v := getAttr(el, "exact"); v != "" {
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return t, fmt.Errorf("time failed to parse exact '%s': %w", v, err)
		}
		t.StartTime = float32(f)
	}
	if v := getAttr(el, "rangeStart"); v != "" {
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return t, fmt.Errorf("time failed to parse rangeStart '%s': %w", v, err)
		}
		t.StartTime = float32(f)
	}
	if v := getAttr(el, "rangeEnd"); v != "" {
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return t, fmt.Errorf("time failed to parse rangeEnd '%s': %w", v, err)
		}
		t.EndTime = float32(f)
	}
	return t, nil
}

// parseNutrition reads a <nutrition> element and its <nutrient> children. The
// prompt lets the model omit any nutrient the source recipe does not state, so
// missing attributes and missing nutrients are left at zero rather than
// treated as errors. Values that are present but unparseable are still errors.
func parseNutrition(decoder *xml.Decoder, el xml.StartElement) (models.RecipeNutrition, error) {
	n := models.RecipeNutrition{}

	if v := getAttr(el, "serving_size"); v != "" {
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return n, fmt.Errorf("nutrition failed to parse serving_size '%s': %w", v, err)
		}
		n.NutritionServings = float32(f)
	}
	if v := getAttr(el, "serving_mass_grams"); v != "" {
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return n, fmt.Errorf("nutrition failed to parse serving_mass_grams '%s': %w", v, err)
		}
		n.NutritionServingMass = float32(f)
		n.NutritionServingUnit = models.IngredientUnitGrams
	}

	for {
		tok, err := decoder.Token()
		if err != nil {
			return n, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "nutrient" {
				if err := applyNutrient(&n, t); err != nil {
					return n, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "nutrition" {
				return n, nil
			}
		}
	}
}

// applyNutrient records a single <nutrient> element on n. Unrecognized nutrient
// types are skipped, matching how the other parsers ignore elements they do not
// know about.
func applyNutrient(n *models.RecipeNutrition, el xml.StartElement) error {
	nutrientType := getAttr(el, "type")
	valueStr := getAttr(el, "value")
	if valueStr == "" {
		return nil
	}

	value, err := strconv.ParseFloat(valueStr, 32)
	if err != nil {
		return fmt.Errorf("nutrient '%s' failed to parse value '%s': %w", nutrientType, valueStr, err)
	}

	switch nutrientType {
	case "calories":
		n.Calories = int(value)
	case "total_fat_grams":
		n.TotalFatGrams = float32(value)
	case "saturated_fat_grams":
		n.SaturatedFatGrams = float32(value)
	case "trans_fat_grams":
		n.TransFatGrams = float32(value)
	case "cholestrol_mg":
		n.CholestrolMilligrams = float32(value)
	case "sodium_mg":
		n.SodiumMilligrams = float32(value)
	case "total_carbs_grams":
		n.TotalCarbsGrams = float32(value)
	case "dietary_fiber_grams":
		n.DietaryFiberGrams = float32(value)
	case "total_sugar_grams":
		n.TotalSugarGrams = float32(value)
	case "protein_grams":
		n.ProteinGrams = float32(value)
	}

	return nil
}

func parseIngredients(decoder *xml.Decoder) ([]models.RecipeIngredient, error) {
	var ingredients []models.RecipeIngredient
	for {
		tok, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "ingredient" {
				ingr, err := parseIngredientElement(t)
				if err != nil {
					return nil, err
				}
				ingredients = append(ingredients, ingr)
			}
		case xml.EndElement:
			if t.Name.Local == "ingredients" {
				return ingredients, nil
			}
		}
	}
}

func parseSteps(decoder *xml.Decoder) ([]models.RecipeStep, error) {
	var steps []models.RecipeStep
	for {
		tok, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "step" {
				step, err := parseStep(decoder, t)
				if err != nil {
					return nil, err
				}
				steps = append(steps, step)
			}
		case xml.EndElement:
			if t.Name.Local == "steps" {
				return steps, nil
			}
		}
	}
}

const brMarker = "\x00BR\x00"

func parseStep(decoder *xml.Decoder, el xml.StartElement) (models.RecipeStep, error) {
	step := models.RecipeStep{Title: getAttr(el, "title")}
	var body strings.Builder

	for {
		tok, err := decoder.Token()
		if err != nil {
			return step, err
		}
		switch t := tok.(type) {
		case xml.CharData:
			body.Write(t)
		case xml.StartElement:
			switch t.Name.Local {
			case "img":
				step.ImageURL = getAttr(t, "src")
			case "ingredient":
				ingr, err := parseIngredientElement(t)
				if err != nil {
					return step, err
				}
				fmt.Fprintf(&body, "{{ingr-%d}}", len(step.Ingredients))
				step.Ingredients = append(step.Ingredients, ingr)
			case "time":
				rt, err := parseTimeElement(t)
				if err != nil {
					return step, err
				}
				fmt.Fprintf(&body, "{{time-%d}}", len(step.Times))
				step.Times = append(step.Times, rt)
			case "br":
				body.WriteString(brMarker)
			}
		case xml.EndElement:
			if t.Name.Local == "step" {
				parts := strings.Split(body.String(), brMarker)
				for i, part := range parts {
					parts[i] = strings.Join(strings.Fields(part), " ")
				}
				step.BodyText = strings.TrimSpace(strings.Join(parts, "\n"))
				return step, nil
			}
		}
	}
}
