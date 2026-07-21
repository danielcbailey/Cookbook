package ai

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/danielcbailey/Cookbook/core/config"
	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"

	_ "embed"
)

//go:embed ingredientAssociationPrompt.txt
var ingredientAssociationPrompt string

// AssociateIngredient solves the problem of picking the best matching ingredient from a list of close candidates - if any at all. If the model determines
// that none of the provided candidates are a suitable match, it will return a new ingredient with ID 0 and the name of the suggested ingredient.
func AssociateIngredient(ctx context.Context, providers config.Providers, candidates []models.Ingredient, queryContext, query string) (models.Ingredient, error) {
	response, err := providers.OpenAI().Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:           openai.ChatModelGPT5_6Terra,
		Temperature:     param.NewOpt[float64](0.2),
		ReasoningEffort: openai.ReasoningEffortNone,
		Messages: []openai.ChatCompletionMessageParamUnion{
			{
				OfSystem: &openai.ChatCompletionSystemMessageParam{
					Content: openai.ChatCompletionSystemMessageParamContentUnion{
						OfString: param.NewOpt[string](ingredientAssociationPrompt),
					},
				},
			},
			{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: param.NewOpt[string](buildIngredientAssociationUserMessage(candidates, queryContext, query)),
					},
				},
			},
		},
	})

	if err != nil {
		return models.Ingredient{}, err
	}

	// Answer is in the format of <answer>x</answer>
	modelOutput := response.Choices[len(response.Choices)-1].Message.Content
	prefixParts := strings.Split(modelOutput, "<answer>")
	if len(prefixParts) < 2 {
		return models.Ingredient{}, fmt.Errorf("invalid model output: %s", modelOutput)
	}
	suffixParts := strings.Split(prefixParts[1], "</answer>")
	if len(suffixParts) < 2 {
		return models.Ingredient{}, fmt.Errorf("invalid model output: %s", modelOutput)
	}

	ingredientIDStr := strings.TrimSpace(suffixParts[0])
	ingredientID, err := strconv.ParseInt(ingredientIDStr, 10, 64)
	if err != nil {
		// It is suggesting a new ingredient
		return models.Ingredient{
			ID:   0,
			Name: ingredientIDStr,
		}, nil
	}

	for _, candidate := range candidates {
		if candidate.ID == ingredientID {
			return candidate, nil
		}
	}

	return models.Ingredient{}, fmt.Errorf("ingredient ID %d not found in candidates", ingredientID)
}

func buildIngredientAssociationUserMessage(candidates []models.Ingredient, queryContext, query string) string {
	builder := strings.Builder{}
	builder.WriteString("CONTEXT: ")
	builder.WriteString(queryContext)
	builder.WriteString("\nQUERY: ")
	builder.WriteString(query)
	builder.WriteString("\nTOP CANDIDATES:\n")
	for _, candidate := range candidates {
		builder.WriteString(fmt.Sprintf("%d: %s\n", candidate.ID, candidate.Name))
	}
	return builder.String()
}
