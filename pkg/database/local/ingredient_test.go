package local_test

import (
	"testing"

	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/pkg/database"
	"github.com/danielcbailey/Cookbook/pkg/database/local"
)

const listIngredientsUserID = 42

// seedIngredients stores a global ingredient, two owned by the user under test,
// and one owned by somebody else, so every visibility branch is exercised.
func seedIngredients(t *testing.T) database.Transaction {
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

	// Inserted out of order so the sort is what produces the expected sequence.
	ingredients := []models.Ingredient{
		{UserID: listIngredientsUserID, Name: "cumin", Category: "spice"},
		{UserID: 0, Name: "salt", Category: "spice"},
		{UserID: listIngredientsUserID + 1, Name: "basil", Category: "herb"},
		{UserID: listIngredientsUserID, Name: "apple", Category: "produce"},
	}
	for i := range ingredients {
		if _, err := tx.CreateIngredient(&ingredients[i]); err != nil {
			t.Fatal(err)
		}
	}

	return tx
}

func ingredientNames(ingredients []*models.Ingredient) []string {
	out := make([]string, len(ingredients))
	for i, ing := range ingredients {
		out[i] = ing.Name
	}
	return out
}

func assertIngredientNames(t *testing.T, got []*models.Ingredient, want []string) {
	t.Helper()

	names := ingredientNames(got)
	if len(names) != len(want) {
		t.Fatalf("got %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("got %v, want %v", names, want)
		}
	}
}

// The other user's "basil" must never appear, and the global "salt" always must.
func TestListIngredients_ScopesToUserAndGlobal(t *testing.T) {
	tx := seedIngredients(t)

	got, err := tx.ListIngredients(listIngredientsUserID, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	assertIngredientNames(t, got, []string{"apple", "cumin", "salt"})
}

func TestListIngredients_Paging(t *testing.T) {
	tests := []struct {
		name   string
		offset int
		limit  int
		want   []string
	}{
		{name: "limit zero means unlimited", offset: 0, limit: 0, want: []string{"apple", "cumin", "salt"}},
		{name: "negative limit means unlimited", offset: 0, limit: -1, want: []string{"apple", "cumin", "salt"}},
		{name: "first page", offset: 0, limit: 2, want: []string{"apple", "cumin"}},
		{name: "second page", offset: 2, limit: 2, want: []string{"salt"}},
		{name: "offset past the end", offset: 5, limit: 2, want: nil},
		{name: "negative offset clamps to zero", offset: -3, limit: 1, want: []string{"apple"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := seedIngredients(t)

			got, err := tx.ListIngredients(listIngredientsUserID, tt.offset, tt.limit)
			if err != nil {
				t.Fatal(err)
			}
			assertIngredientNames(t, got, tt.want)
		})
	}
}

func TestListIngredients_NoneVisible(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tx, err := db.NewTransaction(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	got, err := tx.ListIngredients(listIngredientsUserID, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("got %v, want nil", ingredientNames(got))
	}
}
