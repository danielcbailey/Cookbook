package ai

import (
	"context"
	"encoding/xml"
	"fmt"
	"log/slog"
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

//go:embed foodkeeperAssociationPrompt.txt
var foodkeeperAssociationPrompt string

// AssociateIngredient solves the problem of picking the best matching ingredient from a list of close candidates - if any at all. If the model determines
// that none of the provided candidates are a suitable match, it will return a new ingredient with ID 0 and the name of the suggested ingredient.
func AssociateIngredient(ctx context.Context, providers config.Providers, candidates []*models.Ingredient, queryContext, query string) (*models.Ingredient, error) {
	ingredientIDStr, category, err := ingredientAssociationBase(ctx, providers, buildIngredientAssociationUserMessage(candidates, queryContext, query), ingredientAssociationPrompt)
	if err != nil {
		return nil, err
	}

	ingredientID, err := strconv.ParseInt(ingredientIDStr, 10, 64)
	if err != nil {
		return &models.Ingredient{
			ID:       0,
			Name:     ingredientIDStr,
			Category: category,
		}, nil
	}

	for _, candidate := range candidates {
		if candidate.ID == ingredientID {
			return candidate, nil
		}
	}

	return nil, fmt.Errorf("ingredient ID %d not found in candidates", ingredientID)
}

func AssociateFoodkeeper(ctx context.Context, providers config.Providers, candidates []*models.FoodKeeperProduct, query string) (*models.FoodKeeperProduct, error) {
	ingredientIDStr, _, err := ingredientAssociationBase(ctx, providers, buildFoodkeeperAssociationUserMessage(candidates, query), foodkeeperAssociationPrompt)

	ingredientID, err := strconv.ParseInt(ingredientIDStr, 10, 64)
	if err != nil {
		return &models.FoodKeeperProduct{
			ID:   0,
			Name: ingredientIDStr,
		}, nil
	}

	for _, candidate := range candidates {
		if candidate.ID == ingredientID {
			return candidate, nil
		}
	}

	return nil, fmt.Errorf("foodkeeper ID %d not found in candidates", ingredientID)
}

// returns primary value, type, and error
func ingredientAssociationBase(ctx context.Context, providers config.Providers, prompt, userMessage string) (string, string, error) {
	response, err := providers.OpenAI().Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:           openai.ChatModelGPT5_6Terra,
		Temperature:     param.NewOpt[float64](0.2),
		ReasoningEffort: openai.ReasoningEffortNone,
		Messages: []openai.ChatCompletionMessageParamUnion{
			{
				OfSystem: &openai.ChatCompletionSystemMessageParam{
					Content: openai.ChatCompletionSystemMessageParamContentUnion{
						OfString: param.NewOpt[string](prompt),
					},
				},
			},
			{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: param.NewOpt[string](userMessage),
					},
				},
			},
		},
	})

	if err != nil {
		return "", "", err
	}

	modelOutput := response.Choices[len(response.Choices)-1].Message.Content

	var answer struct {
		XMLName xml.Name `xml:"answer"`
		Type    string   `xml:"type,attr"`
		Value   string   `xml:",chardata"`
	}

	answerStart := strings.Index(modelOutput, "<answer")
	answerEnd := strings.Index(modelOutput, "</answer>")
	if answerStart == -1 || answerEnd == -1 {
		return "", "", fmt.Errorf("invalid model output: %s", modelOutput)
	}
	answerXML := modelOutput[answerStart : answerEnd+len("</answer>")]

	if err = xml.Unmarshal([]byte(answerXML), &answer); err != nil {
		return "", "", fmt.Errorf("invalid model output: %s", modelOutput)
	}

	idStr := strings.TrimSpace(answer.Value)
	category := strings.TrimSpace(answer.Type)

	if idStr == "" {
		providers.Log().Debug("invalid associate ingredients model output", slog.String("output", modelOutput))
	}

	return idStr, category, nil
}

func buildIngredientAssociationUserMessage(candidates []*models.Ingredient, queryContext, query string) string {
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

func buildFoodkeeperAssociationUserMessage(candidates []*models.FoodKeeperProduct, query string) string {
	builder := strings.Builder{}
	builder.WriteString("QUERY: ")
	builder.WriteString(query)
	builder.WriteString("\nTOP CANDIDATES:\n")
	for _, candidate := range candidates {
		builder.WriteString(fmt.Sprintf("%d: %s\n", candidate.ID, candidate.Name))
	}
	return builder.String()
}
