package ai

import (
	"context"
	"math"
	"sort"

	"github.com/danielcbailey/Cookbook/core/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
)

// ConvertToEmbedding takes a string and returns a slice of 1536 float64's representing the embedding of the input text.
func ConvertToEmbedding(ctx context.Context, providers config.Providers, text string) ([]float64, error) {
	response, err := providers.OpenAI().Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: param.NewOpt(text),
		},
		Model:      openai.EmbeddingModelTextEmbedding3Small,
		Dimensions: param.NewOpt[int64](1536),
	})

	if err != nil {
		return nil, err
	}
	return response.Data[0].Embedding, nil
}

func CosineSimilaritySort(embeddings [][]float64, queryEmbedding []float64) ([]int, error) {
	similarities := make([]float64, len(embeddings))
	for i, emb := range embeddings {
		similarities[i] = CosineSimilarity(emb, queryEmbedding)
	}

	indices := make([]int, len(embeddings))
	for i := range indices {
		indices[i] = i
	}

	sort.Slice(indices, func(a, b int) bool {
		return similarities[indices[a]] > similarities[indices[b]]
	})

	return indices, nil
}

func CosineSimilarity(a, b []float64) float64 {
	var dot, magA, magB float64
	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}
	if magA == 0 || magB == 0 {
		return 0
	}
	return dot / (math.Sqrt(magA) * math.Sqrt(magB))
}
