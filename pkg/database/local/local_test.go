package local_test

import (
	"errors"
	"testing"

	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/pkg/database"
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

// TestDeferredRollbackAfterCommit covers the `defer tx.Rollback()` pattern the
// callers rely on: the deferred call must not release the transaction's
// resource a second time, or the next transaction would never be able to start.
func TestDeferredRollbackAfterCommit(t *testing.T) {
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
	if err := tx.Rollback(); !errors.Is(err, database.ErrTxClosed) {
		t.Fatalf("expected ErrTxClosed on rollback after commit, got %v", err)
	}

	tx2, err := db.NewTransaction(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if err := tx2.Commit(); err != nil {
		t.Fatal(err)
	}
}
