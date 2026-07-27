package local

import (
	"sort"

	"github.com/danielcbailey/Cookbook/core/models"
)

func (tx *localTransaction) SearchFoodKeeperProductsBySemanticSimilarity(embedding []float32, limit int) ([]*models.FoodKeeperProduct, error) {
	type scored struct {
		product models.FoodKeeperProduct
		score   float32
	}
	var candidates []scored
	for _, p := range tx.data.FoodKeeperProducts {
		if len(p.Embedding) > 0 {
			candidates = append(candidates, scored{
				product: p,
				score:   cosineSimilarity(p.Embedding, embedding),
			})
		}
	}
	// Map iteration is unordered, so break score ties by ID to keep results stable.
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].product.ID < candidates[j].product.ID
	})
	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}
	out := make([]*models.FoodKeeperProduct, len(candidates))
	for i, c := range candidates {
		cp := copyFoodKeeperProduct(c.product)
		out[i] = &cp
	}
	return out, nil
}

func (tx *localTransaction) DeleteAllFoodKeeperProducts() error {
	tx.data.FoodKeeperProducts = make(map[int64]models.FoodKeeperProduct)
	return nil
}

func (tx *localTransaction) CreateFoodKeeperProduct(product *models.FoodKeeperProduct) error {
	tx.data.FoodKeeperProducts[product.ID] = copyFoodKeeperProduct(*product)
	return nil
}

func copyFoodKeeperProduct(p models.FoodKeeperProduct) models.FoodKeeperProduct {
	p.Embedding = copyFloat32s(p.Embedding)
	p.Pantry = copyFoodKeeperDuration(p.Pantry)
	p.DOPPantry = copyFoodKeeperDuration(p.DOPPantry)
	p.PantryAfterOpening = copyFoodKeeperDuration(p.PantryAfterOpening)
	p.Refrigerate = copyFoodKeeperDuration(p.Refrigerate)
	p.DOPRefrigerate = copyFoodKeeperDuration(p.DOPRefrigerate)
	p.RefrigerateAfterOpening = copyFoodKeeperDuration(p.RefrigerateAfterOpening)
	p.RefrigerateAfterThawing = copyFoodKeeperDuration(p.RefrigerateAfterThawing)
	p.Freeze = copyFoodKeeperDuration(p.Freeze)
	p.DOPFreeze = copyFoodKeeperDuration(p.DOPFreeze)
	return p
}

func copyFoodKeeperDuration(d models.FoodKeeperDuration) models.FoodKeeperDuration {
	d.Min = copyFloat64Ptr(d.Min)
	d.Max = copyFloat64Ptr(d.Max)
	return d
}

func copyFloat64Ptr(v *float64) *float64 {
	if v == nil {
		return nil
	}
	c := *v
	return &c
}
