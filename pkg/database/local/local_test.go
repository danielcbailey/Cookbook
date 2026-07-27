package local_test

import (
	"testing"

	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/pkg/database/local"
)

func testUser() models.User {
	return models.User{
		Email:                      "test@example.com",
		FirstName:                  "Test",
		LastName:                   "User",
		PasswordHash:               "hash123",
		RecipeLimit:                50,
		MaxPhotoStorageMB:          500,
		MaxMonthlyRecipeExtraction: 20,
	}
}

func testRecipe(userID int64) models.Recipe {
	return models.Recipe{
		UserID:      userID,
		Title:       "Pasta",
		Description: "Simple pasta",
		Servings:    4,
		Category:    "dinner",
		Protein:     "none",
		Tags: []models.RecipeTag{
			{Name: "italian"},
			{Name: "quick"},
		},
	}
}

func TestCommitPersists(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.NewTransaction(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	user := testUser()
	if _, err := tx.CreateUser(&user); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	tx2, err := db.NewTransaction(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer tx2.Rollback()

	got, err := tx2.GetUserByID(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != "test@example.com" {
		t.Fatalf("got %q, want %q", got.Email, "test@example.com")
	}
}

func TestRollbackDiscards(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.NewTransaction(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	user := testUser()
	if _, err := tx.CreateUser(&user); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}

	tx2, err := db.NewTransaction(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	defer tx2.Rollback()

	_, err = tx2.GetUserByID(user.ID)
	if err == nil {
		t.Fatal("expected not-found after rollback")
	}
}

func TestDoubleCommitErrors(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.NewTransaction(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err == nil {
		t.Fatal("expected error on double commit")
	}
}

func TestDoubleRollbackErrors(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.NewTransaction(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatal(err)
	}
	if err := tx.Rollback(); err == nil {
		t.Fatal("expected error on double rollback")
	}
}
