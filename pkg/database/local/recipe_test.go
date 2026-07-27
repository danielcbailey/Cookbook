package local_test

import (
	"context"
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/pkg/database"
	"github.com/danielcbailey/Cookbook/pkg/database/local"
)

func setupRecipeTest(t *testing.T) (database.Transaction, models.User) {
	t.Helper()
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.NewTransaction(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { tx.Rollback() })

	u := testUser()
	if _, err := tx.CreateUser(&u); err != nil {
		t.Fatal(err)
	}
	return tx, u
}

func TestCreateRecipe_AssignsIDAndTimestamps(t *testing.T) {
	tx, u := setupRecipeTest(t)

	r := testRecipe(u.ID)
	id, err := tx.CreateRecipe(&r)
	if err != nil {
		t.Fatal(err)
	}

	if id == 0 {
		t.Fatal("expected non-zero ID")
	}
	if r.ID != id {
		t.Fatalf("expected r.ID %d to match returned ID %d", r.ID, id)
	}
	if r.CreatedAt == 0 {
		t.Fatal("expected non-zero CreatedAt")
	}
	if r.UpdatedAt == 0 {
		t.Fatal("expected non-zero UpdatedAt")
	}
}

func TestGetRecipeByID_FullGraph(t *testing.T) {
	tx, u := setupRecipeTest(t)

	r := testRecipe(u.ID)
	tx.CreateRecipe(&r)

	step := models.RecipeStep{
		RecipeID: r.ID,
		Index:    0,
		Title:    "Boil water",
		BodyText: "Bring water to a boil",
		Ingredients: []models.RecipeIngredient{
			{
				Index:      0,
				Ingredient: models.Ingredient{ID: 1, Name: "water"},
				Quantity:   1,
				Unit:       models.IngredientUnitLiters,
			},
		},
		Times: []models.RecipeTime{
			{Index: 0, StartTime: 10, Unit: models.RecipeTimeUnitMinutes},
		},
	}
	if _, err := tx.CreateOrUpdateRecipeStep(&step); err != nil {
		t.Fatal(err)
	}

	got, err := tx.GetRecipeByID(r.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Pasta" {
		t.Fatalf("title: got %q, want %q", got.Title, "Pasta")
	}
	if len(got.Tags) != 2 {
		t.Fatalf("tags: got %d, want 2", len(got.Tags))
	}
	if len(got.Steps) != 1 {
		t.Fatalf("steps: got %d, want 1", len(got.Steps))
	}
	if len(got.Steps[0].Ingredients) != 1 {
		t.Fatalf("step ingredients: got %d, want 1", len(got.Steps[0].Ingredients))
	}
	if len(got.Steps[0].Times) != 1 {
		t.Fatalf("step times: got %d, want 1", len(got.Steps[0].Times))
	}
}

