package local_test

import (
	"errors"
	"testing"

	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/pkg/database"
	"github.com/danielcbailey/Cookbook/pkg/database/local"
)

func TestCreateUser_AssignsID(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tx, _ := db.NewTransaction(t.Context())
	defer tx.Rollback()

	u := testUser()
	if err := tx.CreateUser(&u); err != nil {
		t.Fatal(err)
	}
	if u.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}

func TestGetUserByID(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tx, _ := db.NewTransaction(t.Context())
	defer tx.Rollback()

	u := testUser()
	tx.CreateUser(&u)

	got, err := tx.GetUserByID(u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != u.Email {
		t.Fatalf("got %q, want %q", got.Email, u.Email)
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tx, _ := db.NewTransaction(t.Context())
	defer tx.Rollback()

	_, err = tx.GetUserByID(999)
	if !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetUserByPasswordHash(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tx, _ := db.NewTransaction(t.Context())
	defer tx.Rollback()

	u := testUser()
	tx.CreateUser(&u)

	got, err := tx.GetUserByPasswordHash("test@example.com", "hash123")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != u.ID {
		t.Fatalf("got ID %d, want %d", got.ID, u.ID)
	}
}

func TestGetUserByPasswordHash_NotFound(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tx, _ := db.NewTransaction(t.Context())
	defer tx.Rollback()

	_, err = tx.GetUserByPasswordHash("test@example.com", "nonexistent")
	if !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tx, _ := db.NewTransaction(t.Context())
	defer tx.Rollback()

	u := testUser()
	tx.CreateUser(&u)

	u.FirstName = "Updated"
	if err := tx.UpdateUser(&u); err != nil {
		t.Fatal(err)
	}

	got, _ := tx.GetUserByID(u.ID)
	if got.FirstName != "Updated" {
		t.Fatalf("got %q, want %q", got.FirstName, "Updated")
	}
}

func TestUpdateUser_NotFound(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tx, _ := db.NewTransaction(t.Context())
	defer tx.Rollback()

	u := models.User{ID: 999}
	if err := tx.UpdateUser(&u); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteUser_CascadesToRecipes(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tx, _ := db.NewTransaction(t.Context())
	defer tx.Rollback()

	u := testUser()
	tx.CreateUser(&u)

	r := testRecipe(u.ID)
	tx.CreateRecipe(&r)

	if err := tx.DeleteUser(u.ID); err != nil {
		t.Fatal(err)
	}

	_, err = tx.GetUserByID(u.ID)
	if !errors.Is(err, database.ErrNotFound) {
		t.Fatal("user should be deleted")
	}
	_, err = tx.GetRecipeByID(r.ID)
	if !errors.Is(err, database.ErrNotFound) {
		t.Fatal("recipe should be cascade-deleted with user")
	}
}

func TestUpdateUserUsage(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tx, _ := db.NewTransaction(t.Context())
	defer tx.Rollback()

	u := testUser()
	tx.CreateUser(&u)

	if err := tx.UpdateUserUsage(u.ID, 10, 3); err != nil {
		t.Fatal(err)
	}
	if err := tx.UpdateUserUsage(u.ID, 5, 1); err != nil {
		t.Fatal(err)
	}

	got, _ := tx.GetUserByID(u.ID)
	if got.CurrentPhotoStorageMB != 15 {
		t.Fatalf("photo storage: got %d, want 15", got.CurrentPhotoStorageMB)
	}
	if got.CurrentMonthlyRecipeExtraction != 4 {
		t.Fatalf("recipe extraction: got %d, want 4", got.CurrentMonthlyRecipeExtraction)
	}
}

func TestUpdateUserUsage_NotFound(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tx, _ := db.NewTransaction(t.Context())
	defer tx.Rollback()

	if err := tx.UpdateUserUsage(999, 1, 1); !errors.Is(err, database.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestGetUserByID_ReturnsCopy(t *testing.T) {
	db, err := local.NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	tx, _ := db.NewTransaction(t.Context())
	defer tx.Rollback()

	u := testUser()
	tx.CreateUser(&u)

	got, _ := tx.GetUserByID(u.ID)
	got.FirstName = "mutated"

	got2, _ := tx.GetUserByID(u.ID)
	if got2.FirstName == "mutated" {
		t.Fatal("modifying returned user should not affect store")
	}
}
