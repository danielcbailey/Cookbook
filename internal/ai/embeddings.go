package ai

import (
	"context"
	"encoding/binary"
	"errors"
	"log/slog"
	"math"
	"sort"
	"time"

	"github.com/danielcbailey/Cookbook/core/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
)

const cacheDuration = 7 * 24 * time.Hour // one week

// ConvertToEmbedding takes a string and returns a slice of 1536 float32's representing the embedding of the input text.
func ConvertToEmbedding(ctx context.Context, providers config.Providers, text string) ([]float32, error) {
	if embedding, err := getCacheEmbedding(ctx, providers, text); embedding != nil {
		return embedding, nil
	} else if err != nil {
		providers.Log().Warn("failed to check embedding cache", slog.Any("error", err))
	}

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

	ret := make([]float32, len(response.Data[0].Embedding))
	for i, v := range response.Data[0].Embedding {
		ret[i] = float32(v)
	}

	err = cacheEmbedding(ctx, providers, text, ret)
	if err != nil {
		providers.Log().Warn("failed to write to embedding cache", slog.Any("error", err))
	}

	return ret, nil
}

func cacheEmbedding(ctx context.Context, providers config.Providers, input string, embedding []float32) error {
	b := floatSliceToBytes(embedding)
	return providers.Cache().Set(ctx, embeddingKey(input), string(b), cacheDuration)
}

func getCacheEmbedding(ctx context.Context, providers config.Providers, input string) ([]float32, error) {
	str, err := providers.Cache().Get(ctx, embeddingKey(input))
	if err != nil {
		if providers.Cache().ErrIsNotFound(err) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return bytesToFloatSlice([]byte(str))
}

func floatSliceToBytes(arr []float32) []byte {
	b := make([]byte, len(arr)*4)
	for i, f := range arr {
		binary.LittleEndian.PutUint32(b[i*4:], math.Float32bits(f))
	}
	return b
}

func bytesToFloatSlice(b []byte) ([]float32, error) {
	if len(b)%4 != 0 {
		return nil, errors.New("byte slice must be divisible by four")
	}

	floats := make([]float32, len(b)/4)
	for i := range floats {
		bits := binary.LittleEndian.Uint32(b[i*4 : (i+1)*4])
		floats[i] = math.Float32frombits(bits)
	}

	return floats, nil
}

func embeddingKey(input string) string {
	return "embedding:v1:" + input
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