func TestGetRecipeByID_NotFound(t *testing.T) {
	tx, _ := setupRecipeTest(t)

	_, err := tx.GetRecipeByID(999)
	if !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetRecipeByID_ReturnsCopy(t *testing.T) {
	tx, u := setupRecipeTest(t)

	r := testRecipe(u.ID)
	tx.CreateRecipe(&r)

	got, _ := tx.GetRecipeByID(r.ID)
	got.Title = "mutated"
	got.Tags[0].Name = "mutated"

	got2, _ := tx.GetRecipeByID(r.ID)
	if got2.Title == "mutated" {
		t.Fatal("modifying returned recipe should not affect store")
	}
	if got2.Tags[0].Name == "mutated" {
		t.Fatal("modifying returned tags should not affect store")
	}
}

func TestUpdateRecipe_CoreFieldsOnly(t *testing.T) {
	tx, u := setupRecipeTest(t)

	r := testRecipe(u.ID)
	tx.CreateRecipe(&r)

	step := models.RecipeStep{
		RecipeID: r.ID,
		Index:    0,
		Title:    "Step 1",
	}
	tx.CreateOrUpdateRecipeStep(&step)

	r.Title = "Updated Pasta"
	r.Steps = nil
	if err := tx.UpdateRecipe(&r); err != nil {
		t.Fatal(err)
	}

	got, _ := tx.GetRecipeByID(r.ID)
	if got.Title != "Updated Pasta" {
		t.Fatalf("title: got %q, want %q", got.Title, "Updated Pasta")
	}
	if len(got.Steps) != 1 {
		t.Fatal("UpdateRecipe should preserve existing steps")
	}
	if got.UpdatedAt < got.CreatedAt {
		t.Fatal("UpdatedAt should be >= CreatedAt")
	}
}

func TestDeleteRecipe(t *testing.T) {
	tx, u := setupRecipeTest(t)

	r := testRecipe(u.ID)
	tx.CreateRecipe(&r)

	if err := tx.DeleteRecipe(r.ID); err != nil {
		t.Fatal(err)
	}
	_, err := tx.GetRecipeByID(r.ID)
	if !errors.Is(err, database.ErrNotFound) {
		t.Fatal("recipe should be deleted")
	}
}

func TestListRecipesByUserID_OrderAndShallowCopy(t *testing.T) {
	tx, u := setupRecipeTest(t)

	r1 := testRecipe(u.ID)
	r1.Title = "First"
	tx.CreateRecipe(&r1)

	r2 := testRecipe(u.ID)
	r2.Title = "Second"
	tx.CreateRecipe(&r2)

	step := models.RecipeStep{RecipeID: r1.ID, Index: 0, Title: "S"}
	tx.CreateOrUpdateRecipeStep(&step)

	list, err := tx.ListRecipesByUserID(u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("got %d recipes, want 2", len(list))
	}
	if list[0].Title != "Second" {
		t.Fatal("should be ordered by CreatedAt DESC (newest first)")
	}
	if list[0].Steps != nil {
		t.Fatal("list should not include steps")
	}
	if list[0].Ingredients != nil {
		t.Fatal("list should not include ingredients")
	}
	if list[0].Embedding != nil {
		t.Fatal("list should not include embedding")
	}
	if len(list[0].Tags) == 0 {
		t.Fatal("list should include tags")
	}
}

func TestListRecipesByUserID_FiltersByUser(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tx, _ := db.NewTransaction(t.Context())
	defer tx.Rollback()

	u1 := testUser()
	u1.Email = "a@test.com"
	tx.CreateUser(&u1)

	u2 := testUser()
	u2.Email = "b@test.com"
	tx.CreateUser(&u2)

	r1 := testRecipe(u1.ID)
	tx.CreateRecipe(&r1)

	r2 := testRecipe(u2.ID)
	tx.CreateRecipe(&r2)

	list, _ := tx.ListRecipesByUserID(u1.ID)
	if len(list) != 1 {
		t.Fatalf("got %d, want 1 recipe for user 1", len(list))
	}
}

func TestSearchRecipesBySemanticSimilarity(t *testing.T) {
	tx, u := setupRecipeTest(t)

	close := models.Recipe{
		UserID:    u.ID,
		Title:     "Close match",
		Embedding: []float32{1, 0, 0},
	}
	tx.CreateRecipe(&close)

	far := models.Recipe{
		UserID:    u.ID,
		Title:     "Far match",
		Embedding: []float32{0, 1, 0},
	}
	tx.CreateRecipe(&far)

	noEmbed := models.Recipe{
		UserID: u.ID,
		Title:  "No embedding",
	}
	tx.CreateRecipe(&noEmbed)

	query := []float32{1, 0.1, 0}
	results, err := tx.SearchRecipesBySemanticSimilarity(u.ID, query, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d, want 2 (recipes with embeddings only)", len(results))
	}
	if results[0].Title != "Close match" {
		t.Fatalf("closest should be first, got %q", results[0].Title)
	}
}

func TestSearchRecipesBySemanticSimilarity_Limit(t *testing.T) {
	tx, u := setupRecipeTest(t)

	for i := 0; i < 5; i++ {
		r := models.Recipe{
			UserID:    u.ID,
			Title:     "Recipe",
			Embedding: []float32{1, 0, 0},
		}
		tx.CreateRecipe(&r)
	}

	results, _ := tx.SearchRecipesBySemanticSimilarity(u.ID, []float32{1, 0, 0}, 3)
	if len(results) != 3 {
		t.Fatalf("got %d, want 3", len(results))
	}
}

func TestCreateOrUpdateRecipeStep_Create(t *testing.T) {
	tx, u := setupRecipeTest(t)

	r := testRecipe(u.ID)
	tx.CreateRecipe(&r)

	step := models.RecipeStep{
		RecipeID: r.ID,
		Index:    0,
		Title:    "Step one",
		Ingredients: []models.RecipeIngredient{
			{
				Index:      0,
				Ingredient: models.Ingredient{ID: 1, Name: "salt"},
				Quantity:   1,
				Unit:       models.IngredientUnitTeaspoons,
			},
		},
	}
	id, err := tx.CreateOrUpdateRecipeStep(&step)
	if err != nil {
		t.Fatal(err)
	}
	if id == 0 {
		t.Fatal("step should get an ID")
	}
	if step.ID != id {
		t.Fatalf("expected step.ID %d to match returned ID %d", step.ID, id)
	}
	if step.Ingredients[0].ID == 0 {
		t.Fatal("ingredient should get an ID")
	}

	got, _ := tx.GetRecipeByID(r.ID)
	if len(got.Steps) != 1 {
		t.Fatalf("steps: got %d, want 1", len(got.Steps))
	}
}

func TestCreateOrUpdateRecipeStep_Update(t *testing.T) {
	tx, u := setupRecipeTest(t)

	r := testRecipe(u.ID)
	tx.CreateRecipe(&r)

	step := models.RecipeStep{RecipeID: r.ID, Index: 0, Title: "Original"}
	origID, err := tx.CreateOrUpdateRecipeStep(&step)
	if err != nil {
		t.Fatal(err)
	}

	step2 := models.RecipeStep{RecipeID: r.ID, Index: 0, Title: "Updated"}
	gotID, err := tx.CreateOrUpdateRecipeStep(&step2)
	if err != nil {
		t.Fatal(err)
	}

	if gotID != origID {
		t.Fatalf("upsert should reuse step ID: got %d, want %d", gotID, origID)
	}
	if step2.ID != gotID {
		t.Fatalf("expected step2.ID %d to match returned ID %d", step2.ID, gotID)
	}

	got, _ := tx.GetRecipeByID(r.ID)
	if len(got.Steps) != 1 {
		t.Fatalf("should still have 1 step, got %d", len(got.Steps))
	}
	if got.Steps[0].Title != "Updated" {
		t.Fatalf("title: got %q, want %q", got.Steps[0].Title, "Updated")
	}
}

func TestCreateOrUpdateRecipeStep_ReplacesChildren(t *testing.T) {
	tx, u := setupRecipeTest(t)

	r := testRecipe(u.ID)
	tx.CreateRecipe(&r)

	step := models.RecipeStep{
		RecipeID: r.ID,
		Index:    0,
		Title:    "Step",
		Ingredients: []models.RecipeIngredient{
			{Index: 0, Ingredient: models.Ingredient{ID: 1, Name: "a"}, Quantity: 1},
			{Index: 1, Ingredient: models.Ingredient{ID: 2, Name: "b"}, Quantity: 2},
		},
		Times: []models.RecipeTime{
			{Index: 0, StartTime: 5, Unit: models.RecipeTimeUnitMinutes},
		},
	}
	tx.CreateOrUpdateRecipeStep(&step)

	step2 := models.RecipeStep{
		RecipeID: r.ID,
		Index:    0,
		Title:    "Step",
		Ingredients: []models.RecipeIngredient{
			{Index: 0, Ingredient: models.Ingredient{ID: 3, Name: "c"}, Quantity: 3},
		},
	}
	tx.CreateOrUpdateRecipeStep(&step2)

	got, _ := tx.GetRecipeByID(r.ID)
	if len(got.Steps[0].Ingredients) != 1 {
		t.Fatalf("ingredients should be replaced: got %d, want 1", len(got.Steps[0].Ingredients))
	}
	if got.Steps[0].Ingredients[0].Ingredient.Name != "c" {
		t.Fatalf("ingredient: got %q, want %q", got.Steps[0].Ingredients[0].Ingredient.Name, "c")
	}
	if len(got.Steps[0].Times) != 0 {
		t.Fatalf("times should be replaced: got %d, want 0", len(got.Steps[0].Times))
	}
}

func TestDeleteRecipeStep(t *testing.T) {
	tx, u := setupRecipeTest(t)

	r := testRecipe(u.ID)
	tx.CreateRecipe(&r)

	s0 := models.RecipeStep{RecipeID: r.ID, Index: 0, Title: "Keep"}
	tx.CreateOrUpdateRecipeStep(&s0)
	s1 := models.RecipeStep{RecipeID: r.ID, Index: 1, Title: "Remove"}
	tx.CreateOrUpdateRecipeStep(&s1)

	if err := tx.DeleteRecipeStep(&s1); err != nil {
		t.Fatal(err)
	}

	got, _ := tx.GetRecipeByID(r.ID)
	if len(got.Steps) != 1 {
		t.Fatalf("steps: got %d, want 1", len(got.Steps))
	}
	if got.Steps[0].Title != "Keep" {
		t.Fatalf("wrong step kept: got %q", got.Steps[0].Title)
	}
}

func TestDeleteRecipeStep_ByIDWithoutRecipeID(t *testing.T) {
	tx, u := setupRecipeTest(t)

	r := testRecipe(u.ID)
	tx.CreateRecipe(&r)

	step := models.RecipeStep{RecipeID: r.ID, Index: 0, Title: "Step"}
	tx.CreateOrUpdateRecipeStep(&step)

	orphanRef := &models.RecipeStep{ID: step.ID}
	if err := tx.DeleteRecipeStep(orphanRef); err != nil {
		t.Fatal(err)
	}

	got, _ := tx.GetRecipeByID(r.ID)
	if len(got.Steps) != 0 {
		t.Fatal("step should be deleted via ID scan")
	}
}

func TestTagReconciliation_Dedup(t *testing.T) {
	tx, u := setupRecipeTest(t)

	r1 := testRecipe(u.ID)
	r1.Tags = []models.RecipeTag{{Name: "shared-tag"}}
	tx.CreateRecipe(&r1)

	r2 := testRecipe(u.ID)
	r2.Tags = []models.RecipeTag{{Name: "shared-tag"}}
	tx.CreateRecipe(&r2)

	got1, _ := tx.GetRecipeByID(r1.ID)
	got2, _ := tx.GetRecipeByID(r2.ID)

	if got1.Tags[0].ID != got2.Tags[0].ID {
		t.Fatalf("same tag name should get same ID: %d != %d", got1.Tags[0].ID, got2.Tags[0].ID)
	}
}

func TestTagReconciliation_UpdateReplacesTagSet(t *testing.T) {
	tx, u := setupRecipeTest(t)

	r := testRecipe(u.ID)
	r.Tags = []models.RecipeTag{{Name: "old"}}
	tx.CreateRecipe(&r)

	r.Tags = []models.RecipeTag{{Name: "new"}}
	tx.UpdateRecipe(&r)

	got, _ := tx.GetRecipeByID(r.ID)
	if len(got.Tags) != 1 {
		t.Fatalf("tags: got %d, want 1", len(got.Tags))
	}
	if got.Tags[0].Name != "new" {
		t.Fatalf("tag: got %q, want %q", got.Tags[0].Name, "new")
	}
}

func randEmbedding(rng *rand.Rand, dims int) []float32 {
	v := make([]float32, dims)
	for i := range v {
		v[i] = rng.Float32()*2 - 1
	}
	return v
}

func BenchmarkSearchRecipesBySemanticSimilarity_1000(b *testing.B) {
	const numRecipes = 1000
	const dims = 1536

	db, err := local.NewLocal(b.TempDir())
	if err != nil {
		b.Fatal(err)
	}

	tx, err := db.NewTransaction(context.Background())
	if err != nil {
		b.Fatal(err)
	}

	u := models.User{Email: "bench@test.com", PasswordHash: "h"}
	if _, err := tx.CreateUser(&u); err != nil {
		b.Fatal(err)
	}

	rng := rand.New(rand.NewPCG(42, 0))
	for i := 0; i < numRecipes; i++ {
		r := models.Recipe{
			UserID:    u.ID,
			Title:     "Recipe",
			Embedding: randEmbedding(rng, dims),
		}
		if _, err := tx.CreateRecipe(&r); err != nil {
			b.Fatal(err)
		}
	}

	if err := tx.Commit(); err != nil {
		b.Fatal(err)
	}

	query := randEmbedding(rng, dims)

	b.ResetTimer()
	for b.Loop() {
		tx, err := db.NewTransaction(context.Background())
		if err != nil {
			b.Fatal(err)
		}
		results, err := tx.SearchRecipesBySemanticSimilarity(u.ID, query, 10)
		if err != nil {
			b.Fatal(err)
		}
		if len(results) != 10 {
			b.Fatalf("got %d results, want 10", len(results))
		}
		tx.Rollback()
	}
}
