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
var ingredientAssociationPromptTemplate string

//go:embed ingredientCreationPrompt.txt
var ingredientCreationPromptTemplate string

// The prompts are static once the enum values are substituted in, so they are
// rendered once here rather than on every request.
var (
	ingredientAssociationPrompt = renderPrompt(ingredientAssociationPromptTemplate)
	ingredientCreationPrompt    = renderPrompt(ingredientCreationPromptTemplate)
)

// AssociateIngredient solves the problem of picking the best matching ingredient from a list of close candidates - if any at all. If the model determines
// that none of the provided candidates are a suitable match, it will return a new ingredient with ID 0 and the name of the suggested ingredient.
func AssociateIngredient(ctx context.Context, providers config.Providers, candidates []*models.Ingredient, queryContext, query string) (*models.Ingredient, error) {
	var ingredientIDStr string
	var err error

	attempts := 2
	for range attempts {
		ingredientIDStr, err = ingredientAssociation(ctx, providers, ingredientAssociationPrompt, buildIngredientAssociationUserMessage(candidates, queryContext, query))
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, err
	}

	ingredientID, err := strconv.ParseInt(ingredientIDStr, 10, 64)
	if err != nil {
		return &models.Ingredient{
			ID:   0,
			Name: ingredientIDStr,
		}, nil
	}

	for _, candidate := range candidates {
		if candidate.ID == ingredientID {
			return candidate, nil
		}
	}

	return nil, fmt.Errorf("ingredient ID %d not found in candidates", ingredientID)
}

func AssociateNewIngredient(ctx context.Context, providers config.Providers, candidates []*models.FoodKeeperProduct, query string) (foodkeeper *models.FoodKeeperProduct, density float64, category string, retErr error) {
	var ingredientIDStr string
	var err error

	attempts := 2
	for range attempts {
		ingredientIDStr, category, density, err = ingredientCreation(ctx, providers, ingredientCreationPrompt, buildFoodkeeperAssociationUserMessage(candidates, query))
		if err == nil {
			break
		}
	}

	if err != nil {
		providers.Log().Warn("failed to get model output for ingredient creation", slog.Any("error", err))
	}

	if density == 0 {
		density = 1
	}

	ingredientID, err := strconv.ParseInt(ingredientIDStr, 10, 64)
	if err != nil {
		foodkeeper = &models.FoodKeeperProduct{
			ID:   0,
			Name: ingredientIDStr,
		}
		return
	}

	for _, candidate := range candidates {
		if candidate.ID == ingredientID {
			foodkeeper = candidate
			return
		}
	}

	retErr = fmt.Errorf("foodkeeper ID %d not found in candidates", ingredientID)
	return
}

// returns primary value, and error
func ingredientAssociation(ctx context.Context, providers config.Providers, prompt, userMessage string) (string, error) {
	response, err := providers.OpenAI().Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:           openai.ChatModelGPT5_6Terra,
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
						OfString: param.NewOpt(userMessage),
					},
				},
			},
		},
	})

	if err != nil {
		return "", err
	}

	modelOutput := response.Choices[len(response.Choices)-1].Message.Content

	var answer struct {
		XMLName xml.Name `xml:"answer"`
		Value   string   `xml:",chardata"`
	}

	answerStart := strings.Index(modelOutput, "<answer")
	answerEnd := strings.Index(modelOutput, "</answer>")
	if answerStart == -1 || answerEnd == -1 {
		return "", fmt.Errorf("invalid model output: %s", modelOutput)
	}
	answerXML := modelOutput[answerStart : answerEnd+len("</answer>")]

	if err = xml.Unmarshal([]byte(answerXML), &answer); err != nil {
		return "", fmt.Errorf("invalid model output: %s", modelOutput)
	}

	idStr := strings.TrimSpace(answer.Value)

	if idStr == "" {
		providers.Log().Debug("invalid associate ingredients model output", slog.String("output", modelOutput))
	}

	return idStr, nil
}

// returns primary value, type, density, and error
func ingredientCreation(ctx context.Context, providers config.Providers, prompt, userMessage string) (value string, category string, density float64, retErr error) {
	response, err := providers.OpenAI().Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:           openai.ChatModelGPT5_6Terra,
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
						OfString: param.NewOpt(userMessage),
					},
				},
			},
		},
	})

	if err != nil {
		retErr = err
		return
	}

	modelOutput := response.Choices[len(response.Choices)-1].Message.Content

	var answer struct {
		XMLName xml.Name `xml:"answer"`
		Type    string   `xml:"type,attr"`
		Density float64  `xml:"density,attr"`
		Value   string   `xml:",chardata"`
	}

	answerStart := strings.Index(modelOutput, "<answer")
	answerEnd := strings.Index(modelOutput, "</answer>")
	if answerStart == -1 || answerEnd == -1 {
		retErr = fmt.Errorf("invalid model output: %s", modelOutput)
		return
	}
	answerXML := modelOutput[answerStart : answerEnd+len("</answer>")]

	if err = xml.Unmarshal([]byte(answerXML), &answer); err != nil {
		retErr = fmt.Errorf("invalid model output: %s", modelOutput)
		return
	}

	value = strings.TrimSpace(answer.Value)
	category = strings.TrimSpace(answer.Type)
	density = answer.Density

	if value == "" || density == 0 {
		retErr = fmt.Errorf("invalid associate ingredients model output: %s", modelOutput)
		return
	}

	retErr = nil
	return
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
		name := candidate.Name
		if candidate.NameSubtitle != "" {
			name = name + " - " + candidate.NameSubtitle
		}

		builder.WriteString(fmt.Sprintf("%d: %s\n", candidate.ID, name))
	}
	return builder.String()
}
